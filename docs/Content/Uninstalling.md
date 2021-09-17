# Uninstalling License Service from a Kubernetes cluster

Complete the following procedure to uninstall License Service from your Kubernetes cluster.

**Note:** The following procedure assumes that you have deployed IBM License Service in the `ibm-common-services` namespace.

<b>Before you begin</b>

Before uninstalling License Service, create an audit snapshot to record your license usage until the uninstallation for audit purposes.
If you plan to reinstall License Service, the license usage data is stored in the persistent cluster memory and should not be affected by reinstallation. However, it is still a good practice to create an audit snapshot before reinstalling License Service as a precaution.

Complete the following steps to uninstall License Service in online and offline environments.

- [Step 1: Deleting the IBM Licensing resource](#step-1-deleting-the-ibm-licensing-resource)
- [Step 2: Uninstalling License Service](#step-2-uninstalling-license-service)
    - [Online uninstallation](#online-uninstallation)
    - [Offline uninstallation](#offline-uninstallation)

## Step 1: Deleting the IBM Licensing resource

1\. Delete the `IBMLicensing custom` resource.

Delete the instance and the operator will clean its resources.
First, check what `ibmlicensing` instances you have by running the following command:

```bash
licensingNamespace=ibm-common-services
kubectl get ibmlicensing -n ${licensingNamespace} -o jsonpath="{range .items[*]}{.metadata.name}{'\n'}"
```

The command should return one instance. Delete this instance, if it exists with the following command:

```bash
licensingNamespace=ibm-common-services
instanceName=`kubectl get ibmlicensing -n ${licensingNamespace} -o jsonpath="{range .items[*]}{.metadata.name}{'\n'}"`
kubectl delete ibmlicensing ${instanceName} -n ${licensingNamespace}
```

## Step 2: Uninstalling License Service

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
```

4\. Delete Operator Group.

**Note:** If you have other subscriptions that are tied with that operatorGroup do not delete it.
IBM Licensing Operator is now uninstalled.You can also clean up the operatorgroup that you created for subscription by using the following command:

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

**Note:** Do not uninstall OLM if it is used elsewhere, so if you want to use any other operators or when you have OCP cluster.

Uninstall OLM with the following command:

```bash
# Make sure GLOBAL_CATALOG_NAMESPACE has global catalog namespace value
kubectl delete crd clusterserviceversions.operators.coreos.com \
installplans.operators.coreos.com \
subscriptions.operators.coreos.com \
catalogsources.operators.coreos.com \
operatorgroups.operators.coreos.com
kubectl delete namespace ${GLOBAL_CATALOG_NAMESPACE}
```

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
kubectl delete ClusterRoleBinding ibm-license-service
kubectl delete ServiceAccount ibm-license-service -n ${licensingNamespace}
kubectl delete Role ibm-license-service -n ${licensingNamespace}
kubectl delete ClusterRole ibm-license-service
# delete rbac for operator:
kubectl delete RoleBinding ibm-licensing-operator -n ${licensingNamespace}
kubectl delete ClusterRoleBinding ibm-licensing-operator
kubectl delete ServiceAccount ibm-licensing-operator -n ${licensingNamespace}
kubectl delete Role ibm-licensing-operator -n ${licensingNamespace}
kubectl delete ClusterRole ibm-licensing-operator
```

3\. Remove the remaining License Service elements.

- If you pushed the IBM Licensing Docker images to your private registry, delete the images directly from that registry.

- Delete the images from the system where you downloaded the IBM Licensing images that you later pushed to your private registry with the following command.

```bash
# on machine with access to internet
export my_docker_registry=<YOUR REGISTRY IMAGE PREFIX HERE e.g.: "my.registry:5000" or "quay.io/opencloudio">
export operator_version=$(git tag | tail -n1 | tr -d v)
export operand_version=$(git tag | tail -n1 | tr -d v)
# remove images
docker rmi quay.io/opencloudio/ibm-licensing-operator:${operator_version}
docker rmi ${my_docker_registry}/ibm-licensing-operator:${operator_version}
docker rmi quay.io/opencloudio/ibm-licensing:${operand_version}
docker rmi ${my_docker_registry}/ibm-licensing:${operand_version}
# you might want to check if you don't have other images and delete them as well:
docker images | grep ibm-licensing
```

- If you cloned the [ibm-licensing-operator repository](https://github.com/IBM/ibm-licensing-operator) into your local system, delete it.

**Results**: License Service offline installation is completely removed and License Service uninstallation is completed.

<b>Related links</b>

- [Go back to home page](../License_Service_main.md#documentation)
