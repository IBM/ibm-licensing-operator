# Upgrading to License service 1.4.x from an earlier version

Learn how to upgrade to License Service version 1.4.x from an earlier version.

To upgrade to License Service 1.4.x, you must manually update the subscription channel. Complete the following actions to complete the upgrade.

1. Log in to your cluster.
2. Update the subscription channel by running the following command.

    ```bash
    licensingNamespace=<license_service_namespace>
    subName=ibm-licensing-operator-app
    kubectl patch subscription ${subName} -n ${licensingNamespace} --type=merge -p '{"spec": {"channel":"v3"}}'
    subscription.operators.coreos.com/ibm-licensing-operator-app patched
    ```

    where <license_service_namespace> is the namespace where you installed License Service. For example, `ibm-common-services`.

3. Wait until the `ClusterServiceVersion` status changes to **Succeeded**. To check the status of `ClusterServiceVersion`, run the following command. 

    ```
    bash
    csv_name=$(kubectl get subscription -n "${licensingNamespace}" ibm-licensing-operator-app -o jsonpath='{.status.currentCSV}')
    kubectl get csv -n "${licensingNamespace}" "${csv_name}" -o jsonpath='{.status.phase}'
    ```

After you update the subscription channel, License Service is automatically upgraded to version 1.4.x. In the future, updates will be automatic.

<b>Related links</b>

* [Go back to home page](../License_Service_main.md#documentation)
* [Retrieving license usage data from the cluster](Retrieving_data.md)
* [Uninstalling](Uninstalling.md)
* [Offline installation](Install_offline.md)
