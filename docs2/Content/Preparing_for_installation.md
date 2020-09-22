
# Preparing for installation

- Preparing for installation
  - Supported platforms and versions
  - Operator versions
  - Cleaning existing License Service dependencies 

  Questions: Resources?

## Supported platforms

**Linux x86_64**

License Service is supported on all Kubernetes-orchestrated clouds on Linux x86_64.
   
It was tested on the following systems:
- Red Hat OpenShift Container Platform 3.11, 4.1, 4.2, 4.3 or newer
- Kubernetes 1.11.3 or higher
- IBM Cloud Kubernetes Services (IKS)
- Google Kubernetes Engine (GKE)
- Azure Kubernetes Service (AKS)
- Amazon EKS - Managed Kubernetes Service (EKS)
- Alibaba Cloud - Container Service for Kubernetes (ACK)
   
**Linux on Power (ppc64le), Linux on IBM Z and LinuxONE**

 License Service is supported on Linux on Power (ppc64le), Linux on IBM Z and LinuxONE in the following scenarios:

 |System|Supported deployment scenario|
 |---|---|
 |<ul><li>Red Hat OpenShift Container Platform 3.11, 4.1, 4.2, 4.3 or newer</li><li>Any cluster with pre-installed Operator Lifecycle Manager (OLM)</li></ul>|<ul><li>[Automatic installation using Operator Lifecycle Manager (OLM)](Automatic_installation.md)</li><li>[Manual installation on Kubernetes from scratch with `kubectl`](Install_from_scratch.md)</li><li>[Offline installation](Install_offline.md)</li></ul>|
|<ul><li>A cluster without Operator Lifecycle Manager (OLM)</li></ul>| <ul><li>[Offline installation](Install_offline.md)</li></ul>|

## Operator versions

- 1.0.0, 1.1.0, 1.1.1, 1.1.2, 1.1.3, 1.2.0

**Related links**

- [Go back to home page](../License_Service_main.md#documentation)
- [Installing License Service](Installation_scenarios.md)
