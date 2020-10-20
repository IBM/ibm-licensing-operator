
# Preparing for installation

- [Supported platforms](#supported-platforms)
- [Required resources](#required-resources)

## Supported platforms

<b>Linux x86_64</b>

License Service is supported on all Kubernetes-orchestrated clouds on Linux x86_64.

It was tested on the following systems:

- Red Hat OpenShift Container Platform 3.11, 4.1, 4.2, 4.3 or newer
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
|<ul><li>A cluster without Operator Lifecycle Manager (OLM)</li></ul>| <ul><li>[Offline installation](Install_offline.md)</li></ul>|

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

<b>Related links</b>

- [Go back to home page](../License_Service_main.md#documentation)
- [Installing License Service](Installation_scenarios.md)
