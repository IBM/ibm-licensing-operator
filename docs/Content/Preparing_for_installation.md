
# Preparing for installation

- [Supported platforms](#supported-platforms)
- [Required resources](#required-resources)
- [Hyperthreading](#hyperthreading)
- [Cluster permissions](#cluster-permissions)

## Supported platforms

<b>Linux x86_64</b>

License Service is supported on all Kubernetes-orchestrated clouds on Linux x86_64.

- Red Hat OpenShift Container Platform 4.6 and later
- Kubernetes 1.19 and later
- Operator Lifecycle Manager (OLM) 0.16.1 or later

<b>Linux on Power (ppc64le), Linux on IBM Z and LinuxONE</b>

License Service is supported on Linux on Power (ppc64le), Linux on IBM Z and LinuxONE in the following scenarios:

|Installation|Deployment scenario|
|---|---|
|<ul><li>With Operator Lifecycle Manager (OLM)</li></ul>|<ul><li>[Automatic installation using Operator Lifecycle Manager (OLM)](Automatic_installation.md)</li><li>[Manual installation on Kubernetes from scratch with `kubectl`](Install_from_scratch.md)</li><li>[Offline installation](Install_offline.md)</li></ul>|
|<ul><li>Without Operator Lifecycle Manager (OLM)</li></ul>| <ul><li>[Offline installation](Install_offline.md)</li><li>[Manual installation without the Operator Lifecycle Manager (OLM)](Install_without_OLM.md)</li></ul>|

## Required resources

By default, License Service is installed with the resource settings for medium environments with up to 500 pods and three Cloud Paks. License Service consists of two main components that require resources: the operator deployment and the application deployment. The following table shows the required resources for these components for the medium environment:

|CPU Request (m)| CPU Limit (m)|Memory Request (Mi)|Memory Limit (Mi)|
|---|---|---|---|---|
|Linux® x86_64| 200 | 300| 430| 850|
|Linux® on Power® (ppc64le)|300| 400| 230| 543|
|Linux® on IBM® Z and LinuxONE| 200| 300| 230| 350|

 *_where m stands for Millicores, and Mi for Mebibytes_

If your environment is smaller, or bigger than the default, you can change the limits and resources for the application deployment by editing the IBMLicensing instance. For more information, see [Configuration](Configuration.md#modifying-the-application-deployment-resources).

The following table shows the available deployment profiles with the respective resource requirements for Linux x86_64.

|Profile|Environment|CPU Limit (m)|Memory Limit (Mi) |
|---|---|---|---|
|small|200 pods and 3 Cloud Paks|200 | 850|
|medium|500 pods and 3 Cloud Paks|300| 850|
|large|1000 pods and 3 Cloud Paks|300| 1020|

**Note:** When you have additional software, solution or plugin deployed in your cluster that might require additional memory or CPU resources, for example Dynatrace, check the documentation of this product and add additional resources to prevent memory saturation.

### Minimal resource requirements

For minimal resource requirements for License Service, see License Service requirements in [Hardware requirements of small profile](https://www.ibm.com/docs/en/cpfs?topic=services-hardware-requirements-small-profile).

## Hyperthreading

License Service supports multiple threads per physical core also referred to as Simultaneous multithreading (SMT) or Hyper-Threading (HT).

For more information about how to enable hyperthreading in License Service, and examples, see [Hyperthreading](https://www.ibm.com/docs/en/cpfs?topic=operator-hyperthreading).

## Cluster permissions

The IBM License Service operator requires certain cluster-level permissions to perform the main operations. These permissions are closely tracked and documented so that users can understand any implications that they might have on other workloads in the cluster.

|**API group**| **Resources** | **Verbs**  | **Description**    |
|:------------:|--------------|-------------|--------------------|
|" "|pods </br> namespaces </br> nodes|Get </br> List|The cluster permissions for the `ibm-license-service` service account are **read-only access** permissions that are required to properly discover the running {{site.data.keyword.IBM_notm}} applications to report license usage of the Virtual Processor Core (VPC) and Processor Value Unit (PVU) metrics.|
|operator.openshift.io|servicecas|List|These permissions are required to generate the TLS certificate for License Service. |
|operator.ibm.com| ibmlicensings </br> ibmlicenseservicereporters </br> ibmlicensings/status </br> ibmlicenseservicereporters/status </br> ibmlicensings/finalizers </br> ibmlicenseservicereporters/finalizers|Create </br> Delete </br> Get </br> List </br> Patch </br> Update </br> Watch| The cluster permissions for the `ibm-licensing-operator` service account are required to properly manage the status of the IBM License Service operator.|

<b>Related links</b>

- [Go back to home page](../License_Service_main.md#documentation)
- [Installing License Service](Installation_scenarios.md)
