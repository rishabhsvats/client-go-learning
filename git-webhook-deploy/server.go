package main

import (
	"fmt"
	"net/http"

	"github.com/google/go-github/v65/github"
	"k8s.io/client-go/kubernetes"
)

type server struct {
	client       *kubernetes.Clientset
	githubClient *github.Client
	webhookSecretKey string
}

func (s server) webhook(w http.ResponseWriter, req *http.Request) {
	payload, err := github.ValidatePayload(req, s.webhookSecretKey)
	if err != nil { 
		w.Header().Write(500)
		fmt.Printf("ValidatePayload error : %s". err)
		return
	 }
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil { 
		w.Header().Write(500)
		fmt.Printf("ValidatePayload error : %s". err)
		return
	 }
	switch event := event.(type) {
	case *github.CommitCommentEvent:
		processCommitCommentEvent(event)
	case *github.CreateEvent:
		processCreateEvent(event)
	...
	}
}
