# Uninstalling License Service

Complete the following procedure to uninstall License Service.

**Note:** The following procedure assumes that you have deployed IBM License Service in the `ibm-common-services` namespace.

<b>Before you begin</b>

Before uninstalling License Service, create an audit snapshot to record your license usage until the uninstallation for audit purposes.
If you plan to reinstall License Service, the license usage data is stored in the persistent cluster memory and should not be affected by reinstallation. However, it is still a good practice to create an audit snapshot before reinstalling License Service as a precaution.

Uninstallation path is linked to the installation method. The following table shows how to uninstall License Service depanding on the deployment scenario that was used for installation.

|Uninstallation path| Deployment scenario |
|---|---|
|Follow **online uninstallation** steps|<ul><li>[Automatic installation using Operator Lifecycle Manager (OLM)](Automatic_installation.md)</li><li>[Manual installation on OpenShift Container Platform (OCP) version 4.6 or later](Install_on_OCP.md)</li><li>[Manual installation on Kubernetes from scratch with `kubectl`](Install_from_scratch.md)</li></ul>|
|Follow **offline uninstallation** steps|<ul><li>[Offline installation](Install_offline.md)</li><li>[Manual installation without Operator Lifecycle Manager (OLM)](Install_without_OLM.md)</li></ul>|

Complete the following steps to uninstall License Service.

