# Manual installation without the Operator Lifecycle Manager (OLM)

Learn how to install License Service without the Operator Lifecycle Manager (OLM).

Complete the following procedure to install License Service on a system that does not have the Operator Lifecycle Manager (OLM) deployed.

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Creating an IBM Licensing instance](#creating-an-ibm-licensing-instance)

## Prerequisites

- Complete the installation on a host that meets the following criteria:
    - Has Linux or macOS operating system (or Windows with Linux Bash Shell for example from WSL).
    - Has Docker and Kubernetes CLI installed.
    - Has internet access.
    - Has access to your cluster via Kubernetes config.

## Installation

This procedure guides you through the installation of License Service. It does not cover the installation of License Service Reporter which is not available without an IBM Cloud Pak on OpenShift Container Platform.

Complete the following steps to create the required resources.

1\. Run the following command to create the `ibm-common-services` namespace where you will later install the operator.

```bash
export licensing_namespace=ibm-common-services
kubectl create namespace ${licensing_namespace}
```

2\. Use the following command to set the context so that the resources are created.

```bash
current_context=$(kubectl config current-context)
kubectl config set-context ${current_context} --namespace=${licensing_namespace}
```

or when you are using OpenShift just:

```bash
oc project ${licensing_namespace}
```

3\. Use `git clone`.

```bash
export operator_release_version=v1.8.0
git clone -b ${operator_release_version} https://github.com/IBM/ibm-licensing-operator.git
cd ibm-licensing-operator/
```

4\. Switch namespaces in rbac if different namespace than `ibm-common-services`:

- For **LINUX** users:

```bash
if [ "${licensing_namespace}" != "" ] && [ "${licensing_namespace}" != "ibm-common-services" ]; then
  sed -i 's|ibm-common-services|'"${licensing_namespace}"'|g' config/rbac/*.yaml
fi
```

- For **MAC** users:

```bash
if [ "${licensing_namespace}" != "" ] && [ "${licensing_namespace}" != "ibm-common-services" ]; then
  sed -i "" 's|ibm-common-services|'"${licensing_namespace}"'|g' config/rbac/*.yaml
fi
```

5\. Apply RBAC roles and CRD:

```bash
# add CRD:
kubectl apply -f config/crd/bases/operator.ibm.com_ibmlicensings.yaml
kubectl apply -f config/crd/bases/operator.ibm.com_ibmlicenseservicereporters.yaml
kubectl apply -f config/crd/bases/operator.ibm.com_ibmlicensingmetadatas.yaml
# add RBAC:
kubectl apply -f config/rbac/role.yaml
kubectl apply -f config/rbac/role_operands.yaml
kubectl apply -f config/rbac/service_account.yaml
kubectl apply -f config/rbac/role_binding.yaml
```

6\. Modify the `operator.yaml` image based on tags.

- For **LINUX** users:

```bash
sed -i "s/annotations\['olm.targetNamespaces'\]/namespace/g" config/manager/manager.yaml
kubectl apply -f config/manager/manager.yaml
```

- For **MAC** users:

```bash
sed -i "" "s/annotations\['olm.targetNamespaces'\]/namespace/g" config/manager/manager.yaml
kubectl apply -f config/manager/manager.yaml
```

**Results:**
You have created the **Operator** for **IBM Licensing Service**. The **Operator** is only responsible for watching over the configuration and managing resources used by **IBM Licensing Service**.

**What's Next:**
Create an IBM Licensing instance.

## Creating an IBM Licensing instance

To create the IBM Licensing instance, run the following command.

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: operator.ibm.com/v1alpha1
kind: IBMLicensing
metadata:
  name: instance
spec:
  datasource: datacollector
  httpsCertsSource: self-signed
  httpsEnable: true
EOF
```

**Results:**
Give operator couple minutes to configure all needed components.
Installation is complete and **License Service** is running in your cluster.

<b>Related links</b>

- [Go back to home page](../License_Service_main.md#documentation)
- [Configuration](Configuration.md)
- [Retrieving license usage data from the cluster](Retrieving_data.md)
