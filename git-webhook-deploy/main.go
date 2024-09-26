package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/go-github/v65/github"
	"golang.org/x/term"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {

	var (
		err error
	)
	s := server{
		webhookSecretKey: os.Getenv("WEBHOOK_SECRET"),
	}
	if s.client, err = getClient(false); err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}
	if s.githubClient, err = getGithubClient(); err != nil {
		fmt.Println("Error: %s", err)
		os.Exit(1)
	}
	http.HandleFunc("/webhook", s.webhook)
	http.ListenAndServe(":8080", nil)
}

func getClient(inCluster bool) (*kubernetes.Clientset, error) {
	var (
		config *rest.Config
		err    error
	)
	if inCluster {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		kubeConfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			return nil, err
		}
	}

	// use the current context in kubeconfig

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}
func getGithubClient() (*github.Client, error) {
	fmt.Print("GitHub Token: ")
	byteToken, _ := term.ReadPassword(int(os.Stdin.Fd()))
	println()
	token := string(byteToken)
	return github.NewClient(nil).WithAuthToken(token), nil
}

func deploy(client *kubernetes.Clientset, ctx context.Context, appFile []byte) (map[string]string, int32, error) {
	var deployment *v1.Deployment

	obj, groupVersionKind, err := scheme.Codecs.UniversalDeserializer().Decode(appFile, nil, nil)

	switch obj := obj.(type) {
	case *v1.Deployment:
		deployment = obj
	default:
		return nil, 0, fmt.Errorf("unrecognized type: %s\n", groupVersionKind)
	}
	_, err = client.AppsV1().Deployments("nginx").Get(ctx, deployment.Name, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		deploymentResponse, err := client.AppsV1().Deployments("nginx").Create(ctx, deployment, metav1.CreateOptions{})
		if err != nil {
			return nil, 0, fmt.Errorf("deployment error: %s", err)
		}
		return deploymentResponse.Spec.Template.Labels, *deploymentResponse.Spec.Replicas, nil
	} else if err != nil && !errors.IsNotFound(err) {
		return nil, 0, fmt.Errorf("deployment get error: %s", err)
	}
	deploymentResponse, err := client.AppsV1().Deployments("nginx").Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("deployment error: %s", err)
	}
	return deploymentResponse.Spec.Template.Labels, *deploymentResponse.Spec.Replicas, nil
}

func waitForPods(client *kubernetes.Clientset, ctx context.Context, deploymentLabels map[string]string, expectedPods int32) error {
	//fmt.Printf("expected pod %d", int(expectedPods))
	for {
		validatedLabels, _ := labels.ValidatedSelectorFromSet(deploymentLabels)
		podList, err := client.CoreV1().Pods("nginx").List(ctx, metav1.ListOptions{
			LabelSelector: validatedLabels.String(),
		})
		if err != nil {
			return fmt.Errorf("pod list error: %s", err)
		}
		podsRunning := 0
		for _, pod := range podList.Items {
			if pod.Status.Phase == "Running" {
				podsRunning++
			}
		}
		fmt.Printf("Waiting for pods to become ready (runninng %d / %d) expected pods : %d\n", podsRunning, len(podList.Items), expectedPods)
		if podsRunning > 0 && podsRunning == len(podList.Items) && podsRunning == int(expectedPods) {
			break
		}

		time.Sleep(5 * time.Second)
	}

	return nil
}
