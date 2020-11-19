# Validating License Service deployment

After the installation, check whether License Service was successfully deployed on the cluster by using any available tools. For example, you can log in to the cluster and run the following command:

`kubectl get pods --all-namespaces | grep ibm-licensing | grep -v operator`

The following response is a confirmation of successful deployment:

 `1/1     Running`
