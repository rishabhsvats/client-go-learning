package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {

	var (
		client           *kubernetes.Clientset
		deploymentLabels map[string]string
		err              error
		expectedPods     int32
	)
	ctx := context.Background()
	if client, err = getClient(); err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}
	if deploymentLabels, expectedPods, err = deploy(client, ctx); err != nil {
		fmt.Printf("Error while deploying: %s", err)
		os.Exit(1)
	}
	//fmt.Printf("expected pod %d", int(expectedPods))
	if err = waitForPods(client, ctx, deploymentLabels, expectedPods); err != nil {
		fmt.Printf("Error while pod startup: %s", err)
		os.Exit(1)
	}
	fmt.Printf("Deployment finished, did a deploy with labels:  %+v\n", deploymentLabels)
}

func getClient() (*kubernetes.Clientset, error) {

	var kubeconfig = flag.String("kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), "(optional) absolute path to the kubeconfig file")

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func deploy(client *kubernetes.Clientset, ctx context.Context) (map[string]string, int32, error) {
	var deployment *v1.Deployment
	appFile, err := ioutil.ReadFile("app.yaml")
	if err != nil {
		return nil, 0, fmt.Errorf("error while reading the app.yaml file : %s", err)
	}
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
