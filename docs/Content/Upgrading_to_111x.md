# Upgrading to License service 1.11.x from an earlier version

Learn how to upgrade to License Service version 1.11.x from an earlier version.

To upgrade to License Service 1.11.x, you must manually update the CatalogSource image. Complete the following actions to complete the upgrade.

1\. Log in to your cluster.

2\. Update the CatalogSource by running the following command.

```bash
    catalogSourceNamespace=openshift-marketplace
    csName=opencloud-operators
    kubectl patch catalogsource ${csName} -n ${catalogSourceNamespace} --type=merge -p '{"spec": {"image":"icr.io/cpopen/ibm-operator-catalog"}}'
```

   **Note:** Default `catalogSourceNamespace` on OpenShift is `openshift-marketplace`. If you are using Kubernetes change it to your CatalogSource namespace.

   **Note:** If you deployed License Service CatalogSource with a custom name, change the value of `csName` from the default `opencloud-operators` to your custom name.

3\. Wait until the `ClusterServiceVersion` status changes to **Succeeded**. To check the status of `ClusterServiceVersion`, run the following command.

```bash
    csv_name=$(kubectl get subscription -n "${licensingNamespace}" ibm-licensing-operator-app -o jsonpath='{.status.currentCSV}')
    kubectl get csv -n "${licensingNamespace}" "${csv_name}" -o jsonpath='{.status.phase}'
```

After you update the CatalogSource image, License Service should be automatically upgraded to version 1.11.x. In the future, updates will be automatic.

<b>Related links</b>

* [Go back to home page](../License_Service_main.md#documentation)
* [Retrieving license usage data from the cluster](Retrieving_data.md)
* [Uninstalling](Uninstalling.md)
* [Offline installation](Install_offline.md)