- [Step 1: Deleting the IBM Licensing instance](#step-1-deleting-the-ibm-licensing-instance)
- [Step 2: Deleting the remaining License Service resources](#step-2-deleting-the-remaining-license-service-resources)
    - [Online uninstallation](#online-uninstallation)
    - [Offline uninstallation](#offline-uninstallation)
- [Step 3: Verifying uninstallation](#step-3-verifying-uninstallation)
    - [Online](#online)
    - [Offline](#offline) 

## Step 1: Deleting the IBM Licensing instance

Delete the `IBMLicensing custom` resource.

Delete the instance and the operator will clean its resources.

1\. Check what `ibmlicensing` instances you have by running the following command:

```bash
licensingNamespace=ibm-common-services
kubectl get ibmlicensing -n ${licensingNamespace} -o jsonpath="{range .items[*]}{.metadata.name}{'\n'}"
```

The command should return one instance.

2\. Delete this instance, if it exists with the following command:

```bash
licensingNamespace=ibm-common-services
instanceName=`kubectl get ibmlicensing -n ${licensingNamespace} -o jsonpath="{range .items[*]}{.metadata.name}{'\n'}"`
kubectl delete ibmlicensing ${instanceName} -n ${licensingNamespace}
```

## Step 2: Deleting the remaining License Service resources

**Note:** The number and type of resources that are created on a cluster depend on a version and License Service installation method. Because of that, some resources that are listed might not be present.

Select the procedure for your environment:

- [Online uninstallation](#online-uninstallation)
- [Offline uninstallation](#offline-uninstallation)

### Online uninstallation

1\. Delete the operator subscription.

Run the following command to see your subscriptions:

```bash
licensingNamespace=ibm-common-services
kubectl get subscription -n ${licensingNamespace} -o jsonpath="{range .items[*]}{.metadata.name}{'\n'}"
```

Delete the `ibm-licensing-operator-app` subscription by using the following command:

```bash
licensingNamespace=ibm-common-services
subName=ibm-licensing-operator-app
kubectl delete subscription ${subName} -n ${licensingNamespace}
```

2\. Delete Cluster Service Version (CSV).

Delete CSV that manages the Operator image.
Run the following command to get your CSV name, look for `ibm-licensing-operator`:

```bash
licensingNamespace=ibm-common-services
kubectl get clusterserviceversion -n ${licensingNamespace} -o jsonpath="{range .items[*]}{.metadata.name}{'\n'}"
kubectl get clusterserviceversion -n ${licensingNamespace} -o jsonpath="{range .items[*]}{.metadata.name}{'\n'}" | grep ibm-licensing-operator
```

Delete it by using the following command:

```bash
licensingNamespace=ibm-common-services
csvName=`kubectl get clusterserviceversion -n ${licensingNamespace} -o jsonpath="{range .items[*]}{.metadata.name}{'\n'}" | grep ibm-licensing-operator`
kubectl delete clusterserviceversion ${csvName} -n ${licensingNamespace}
```

3\. Delete Custom Resource Definition (CRD).

Delete the custom resource definition with the following command:

```bash
kubectl delete CustomResourceDefinition ibmlicensings.operator.ibm.com
kubectl delete CustomResourceDefinition ibmlicenseservicereporters.operator.ibm.com
kubectl delete CustomResourceDefinition ibmlicensingdefinitions.operator.ibm.com
kubectl delete CustomResourceDefinition ibmlicensingmetadatas.operator.ibm.com
kubectl delete CustomResourceDefinition ibmlicensingquerysources.operator.ibm.com
```

4\. Delete Operator Group.

**Note:** If you have other subscriptions that are tied with that operatorGroup do not delete it.
IBM Licensing is now uninstalled.You can also clean up the operatorgroup that you created for subscription by using the following command:

```bash
licensingNamespace=ibm-common-services
operatorGroupName=operatorgroup
kubectl delete OperatorGroup ${operatorGroupName} -n ${licensingNamespace}
```

5\. Delete CatalogSource.

**Note:** If you have other services that use the opencloudio CatalogSource do not delete it.
Otherwise, you can delete the CatalogSource with the following command:

```bash
# Make sure GLOBAL_CATALOG_NAMESPACE has global catalog namespace value.
opencloudioSourceName=opencloud-operators
kubectl delete CatalogSource ${opencloudioSourceName} -n ${GLOBAL_CATALOG_NAMESPACE}
```

6\. Uninstall OLM.

For more information, see [Uninstall in OLM documentation](https://github.com/operator-framework/operator-lifecycle-manager/blob/master/doc/install/install.md#uninstall).

### Offline uninstallation

1\. Delete the operator deployment by running the following command.

```bash
licensingNamespace=ibm-common-services
kubectl delete deployment ibm-licensing-operator -n ${licensingNamespace}
```

2\. Delete role-based access control (RBAC) with the following command.

```bash
# configure namespace:
licensingNamespace=ibm-common-services
# delete rbac for operand:
kubectl delete RoleBinding ibm-license-service -n ${licensingNamespace}
kubectl delete RoleBinding ibm-license-service-restricted -n ${licensingNamespace}
kubectl delete ClusterRoleBinding ibm-license-service
kubectl delete ClusterRoleBinding ibm-license-service-restricted
kubectl delete ClusterRoleBinding ibm-licensing-default-reader
kubectl delete ServiceAccount ibm-license-service -n ${licensingNamespace}
kubectl delete ServiceAccount ibm-license-service-restricted -n ${licensingNamespace}
kubectl delete ServiceAccount ibm-licensing-default-reader -n ${licensingNamespace}
kubectl delete Role ibm-license-service -n ${licensingNamespace}
kubectl delete Role ibm-license-service-restricted -n ${licensingNamespace}
kubectl delete ClusterRole ibm-license-service
kubectl delete ClusterRole ibm-license-service-restricted
kubectl delete ClusterRole ibm-licensing-default-reader
# delete rbac for operator:
kubectl delete RoleBinding ibm-licensing-operator -n ${licensingNamespace}
kubectl delete ClusterRoleBinding ibm-licensing-operator
kubectl delete ServiceAccount ibm-licensing-operator -n ${licensingNamespace}
kubectl delete Role ibm-licensing-operator -n ${licensingNamespace}
kubectl delete ClusterRole ibm-licensing-operator
```

3\. Delete Custom Resource Definition (CRD).

Delete the custom resource definition with the following command:

```bash
kubectl delete CustomResourceDefinition ibmlicensings.operator.ibm.com
kubectl delete CustomResourceDefinition ibmlicenseservicereporters.operator.ibm.com
kubectl delete CustomResourceDefinition ibmlicensingdefinitions.operator.ibm.com
kubectl delete CustomResourceDefinition ibmlicensingmetadatas.operator.ibm.com
kubectl delete CustomResourceDefinition ibmlicensingquerysources.operator.ibm.com
```

4\. Remove the remaining License Service elements.

- If you pushed the IBM Licensing Docker images to your private registry, delete the images directly from that registry.

- Delete the images from the system where you downloaded the IBM Licensing images that you later pushed to your private registry with the following command.

```bash
# on machine with access to internet
export my_docker_registry=<YOUR REGISTRY IMAGE PREFIX HERE e.g.: "my.registry:5000" or "quay.io/opencloudio">
export operator_version=$(git describe --tags `git rev-list --tags --max-count=1` | tr -d v)
export operand_version=$(git describe --tags `git rev-list --tags --max-count=1` | tr -d v)
# remove images
docker rmi icr.io/cpopen/ibm-licensing-operator:${operator_version}
docker rmi ${my_docker_registry}/ibm-licensing-operator:${operator_version}
docker rmi icr.io/cpopen/cpfs/ibm-licensing:${operand_version}
docker rmi ${my_docker_registry}/ibm-licensing:${operand_version}
# you might want to check if you don't have other images and delete them as well:
docker images | grep ibm-licensing
```

- If you cloned the [ibm-licensing-operator repository](https://github.com/IBM/ibm-licensing-operator) into your local system, delete it.

**Results**: License Service offline installation is completely removed and License Service uninstallation is completed.

<b>Related links</b>

- [Go back to home page](../License_Service_main.md#documentation)

## Step 3: Verifying uninstallation

To verify that License Service is fully uninstalled, complete the following steps:

### Online

1\. Verify whether the License Service Catalog Source is deleted.

**Note:** If you have other services that use the `opencloudio` Catalog Source, do not delete it.

```bash
kubectl get CatalogSource -A | grep opencloud-operators
```

As a response you should get an empty list.

2\. Verify whether the License Service Operator Group is deleted.

**Note:** If you have other subscriptions that are tied with this Operator Group, do not delete it.

```bash
licensingNamespace=ibm-common-services
operatorGroupName=operatorgroup
kubectl get OperatorGroup ${operatorGroupName} -n ${licensingNamespace}
```

As a response you should get an empty list.

3\. Verify whether the Custom Resource Definitions (CRDs) are deleted.

```bash
kubectl get CustomResourceDefinition | grep ibmlicens
```

As a response you should get an empty list.

4\.. Verify whether the Cluster Service Versions are deleted.

```bash
kubectl get clusterserviceversion -A | grep ibm-licensing-operator
```

As a response you should get an empty list.

5\. Verify whether the Subscription is deleted.

```bash
kubectl get subscription -A | grep ibm-licensing-operator
```

As a response you should get an empty list.

6\. Verify whether role-based access controls (RBAC) are deleted.

```bash
kubectl get ClusterRole | grep ibm-licens
kubectl get ClusterRoleBinding | grep ibm-licens
kubectl get Role -A | grep ibm-licens
kubectl get RoleBinding -A | grep ibm-licens
kubectl get ServiceAccount -A | grep ibm-licens
```

As a response you should get an empty list.

7\. Verify whether License Service deployments are deleted.

```bash
kubectl get Deployment -A | grep ibm-licens
```

As a response you should get an empty list.

### Offline

1\. Verify whether the Custom Resource Definitions (CRDs) are deleted.

```bash
kubectl get CustomResourceDefinition | grep ibmlicens
```

As a response you should get an empty list.

2\. Verify whether role-based access controls (RBAC) are deleted.

```bash
kubectl get ClusterRole | grep ibm-licens
kubectl get ClusterRoleBinding | grep ibm-licens
kubectl get Role -A | grep ibm-licens
kubectl get RoleBinding -A | grep ibm-licens
kubectl get ServiceAccount -A | grep ibm-licens
```

As a response you should get an empty list.

3\. Verify whether License Service deployments are deleted.

```bash
kubectl get Deployment -A | grep ibm-licens
```

As a response you should get an empty list.
