
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
 |<ul><li>Red Hat OpenShift Container Platform 3.11, 4.1, 4.2, 4.3 or newer</li><li>Any cluster with pre-installed Operator Lifecycle Manager (OLM)</li></ul>|<ul><li>[Automatically installing ibm-licensing-operator with a stand-alone IBM Containerized Software using Operator Lifecycle Manager (OLM)](#automatically-installing-ibm-licensing-operator-with-a-stand-alone-ibm-containerized-software-using-operator-lifecycle-manager-olm)</li><li>[Manually installing License Service on Kubernetes from scratch with `kubectl`](#manually-installing-license-service-on-kubernetes-from-scratch-with-kubectl)</li><li>[Offline installation](#offline-installation)</li></ul>|
|<ul><li>A cluster without Operator Lifecycle Manager (OLM)</li></ul>| <ul><li>[Offline installation](#offline-installation)</li></ul>|

## Operator versions

- 1.0.0, 1.1.0, 1.1.1, 1.1.2, 1.1.3, 1.2.0

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

## What's Next:
Installing License Service
