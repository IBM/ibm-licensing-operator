# Manual installation on OpenShift Container Platform (OCP) version 4.6 or higher

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Creating an IBM Licensing instance](#creating-an-ibm-licensing-instance)

## Prerequisites

- A cluster with OCP version 4.6 or higher
- Administrator permissions for the OCP cluster
- Access to the OpenShift Console

## Installation

This procedure guides you through the installation of License Service. It does not cover the installation of License Service Reporter which is not available without an IBM Cloud Pak on OpenShift Container Platform.

1\. **Create the CatalogSource**

Create the CatalogSource to get the operator bundles that are available at the public website: `quay.io`. The CatalogSource allows your cluster to establish connection to `quay.io`.

a. Log in to the OpenShift console.

b. Click the plus button on the right of the header.

c. Copy the following CatalogSource into the editor.

```yaml
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: opencloud-operators
  namespace: openshift-marketplace
spec:
  displayName: IBMCS Operators
  publisher: IBM
  sourceType: grpc
  image: icr.io/cpopen/ibm-operator-catalog
  updateStrategy:
   registryPoll:
     interval: 45m
```

d. Click **Create**.

2\. **Create the `ibm-common-services` namespace**

a. From the hamburger menu in the OpenShift console, go to **Operators>Operator Hub**.

b. Expand the list of Projects.

c. Select **Create Project**.

![Create Project](/images/create-project.png)

d. Enter **ibm-common-services** as a name and click **Create**.

![Create Project](/images/create-project-2.png)

3\. **Install IBM Licensing Operator package in OperatorHub**

a. Go to **OperatorHub** and search for **IBM Licensing Operator**.

   **Note:** It might take a few minutes for the operator to show up. If, after a while, the operator will not show up, there might be an issue with the CatalogSource.

b. Select **IBM Licensing Operator** and click **Install**.

![Operator Hub IBM Licensing](/images/operator-hub-licensing.png)

c. Set **Update Channel** to the `v3` version.

d. As an **Installation Mode** select **A specific namespace on the cluster**, and set it to `ibm-common-services` namespace that you created in the previous step.

e. Set **Approval Strategy** to **Automatic**.

f. Click **Install**.

4\. **Verify that the installation is successful**

To check whether the installation is successful, wait for about 1 minute, and go to **Installed operators**. You should see IBM Licensing Operator with the **Succeeded** status.

![IBM Licensing Installed](/images/installed.png)

**Results:**
You have created the **Operator** for **IBM Licensing Service**. The **Operator** is only responsible for watching over the configuration and managing resources used by **IBM Licensing Service**.

**What's next:**
Create an IBM Licensing instance.

## Creating an IBM Licensing instance

1\. Log in to the OpenShift console and go to **Installed Operators > IBM Licensing Operator > IBM Licensing tab > Create IBMLicensing**.

![OCP click Create IBM Licensing](/images/ocp_create_instance.png)

2\. Click **Create IBMLicensing** and edit the available parameters if needed.  For more information about the parameters, see the [Custom Resource Definition](/bundle/manifests/operator.ibm.com_ibmlicensings.yaml).

3\. Click **Create**.

   **Note:** To edit your instance in the future, from the OpenShift console, go to **Administration > Custom Resource Definitions >** select **IBMLicensing>instances>Edit IBMLicensing**.

![OCP Edit Instance](/images/ocp_edit_instance.png)

**Troubleshooting**: If the instance is not updated properly, try deleting the instance and creating a new one with new parameters.

4\. Check whether the pod is created and has `Running` status. Give it a few minutes if its not `Running` yet.
To see the logs, go to **OCP UI->Workloads->Pods** and search for **licensing** in the `ibm-common-services` project:

![OCP Pod](/images/ocp_pod.png)

5\. To investigate further, click the name of the pod starting with `ibm-licensing-service-instance` and check its logs and events.

**Results:**
Installation is complete and **License Service** is running in your cluster. To check if License Service components are properly installed, and perform extra configuration, see [Configuration](Configuration.md).

<b>Related links</b>

- [Go back to home page](../License_Service_main.md#documentation)
- [Configuration](Configuration.md)
- [Retrieving license usage data from the cluster](Retrieving_data.md)
