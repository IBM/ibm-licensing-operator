
# Preparing for installation

- [Supported platforms](#supported-platforms)
- [Required resources](#required-resources)
- [Cluster permissions](#cluster-permissions)

## Supported platforms

<b>Linux x86_64</b>

License Service is supported on all Kubernetes-orchestrated clouds on Linux x86_64.

It was tested on the following systems:

- Red Hat OpenShift Container Platform 4.6 or newer
- Kubernetes 1.11.3 or higher
- IBM Cloud Kubernetes Services (IKS)
- Google Kubernetes Engine (GKE)
- Azure Kubernetes Service (AKS)
- Amazon EKS - Managed Kubernetes Service (EKS)
- Alibaba Cloud - Container Service for Kubernetes (ACK)

<b>Linux on Power (ppc64le), Linux on IBM Z and LinuxONE</b>

License Service is supported on Linux on Power (ppc64le), Linux on IBM Z and LinuxONE in the following scenarios:

 |System|Supported deployment scenario|
 |---|---|
 |<ul><li>Any cluster with pre-installed Operator Lifecycle Manager (OLM)</li></ul>|<ul><li>[Automatic installation using Operator Lifecycle Manager (OLM)](Automatic_installation.md)</li><li>[Manual installation on Kubernetes from scratch with `kubectl`](Install_from_scratch.md)</li><li>[Offline installation](Install_offline.md)</li></ul>|
|<ul><li>A cluster without Operator Lifecycle Manager (OLM)</li></ul>| <ul><li>[Offline installation](Install_offline.md)</li><li>[Manual installation without the Operator Lifecycle Manager (OLM)](Install_without_OLM.md)</li></ul>|

## Required resources

License Service consists of two main components that require resources: the operator deployment and the application deployment.

 |Parameter|Operator|Application|Overall resources|
 |---|---|---|---|
 |CPU Limits| 20m| 500m| |
 |Memory Limits| 150Mi|512Mi|
 |CPU Requests| 10m|200m|**210m**|
 |Memory Requests|50Mi|256Mi|**306Mi**|

 *_where m stands for Millicores, and Mi for Mebibytes_

 **Note:** You can modify the limits and requests for the application deployment by editing the IBMLicensing instance. For more information, see [Configuration](Configuration.md).

### Minimal resource requirements

For minimal resource requirements for License Service, see License Service requirements in [Hardware requirements of small profile](https://www.ibm.com/docs/en/cpfs?topic=services-hardware-requirements-small-profile).

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
