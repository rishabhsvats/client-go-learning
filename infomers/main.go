package main

import (
	"flag"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	kubeconfig := flag.String("kubeconfig", "/home/rissingh/.kube/config", "location to your kubeconfig file")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("error while creating clientset %s", err.Error())
	}

	informerfactory := informers.NewSharedInformerFactory(clientset, 30*time.Second)
	podinformer := informerfactory.Core().V1().Pods().Informer()
	// podinformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
	// 	AddFunc: func(new interface{}) {
	// 		fmt.Println("Add was called")
	// 	},
	// 	UpdateFunc: func(old, new interface{}) {
	// 		fmt.Println("update was called")
	// 	},
	// 	DeleteFunc: func(obj interface{}) {
	// 		fmt.Println("delete was called")
	// 	},
	// })
	informerfactory.Start(wait.NeverStop)
	informerfactory.WaitForCacheSync(wait.NeverStop)
	podItem, _, _ := podinformer.GetIndexer().GetByKey("openshift-gitops" + "/" + "openshift-gitops-application-controller-0")

	pod := podItem.(*corev1.Pod)
	fmt.Println("Startime is : ", pod.Status.StartTime)

}
