package main

import (
	"context"
	"flag"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "/home/rishabh/.kube/config", "location to your kubeconfig file")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	cm, err := clientset.CoreV1().ConfigMaps("test").List(context.Background(), metav1.ListOptions{})
	for _, c := range cm.Items {
		fmt.Printf("Configmap: %s\n", c.Name)

	}
	deployments, err := clientset.AppsV1().Deployments("test").List(context.Background(), metav1.ListOptions{})
	for _, deployment := range deployments.Items {
		fmt.Printf("Deployment %s\n", deployment.Name)
	}
}
