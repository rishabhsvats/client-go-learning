package main

import (
	"net/http"

	"k8s.io/client-go/kubernetes"
)

type server struct {
	client *kubernetes.Clientset
}

func (s server) webhook(req http.ResponseWriter, w *http.Request) {

}
