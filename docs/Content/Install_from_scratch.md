# Manual installation on Kubernetes from scratch with `kubectl`

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Creating an IBM Licensing instance](#creating-an-ibm-licensing-instance)

## Prerequisites

- Administrator permissions for the cluster
- `kubectl` 1.11.3 or higher
- Linux or iOS

## Installation

   **Note:** To install License Service on Windows, adjust the commands to fit the Windows standard.

1\. **Install the Operator Lifecycle Manager (OLM)**

a. Make sure that you are connected to your cluster. You can run the following command:

```bash
kubectl get node
```

The response should contain a list of your nodes.

b. Check if you have OLM installed. For example, run the following command.

```bash
kubectl get crd clusterserviceversions.operators.coreos.com
```

- If you get the following response, OLM is installed and you can go to step 2: Create the CatalogSource.

```console
NAME                                          CREATED AT
clusterserviceversions.operators.coreos.com   2020-06-04T14:42:13Z
```

- If you get the following response, OLM CRD is not installed.

```console
Error from server (NotFound): customresourcedefinitions.apiextensions.k8s.io "clusterserviceversions.operators.coreos.com" not found
```

c.  If you do not have OLM, download it from [the OLM GitHub repository](https://github.com/operator-framework/operator-lifecycle-manager/releases). Use following script to download and install OLM v13.0

**Note:** For versions newer than 13.0, the process might differ.

```bash
olm_version=0.13.0
curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/${olm_version}/install.sh | bash -s ${olm_version}
```

   **Troubleshooting:** If you get an error, you might have the old version of Kubernetes. You can try either upgrading your Kubernetes server version or using the older version of OLM.

2\. **Create the CatalogSource**

a. Make sure that GLOBAL_CATALOG_NAMESPACE has the global catalog namespace value and create `CatalogSource` to get operator bundles from `quay.io`.

b. In order to get GLOBAL_CATALOG_NAMESPACE, check your `global catalog namespace` at OLM `packageserver` pod yaml somewhere in your cluster. For example, you can use this command:

```bash
olm_namespace=$(kubectl get csv --all-namespaces -l olm.version -o jsonpath="{.items[?(@.metadata.name=='packageserver')].metadata.namespace}")
GLOBAL_CATALOG_NAMESPACE=$(kubectl get deployment --namespace="${olm_namespace}" packageserver -o yaml | grep -A 1 -i global-namespace | tail -1 | cut -d "-" -f 2- | sed -e 's/^[ \t]*//')
# check if the namespace is found
echo ${GLOBAL_CATALOG_NAMESPACE}
```

c. When you have the `global catalog namespace` set, you can create the CatalogSource by using the following command:

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: opencloud-operators
  namespace: $GLOBAL_CATALOG_NAMESPACE
spec:
  displayName: IBMCS Operators
  publisher: IBM
  sourceType: grpc
  image: docker.io/ibmcom/ibm-common-service-catalog:latest
  updateStrategy:
    registryPoll:
      interval: 45m
EOF
```

<b>Check the results</b>
- Check if a `CatalogSource` is created in the `$GLOBAL_CATALOG_NAMESPACE` namespace:

```console
$ kubectl get catalogsource -n $GLOBAL_CATALOG_NAMESPACE
NAME                           DISPLAY                        TYPE   PUBLISHER   AGE
opencloud-operators            IBMCS Operators                grpc   IBM         20m
[...]
```

- If everything goes well, you should see the following pods:

```console
$ kubectl get pod -n $GLOBAL_CATALOG_NAMESPACE
NAME                                            READY   STATUS    RESTARTS   AGE
opencloud-operators-66df4d97ff-4rhjj            1/1     Running   0          80s
upstream-community-operators-7ffb6b674b-7qlvx   1/1     Running   0          80s
[...]
```

3\. **Create an OperatorGroup**

An `OperatorGroup` is used to denote which namespaces your Operator should watch.
It must exist in the namespace where your operator is deployed, for example, `ibm-common-services`.

a. Create a namespace for IBM Licensing Operator with the following command.

```bash
kubectl create namespace ibm-common-services
```

b. Check if you have tha operator group in that namespace by running the following command.

```bash
kubectl get OperatorGroup -n ibm-common-services
```

- If you get the following response, the operator group is found and you can go to step 4. Create a Subscription.

```console
NAME            AGE
operatorgroup   39d
```

- If you get the following result, the operator group is not found and you need to create it.

```console
No resources found.
```

c. Create the operator group.
Use the following command to deploy the `OperatorGroup` resource.

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: operatorgroup
  namespace: ibm-common-services
spec:
  targetNamespaces:
  - ibm-common-services
EOF
```

4\. **Create a Subscription**
A subscription is created for the operator and is responsible for upgrades of IBM Licensing Operator when needed.

a. Make sure GLOBAL_CATALOG_NAMESPACE has global catalog namespace value.

b. Create the **Subscription** using the following command:

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: ibm-licensing-operator-app
  namespace: ibm-common-services
spec:
  channel: stable-v1
  name: ibm-licensing-operator-app
  source: opencloud-operators
  sourceNamespace: $GLOBAL_CATALOG_NAMESPACE
EOF
```

5\. **Verify Operator health**

a. See if the IBM Licensing Operator is deployed by OLM from the `CatalogSource` with the following command.

```console
$ kubectl get clusterserviceversion -n ibm-common-services
NAME                            DISPLAY                  VERSION   REPLACES                        PHASE
ibm-licensing-operator.v1.2.2   IBM Licensing Operator   1.2.2     ibm-licensing-operator.v1.2.1   Succeeded
```

**Note:** The above command assumes that you have created the Subscription in the `ibm-common-services` namespace.
If your Operator deployment (CSV) shows `Succeeded` in the `InstallPhase` status, your Operator is deployed successfully. Otherwise, check the `ClusterServiceVersion` objects status for details.

b. **Optional**: Check if the operator is deployed. Run the following command:

```bash
kubectl get deployment -n ibm-common-services | grep ibm-licensing-operator
```

**Results:**
You have created the **Operator** for **IBM Licensing Service**. The **Operator** is only responsible for watching over the configuration and managing resources used by **IBM Licensing Service**.

**What's Next:**
Configure the IBM Licensing instance.

## Creating an IBM Licensing instance

**Important:** The minimal setup requires applying this IBMLicensing instance. However, before applying the instance, get familiar with the entire configuration process.

To create the the IBM Licensing instance, run the following command:

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: operator.ibm.com/v1alpha1
kind: IBMLicensing
metadata:
  name: instance
spec:
  apiSecretToken: ibm-licensing-token
  datasource: datacollector
  httpsEnable: true
  instanceNamespace: ibm-common-services
EOF
```

**Results:** Installation is complete and **License Service** is running in your cluster. To check if License Service components are properly installed, and perform extra configuration, see [Configuration](Configuration.md).

<b>Related links</b>

- [Go back to home page](../License_Service_main.md#documentation)
- [Configuration](Configuration.md)
- [Retrieving license usage data from the cluster](Retrieving_data.md)
