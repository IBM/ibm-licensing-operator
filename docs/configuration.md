# Configuration

After you install License Service, you can configure License Service if needed.

## Configuring ingress

You might want to configure ingress. Here is an <b>example</b> of how you can do it:

1\. Get the nginx ingress controller. You might get it, for example, from here: [https://kubernetes.github.io/ingress-nginx/deploy](https://kubernetes.github.io/ingress-nginx/deploy)

2\. Apply this IBMLicensing instance to your cluster:

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: operator.ibm.com/v1alpha1
kind: IBMLicensing
metadata:
  name: instance
spec:
  apiSecretToken: ibm-licensing-token
  datasource: datacollector
  httpsEnable: false
  instanceNamespace: ibm-common-services
  ingressEnabled: true
  ingressOptions:
    annotations:
      "nginx.ingress.kubernetes.io/rewrite-target": "/\$2"
    path: /ibm-licensing-service-instance(/|$)(.*)
EOF
```

3\. Access the instance at your ingress host with the following path: `/ibm-licensing-service-instance`.

**Other Examples:**

- ICP cluster

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: operator.ibm.com/v1alpha1
kind: IBMLicensing
metadata:
  name: instance
spec:
  apiSecretToken: ibm-licensing-token
  datasource: datacollector
  httpsEnable: false
  instanceNamespace: ibm-common-services
  ingressEnabled: true
  ingressOptions:
    annotations:
      "icp.management.ibm.com/rewrite-target": "/"
      "kubernetes.io/ingress.class": "ibm-icp-management"
EOF
```

- IBM Cloud with bluemix ingress

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: operator.ibm.com/v1alpha1
kind: IBMLicensing
metadata:
  name: instance
spec:
  apiSecretToken: ibm-licensing-token
  datasource: datacollector
  httpsEnable: false
  instanceNamespace: ibm-common-services
  ingressEnabled: true
  ingressOptions:
    annotations:
      ingress.bluemix.net/rewrite-path: "serviceName=ibm-licensing-service-instance rewrite=/"
    path: /ibm-licensing-service-instance
    host: <your_host> # maybe this value can be skipped, you need to check
EOF
```

**Note:** For HTTPS, set `spec.httpsEnable` to `true`, and edit `ingressOptions`. Read more about the options here:
[IBMLicensingOperatorParameters](images/IBMLicensingOperatorParameters.csv)

**Troubleshooting**: If the instance is not updated properly (for example after updating ingressOptions), try deleting the instance and creating new one with new parameters.

## Check Components

After you apply appropriate configuration for **IBM License Service** follow these steps to check whether it works:

1\. Check if the pod is running, by running the following commands:

```bash
podName=`kubectl get pod -n ibm-common-services -o jsonpath="{range .items[*]}{.metadata.name}{'\n'}" | grep ibm-licensing-service-instance`
kubectl logs $podName -n ibm-common-services
kubectl describe pod $podName -n ibm-common-services
```

2\. Check Route or Ingress settings depending on your parameter settings, for example, using these commands.

```bash
kubectl get ingress -n ibm-common-services -o yaml
```

Then examine the status part of the output. It should include host, path, tls (if configured), and other networking information.

3\. At this point **License Service** should be running in your cluster. For more information about how to use License Service to retrieve license usage data, see [IBM Cloud Platform Common Services documentation](https://www.ibm.com/support/knowledgecenter/SSHKN6/license-service/1.x.x/retrieving.html).

## Using custom certificates

You can use either a self-signed certificate or a custom certificate when you use License Service API over https. In order to set up a custom certificate complete the following steps:

1\. Make sure that the IBM Licensing operator is installed.

2\. Create a Kubernetes TLS secret in the namespace where License Service is deployed.

   a. Change the certificate name to 'tls.crt'.

   b. Change the key name to 'tls.key'.

   c. In the terminal, change the directory to where the key and the certificate are stored.

```bash
cd <directory with the certificate and the key>
```

   d. Create the secret with the following command:

```bash
licensingNamespace=ibm-common-services
kubectl create secret tls ibm-licensing-certs --key tls.key --cert tls.crt -n ${licensingNamespace}
```

3\. Edit a new IBM Licensing instance, or edit the existing one to include the certificate:

   ```yaml
   apiVersion: operator.ibm.com/v1alpha1
   kind: IBMLicensing
   metadata:
      name: instance
   # ...
   spec:
      httpsEnable: true # <- this enables https
      httpsCertsSource: custom # <- this makes License Service API use ibm-licensing-certs secret
   # ... append rest of the License Service configuration here
   ```