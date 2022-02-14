# Backup and upgrade

* [Backup](#backup)
* [Upgrade](#upgrade)

## Backup

The license usage data that is collected by License Service is stored in the persistent cluster memory and is not affected when you kill or restart a pod.

Nonetheless, it is a good practice to generate an audit snapshot periodically for backup purposes and store it in a safe location. You do not need to perform any other backup.

**Note:** Before decommissioning a cluster, record the license usage of the products that are deployed on this cluster by generating an audit snapshot until the day of decommissioning.

## Upgrade

* For online environments, License Service is automatically upgraded with each new operator release.
* For online environments, to upgrade to License Service version 1.4.x from an earlier version, you must manually update the subscription channel. For more information, see [Updating the subscription channels](#updating-the-subscription-channel).
* For online environments, to upgrade from License Service version 1.10.x to the latest version, you must manually update the `CatalogSource` image. For more information, see [Updating the CatalogSource image](#updating-thecatalogsource-image).
* For online environments, to upgrade from License Service version 1.3.x or earlier, to the latest version first update the subscription channel, and next update the `CatalogSource`. For more information, see [Updating the subscription channels](#updating-the-subscription-channel) and [Updating the CatalogSource image](#updating-thecatalogsource-image). 
* For offline environments, to upgrade License Service to a new version, first uninstall License Service from the cluster and redeploy it.

    **Note:** The license usage data is stored in the persistent cluster memory and should not be affected by reinstallation of License Service. However, it is a good practice to create an audit snapshot before reinstalling License Service as a precaution.
    
### Updating the subscription channel

Learn how to upgrade to License Service version 1.4.x from an earlier version.

To upgrade to License Service 1.4.x, you must manually update the subscription channel. Complete the following actions to complete the upgrade.

1\. Log in to your cluster.

2\. Update the subscription channel by running the following command.

    ```bash
    licensingNamespace=ibm-common-services
    subName=ibm-licensing-operator-app
    kubectl patch subscription ${subName} -n ${licensingNamespace} --type=merge -p '{"spec": {"channel":"v3"}}'
    subscription.operators.coreos.com/ibm-licensing-operator-app patched
    ```

   **Note:** If you installed License Service in a custom namespace, change the value of `licensingNamespace` from the default `ibm-common-services` to your custom namespace.

3\. Wait until the `ClusterServiceVersion` status changes to **Succeeded**. To check the status of `ClusterServiceVersion`, run the following command.

    ```bash
    csv_name=$(kubectl get subscription -n "${licensingNamespace}" ibm-licensing-operator-app -o jsonpath='{.status.currentCSV}')
    kubectl get csv -n "${licensingNamespace}" "${csv_name}" -o jsonpath='{.status.phase}'
    ```

After you update the subscription channel, License Service is automatically upgraded to version 1.4.x. In the future, updates will be automatic.

### Updating the CatalogSource image

Learn how to upgrade from License Service version 1.10.x to the latest version.

Starting from November 2021, IBM no longer publishes the catalog image updates to `docker.io` which was used by default in IBM License Service installation scripts and procedures that depended on `CatalogSource`.
If you installed IBM License Service before December 2021, you must perform manual steps to update `CatalogSource` image and upgrade License Service version. 

To upgrade to the latest version of License Service, you must manually update the `CatalogSource` image. Complete the following actions to complete the upgrade.

1\. Log in to your cluster.

2\. Update the `CatalogSource` by running the following command.

```bash
    catalogSourceNamespace=openshift-marketplace
    csName=opencloud-operators
    kubectl patch catalogsource ${csName} -n ${catalogSourceNamespace} --type=merge -p '{"spec": {"image":"icr.io/cpopen/ibm-operator-catalog"}}'
```

   **Note:** Default `catalogSourceNamespace` on OpenShift is `openshift-marketplace`. If you are using Kubernetes change it to your `CatalogSource` namespace.

   **Note:** If you deployed License Service `CatalogSource` with a custom name, change the value of `csName` from the default `opencloud-operators` to your custom name.

3\. Wait until the `ClusterServiceVersion` status changes to **Succeeded**. To check the status of `ClusterServiceVersion`, run the following command.

```bash
    licensingNamespace=ibm-common-services
    csv_name=$(kubectl get subscription -n "${licensingNamespace}" ibm-licensing-operator-app -o jsonpath='{.status.currentCSV}')
    kubectl get csv -n "${licensingNamespace}" "${csv_name}" -o jsonpath='{.status.phase}'
```

After you update the `CatalogSource` image, License Service is automatically upgraded to the latest version. In the future, updates will be automatic.

<b>Related links</b>

* [Go back to home page](../License_Service_main.md#documentation)
* [Retrieving license usage data from the cluster](Retrieving_data.md)
* [Uninstalling](Uninstalling.md)
* [Offline installation](Install_offline.md)
