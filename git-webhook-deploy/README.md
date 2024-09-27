#Git Webhook Deployment 

1. Create a Github repository where the Deployment manifest will be placed.
2. Create a Webhook in the above created repository. This webhook has to be passed as an environment variable with key `WEBHOOK_SECRET`.
3. Start this go application which will be listening on /webhook endpoint. Once the manifests are placed in above created repository, the manifest is captured from github push event by this application and deployed on current logged in Kubernetes cluster. 