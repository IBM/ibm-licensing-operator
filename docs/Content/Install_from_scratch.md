# Manual installation on Kubernetes from scratch with `kubectl`

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Creating an IBM Licensing instance](#creating-an-ibm-licensing-instance)
- [Verification](#verification)

## Prerequisites

- Administrator permissions for the cluster
- `kubectl` 1.19 or higher
- Linux or iOS

Before installation, see [Preparing for installation](Preparing_for_installation.md) to check the supported platforms, required resources, and cluster permissions.

## Installation

This procedure guides you through the installation of License Service. It does not cover the installation of License Service Reporter, which is not available without an IBM Cloud Pak on OpenShift Container Platform.

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

- If you get the following response, OLM is installed.

  ```{: .text .no-copy }
  NAME                                          CREATED AT
  clusterserviceversions.operators.coreos.com   2020-06-04T14:42:13Z
  ```

- If you get the following response, OLM CRD is not installed. Continue with step 1c.

  `Error from server (NotFound): customresourcedefinitions.apiextensions.k8s.io "clusterserviceversions.operators.coreos.com" not found`

c.  If OLM is not installed, download it from [the OLM GitHub repository](https://github.com/operator-framework/operator-lifecycle-manager/releases). Use following script to download and install OLM v0.16.1

**Note:** For versions newer than 0.16.1, the process might differ.

```bash
olm_version=0.17.1
curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/${olm_version}/install.sh | bash -s ${olm_version}
```

   **Troubleshooting:** If you get an error, you might have the old version of Kubernetes. You can try either upgrading your Kubernetes server version or using the older version of OLM.

2\. **Create the CatalogSource**

a. To get GLOBAL_CATALOG_NAMESPACE, check `global catalog namespace` in a yaml in a `packageserver` OLM pod that is somewhere in your cluster. You can, for example, use the following command:

```bash
olm_namespace=$(kubectl get csv --all-namespaces -l olm.version -o jsonpath="{.items[?(@.metadata.name=='packageserver')].metadata.namespace}")
GLOBAL_CATALOG_NAMESPACE=$(kubectl get deployment --namespace="${olm_namespace}" packageserver -o yaml | grep -A 1 -i global-namespace | tail -1 | cut -d "-" -f 2- | sed -e 's/^[ \t]*//')
# check if the namespace is found
echo ${GLOBAL_CATALOG_NAMESPACE}
```

If you get an empty response to the `echo` command, you can get global catalog namespace using the following command.

**Note:** The following method should only be used for getting global catalog namespace if the previous method failed.

```bash
GLOBAL_CATALOG_NAMESPACE=$(kubectl get pod --all-namespaces -l app=olm-operator -o jsonpath="{.items[0].metadata.namespace}")
echo ${GLOBAL_CATALOG_NAMESPACE}
```

b. Create the `CatalogSource` by using the following command:

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
  image: icr.io/cpopen/ibm-operator-catalog
  updateStrategy:
    registryPoll:
      interval: 45m
EOF
```

<b>Check the results</b>
- Run the following command to check if the `CatalogSource` is created in the `$GLOBAL_CATALOG_NAMESPACE` namespace:

```console
kubectl get catalogsource -n $GLOBAL_CATALOG_NAMESPACE
```

The following is the sample output:

```{: .text .no-copy }
NAME                           DISPLAY                        TYPE   PUBLISHER   AGE
opencloud-operators            IBMCS Operators                grpc   IBM         20m
[...]
```

- If everything goes well, you should see similar pod running. Run the following command to check if the pod is running:

```console
kubectl get pod -n $GLOBAL_CATALOG_NAMESPACE
```

The following is the sample output:

```{: .text .no-copy }
NAME                                            READY   STATUS    RESTARTS   AGE
opencloud-operators-66df4d97ff-4rhjj            1/1     Running   0          80s
[...]
```

3\. **Create an OperatorGroup**

An `OperatorGroup` is used to denote which namespaces your Operator should watch.
It must exist in the namespace where your operator is deployed, for example, `ibm-common-services`.

a. Create a namespace for IBM Licensing with the following command.

```bash
kubectl create namespace ibm-common-services
```

b. Check if you have tha operator group in that namespace by running the following command.

```bash
kubectl get OperatorGroup -n ibm-common-services
```

- If you get the following response, the operator group was found, and you can go to step 4. Create a Subscription.

```{: .text .no-copy }
NAME            AGE
operatorgroup   39d
```

- If you get the following response, the operator group was not found, and you need to create it.

```{: .text .no-copy }
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
A subscription is created for the operator and is responsible for upgrades of IBM Licensing when needed.

a. Make sure that the `GLOBAL_CATALOG_NAMESPACE` variable has the global catalog namespace value. The global catalog namespace was retrieved in step 2a.

b. Create the **Subscription** using the following command:

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: ibm-licensing-operator-app
  namespace: ibm-common-services
spec:
  channel: v3
  name: ibm-licensing-operator-app
  source: opencloud-operators
  sourceNamespace: $GLOBAL_CATALOG_NAMESPACE
EOF
```

5\. **Verify Operator health**

a. To check whether the IBM Licensing is deployed by OLM from the `CatalogSource`, run the following command.

```console
kubectl get clusterserviceversion -n ibm-common-services
```

The following is the sample output:

```{: .text .no-copy }
NAME                             DISPLAY                 VERSION   REPLACES                        PHASE
ibm-licensing-operator.v1.16.0   IBM Licensing           1.16.0    ibm-licensing-operator.v1.15.0  Succeeded
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

- If you are installing License Service on IBM Cloud Kubernetes Services (IKS) or Amazon Elastic Kubernetes Service (EKS), as the following step you need to configure ingress. For more information, see [Configuring ingress](Configuration.md#configuring-ingress). After you do, verify the installation. You do not need to configure IBM Licensing instance.
- If you are installing on OCP, create an IBM Licensing instance.

## Creating an IBM Licensing instance

**Important:** The minimal setup requires applying this IBMLicensing instance. However, before applying the instance, get familiar with the entire configuration process.

To create the IBM Licensing instance, run the following command:

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: operator.ibm.com/v1alpha1
kind: IBMLicensing
metadata:
  name: instance
spec:
  apiSecretToken: ibm-licensing-token
  httpsEnable: true
  instanceNamespace: ibm-common-services
  datasource: datacollector
EOF
```

**Results:** Installation is complete and **License Service** is running in your cluster.

## Verification

To check whether License Service components are properly installed and running, see [Checking License Service components](Configuration.md#checking-license-service-components).

<b>Related links</b>

- [Go back to home page](../License_Service_main.md#documentation)
- [Configuration](Configuration.md)
- [Retrieving license usage data from the cluster](Retrieving_data.md)
