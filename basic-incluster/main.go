package main

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {

	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Printf("error %s getting inclusterconfig", err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("error while creating clientset %s", err.Error())
	}
	cm, err := clientset.CoreV1().ConfigMaps("test").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("error while listing configmap %s", err.Error())
	}
	for _, c := range cm.Items {
		fmt.Printf("Configmap: %s\n", c.Name)

	}

	deployments, err := clientset.AppsV1().Deployments("test").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("error while listing deployment %s", err.Error())
	}
	for _, deployment := range deployments.Items {
		fmt.Printf("Deployment %s\n", deployment.Name)
	}
}
