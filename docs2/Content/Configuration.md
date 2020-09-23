# Configuration

After you install License Service, you can configure License Service if needed.

- [Configuring ingress](#configuring-ingress)
- [Checking License Service components](#checking-license-service-components)
- [Using custom certificates](#using-custom-certificates)
- [Cleaning existing License Service dependencies](#cleaning-existing-license-service-dependencies)
- [Modifying the application deployment resources](#modifying-the-application-deployment-resources)

## Configuring ingress

You might want to configure ingress. Here is an <b>example</b> of how you can do it:

1\. Get the nginx ingress controller. You might get it, for example, from here: [https://kubernetes.github.io/ingress-nginx/deploy](https://kubernetes.github.io/ingress-nginx/deploy).

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
[IBMLicensingOperatorParameters](../../images/IBMLicensingOperatorParameters.csv)

**Troubleshooting**: If the instance is not updated properly (for example after updating ingressOptions), try deleting the instance and creating new one with new parameters.

## Checking License Service components

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

## Cleaning existing License Service dependencies 

Earlier versions of License Service, up to 1.1.3, used OperatorSource and Operator Marketplace. These dependencies are no longer needed. If you installed the earlier version of License Service, before installing the new version remove the existing dependencies from your system. 

### Cleaning existing License Service dependencies outside of OpenShift

1. Delete OperatorSource.

To delete an existing OpenSource, run the following command:

```
GLOBAL_CATALOG_NAMESPACE=olm
opencloudioSourceName=opencloud-operators
kubectl delete OperatorSource ${opencloudioSourceName} -n ${GLOBAL_CATALOG_NAMESPACE}
```

where `GLOBAL_CATALOG_NAMESPACE` value is your global catalog namespace.

2. Delete OperatorMarketplace.

**Note:** Before deleting OperatorMarketplace check whether it is not used elsewhere, for example, for other Operators from OperatorMarketplace, or an OCP cluster.

To delete OperatorMarketplace, run the following command:

```
GLOBAL_CATALOG_NAMESPACE=olm
kubectl delete Deployment marketplace-operator -n ${GLOBAL_CATALOG_NAMESPACE}
kubectl delete RoleBinding marketplace-operator -n ${GLOBAL_CATALOG_NAMESPACE}
kubectl delete ClusterRoleBinding marketplace-operator
kubectl delete ServiceAccount marketplace-operator -n ${GLOBAL_CATALOG_NAMESPACE}
kubectl delete Role marketplace-operator -n ${GLOBAL_CATALOG_NAMESPACE}
kubectl delete ClusterRole marketplace-operator
```

where `GLOBAL_CATALOG_NAMESPACE` value is your global catalog namespace.

### Cleaning existing License Service dependencies on OpenShift Container Platform

1. Delete OperatorSource.

To delete an existing OpenSource, run the following command:

```
GLOBAL_CATALOG_NAMESPACE= openshift-marketplaceolm
opencloudioSourceName=opencloud-operators
kubectl delete OperatorSource ${opencloudioSourceName} -n ${GLOBAL_CATALOG_NAMESPACE}
```

## Modifying the application deployment resources

You can modify the resources that are requested and limited by the Deployment for Application by editing the IBMLicensing instance.

1. To modify the IBMLicensing instance, run the following command:

```bash
kubectl edit IBMLicensing instance
```
2. Modify the resource limits and resources in the following yaml and paste it in the command line. 

```yaml
apiVersion: operator.ibm.com/v1alpha1
kind: IBMLicensing
metadata:
  name: instance
spec:
# ...
  resources:
    limits:
      cpu: 500m
      memory: 512Mi
    requests:
      cpu: 200m
      memory: 256Mi
# ...
```

**Related links**

- [Go back to home page](../License_Service_main.md#documentation)
