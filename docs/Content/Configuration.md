# Configuration

After you install License Service, you can configure License Service if needed.

- [Configuring ingress](#configuring-ingress)
- [Checking License Service components](#checking-license-service-components)
- [Using custom certificates](#using-custom-certificates)
- [Cleaning existing License Service dependencies](#cleaning-existing-license-service-dependencies)
    - [Cleaning existing License Service dependencies outside of OpenShift](#cleaning-existing-license-service-dependencies-outside-of-openshift)
    - [Cleaning existing License Service dependencies on OpenShift Container Platform](#cleaning-existing-license-service-dependencies-on-openshift-container-platform)
- [Modifying the application deployment resources](#modifying-the-application-deployment-resources)

## Configuring ingress

You can configure ingress, for example, by completing the following steps:

1\. Get the nginx ingress controller. You might get it, for example, from [https://kubernetes.github.io/ingress-nginx/deploy](https://kubernetes.github.io/ingress-nginx/deploy).

**Note:** If you already have ingress controller on the cluster, you do not need to get another one.

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
  ingressEnabled: true
  ingressOptions:
    annotations:
      "nginx.ingress.kubernetes.io/rewrite-target": "/\$2" # <- if you copy it into yaml file, then use "/$2"
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
  ingressEnabled: true
  ingressOptions:
    annotations:
      "icp.management.ibm.com/rewrite-target": "/"
      "kubernetes.io/ingress.class": "ibm-icp-management"
EOF
```

- IBM Cloud Kubernetes Services (IKS) on IBM Cloud

First get your cluster name:

```yaml
cluster=<your iks cluster name from ibmcloud>
```

Then, get the subdomain for your cluster (or just fill subdomain variable if you know your subdomain):

```yaml
subdomain=$(ibmcloud ks cluster get --cluster <cluster name or cluster id> | grep "Ingress Sub" | awk '{print $3}')
```

Then apply the instance:

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
  ingressEnabled: true
  ingressOptions:
    annotations:
      "nginx.ingress.kubernetes.io/rewrite-target": "/\$2" # <- if you copy it into yaml file, then use "/$2"
      "kubernetes.io/ingress.class": "public-iks-k8s-nginx"
    host: $subdomain
    path: /ibm-licensing-service-instance(/|$)(.*)
EOF
```

- Amazon Elastic Kubernetes Service (EKS)

To retrieve your host, consult EKS documentation.

Then, get the subdomain for your cluster (or just fill subdomain variable if you know your subdomain):

```yaml
subdomain=$(kubectl get svc ingress-nginx-controller -n ingress-nginx -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
```

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
  ingressEnabled: true
  ingressOptions:
    annotations:
      "kubernetes.io/ingress.class": nginx
      "nginx.ingress.kubernetes.io/rewrite-target": "/\$2" # <- if you copy it into yaml file, then use "/$2"
    path: /ibm-licensing-service-instance(/|$)(.*)
    host: $subdomain
EOF
```

For more information, see [Setting up Kubernetes Ingress](https://cloud.ibm.com/docs/containers?topic=containers-ingress-types) in IBM Cloud Docs.

**Note:** For HTTPS, set `spec.httpsEnable` to `true`, and edit `ingressOptions`. Read more about the options here:
[IBMLicensingOperatorParameters](../../images/IBMLicensingOperatorParameters.csv)

**Troubleshooting**: If the instance is not updated properly (for example after updating ingressOptions), try deleting the instance and creating new one with new parameters.

## Checking License Service components

After you install **IBM License Service**, complete the following steps to check whether License Service works properly:

1\. To check if the pod is running, by running the following commands:

```bash
podName=`kubectl get pod -n ibm-common-services -o jsonpath="{range .items[*]}{.metadata.name}{'\n'}" | grep ibm-licensing-service-instance`
kubectl logs $podName -n ibm-common-services
kubectl describe pod $podName -n ibm-common-services
```

2\. Check Route or Ingress settings depending on your parameter settings, for example, using these commands.

- To check Ingres, run the following command:

```bash
kubectl get ingress -n ibm-common-services -o yaml
```

- To check Route, run the following command:

```bash
kubectl get route -n ibm-common-services -o yaml
```

Then examine the status part of the output. It should include host, path, tls (if configured), and other networking information.

3\. Run the License Service APIs, and make sure that you get results that reflect your environment's license usage. For more information, see [APIs for retrieving License Service data in IBM Documentation](https://www.ibm.com/docs/en/cpfs?topic=data-per-cluster-from-license-service).

## Using custom certificates

You can use either a self-signed certificate or a custom certificate when you use License Service API over https. In order to set up a custom certificate complete the following steps:

1\. Change the certificate name to `tls.crt`.

2\. Change the key name to `tls.key`.

3\. Run the following command to change the directory to where the certificate and the key are stored:

```bash
cd <certificate_directory>
```

4\. Create a secret by using the following command:

```bash
licensingNamespace=$(oc get pods --all-namespaces | grep "ibm-licensing-service-" | awk {'print $1'})
kubectl create secret tls ibm-licensing-certs --key tls.key --cert tls.crt -n ${licensingNamespace}
```

5\. Open the IBMLicensing instance YAML to include the certificate by running the following command:

```bash
kubectl edit IBMLicensing instance
```

6\. Edit the YAML and add the following parameters to the `IBMLicensing` section, under `spec`:

- To enable the https connection, add the following line:

`httpsEnable: true`

- To apply the custom certificate that you created as `ibm-licensing-certs`, add the following line:

`httpsCertsSource: custom`

For example:

```yaml
apiVersion: operator.ibm.com/v1alpha1
kind: IBMLicensing
metadata:
    name: instance
  spec:
    httpsEnable: true
    httpsCertsSource: custom
```

7\. Save the changes in YAML.

## Cleaning existing License Service dependencies

Earlier versions of License Service, up to 1.1.3, used OperatorSource and Operator Marketplace. These dependencies are no longer needed. If you installed the earlier version of License Service, before installing the new version remove the existing dependencies from your system.

### Cleaning existing License Service dependencies outside of OpenShift

1\. Delete OperatorSource.

To delete an existing OpenSource, run the following command:

```bash
GLOBAL_CATALOG_NAMESPACE=olm
opencloudioSourceName=opencloud-operators
kubectl delete OperatorSource ${opencloudioSourceName} -n ${GLOBAL_CATALOG_NAMESPACE}
```

where `GLOBAL_CATALOG_NAMESPACE` value is your global catalog namespace.

2\. Delete OperatorMarketplace.

**Note:** Before deleting OperatorMarketplace check whether it is not used elsewhere, for example, for other Operators from OperatorMarketplace.

To delete OperatorMarketplace, run the following command:

```bash
GLOBAL_CATALOG_NAMESPACE=olm
kubectl delete Deployment marketplace-operator -n ${GLOBAL_CATALOG_NAMESPACE}
kubectl delete RoleBinding marketplace-operator -n ${GLOBAL_CATALOG_NAMESPACE}
kubectl delete ClusterRoleBinding marketplace-operator
kubectl delete ServiceAccount marketplace-operator -n ${GLOBAL_CATALOG_NAMESPACE}
kubectl delete Role marketplace-operator -n ${GLOBAL_CATALOG_NAMESPACE}
kubectl delete ClusterRole marketplace-operator
```

where `GLOBAL_CATALOG_NAMESPACE` value is your global catalog namespace.

3\. Reinstall License Service to get CatalogSource that is missing. For more information, see [Installing License Service](Content/Installation_scenarios.md).

### Cleaning existing License Service dependencies on OpenShift Container Platform

1\. Delete OperatorSource.

To delete an existing OpenSource, run the following command:

```bash
GLOBAL_CATALOG_NAMESPACE= openshift-marketplaceolm
opencloudioSourceName=opencloud-operators
kubectl delete OperatorSource ${opencloudioSourceName} -n ${GLOBAL_CATALOG_NAMESPACE}
```

2\. Reinstall License Service to get CatalogSource that is missing. For more information, see [Installing License Service](Content/Installation_scenarios.md).

## Modifying the application deployment resources

You can modify the resources that are requested and limited by the Deployment for Application by editing the IBMLicensing instance.

To learn what resources are required for License Service in your environment, see [Preparing for installation: Required resources](Preparing_for_installation.md#required-resources).

1\. To modify the IBMLicensing instance, run the following command:

```bash
kubectl edit IBMLicensing instance
```

2\. Modify the resource limits and resources in the following yaml and paste it in the command line.

```yaml
apiVersion: operator.ibm.com/v1alpha1
kind: IBMLicensing
metadata:
  name: instance
spec:
# ...
  resources:
    limits:
      cpu: 500m <- set the CPU limit to the desired value
      memory: 512Mi <- set the memory limit to the desired value
    requests:
      cpu: 200m <- set the requests limit to the desired value
      memory: 256Mi <- set the memory limit to the desired value
# ...
```

*where m stands for Millicores, and Mi for Mebibytes

<b>Related links</b>

- [Go back to home page](../License_Service_main.md#documentation)
