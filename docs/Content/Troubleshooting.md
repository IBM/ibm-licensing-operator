# Troubleshooting

- [Verifying completeness of license usage data](#verifying-completeness-of-license-usage-data)
- [License Service pods are crashing and License Service cannot run](#license-service-pods-are-crashing-and-license-service-cannot-run)
- [License Service API is unavailable with 503 Service Unavailable error](#license-service-api-is-unavailable-with-503-service-unavailable-error)

## Verifying completeness of license usage data

You can verify if License Service is properly deployed and whether it collects the complete license usage data. For more information, see [Verifying completeness of license usage data and troubleshooting](https://www.ibm.com/docs/en/cpfs?topic=operator-verifying-completeness-license-usage-data-troubleshooting).

## License Service pods are crashing and License Service cannot run

If your License Service pods are crashing, and you see multiple instances of License Service with the `CrashLoopBackOff` status in your OpenShift console, you might have License Service deployed to more than one namespace. As a result, two License Service operators are running in two namespaces and the service crashes. The `ibm-licensing-operator` should only be deployed in the `ibm-common-services` namespace, however, if you deployed License Service more than once, the older version of License Service might be deployed to `kube-system` namespace.

Complete the following steps to fix the problem:

1\. To check whether the `ibm-licensing-operator` is deployed to `kube-system` namespace, run the following command:

- **Linux:** `kubectl get pod -n kube-system | grep ibm-licensing-operator`
- **Windows:** `kubectl get pod -n kube-system | findstr ibm-licensing-operator`

2\. If the response contains information about the running pod, uninstall License Service from `kube-system` namespace.

<b>Related links</b>

## License Service API is unavailable with 503 Service Unavailable error

You might get 503 Service Unavailable error when you make a License Service API call when you use the custom ingress certificate. The custom ingress certificate is not acceptable for License Service. To fix this issue, complete the following actions:

1\. Generate the correct certificate for Fully Qualified Domain Name (FQDN) of License Service. To check the License Service URL, go to the OpenShift console, go to **Networking** > **Routes**. Find the `ibm-licensing-service-instance` route. The License Service URL is listed as **Location**.

2\. [Configure your custom certificate](Configuration.md#using-custom-certificates).

- [Go back to home page](../License_Service_main.md#documentation)
