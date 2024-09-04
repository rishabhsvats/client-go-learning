# Incluster client-go example

Note: This example works in `test` namespace. 

1. Build go application
~~~
go build
~~~

2. Build podman image
~~~
podman build -t basic-incluster:4.0 .
~~~

3. Tag Image
~~~
podman tag basic-incluster:4.0 quay.io/rishabh/basic-incluster:4.0
~~~

4. Push Image
~~~
podman push quay.io/rishabh/basic-incluster:4.0
~~~

5. Create the required role and rolebinding to list Configmap and Deployment
~~~
oc create -f role.yaml
oc create -f rolebinding.yaml
~~~

6. Deploy the deployment in Kubernetes cluster
~~~
oc create -f  deployment.yaml
~~~