# Retrieving license usage data from the cluster

- [Available APIs](#available-apis)
- [Obtaining a status page](#obtaining-a-status-page)
- [Tracking license usage in multicluster environment](#tracking-license-usage-in-multicluster-environment)

## Available APIs

License Service collects and measures information about license usage of your IBM containerized software per cluster.

You can use a set of dedicated APIs to retrieve the following information:

|API|Decription|
|---|---|
|Retrieving an audit snapshot|Retrieve an audit snapshot that is needed to fulfill the requirements of Container Licensing rules.|
|Retrieving license usage of products on a cluster level and retrieving license usage of product deployed in user-defined groups|Retrieve the list of products that are deployed on the cluster with their license usage. Use this API to integrate with external tools to monitor you license usage.|
|Retrieving license usage of bundled products|Retrieve the list of bundled products that are deployed on the cluster with their license usage. Use this API to integrate with external tools to monitor you license usage.|
|Retrieving contribution of services|Retrieve the contribution of services in the overall license usage of your bundled products. This information is additional and applies only to limited number of enabled products.|
|Retrieving information about License Service version|Retrieve information about License Service version for troubleshooting or upgrade purposes.|
|Retrieving information about License Service health|Retrieve information about License Service health to identify problems.|

To learn how to use the APIs to retrieve data and generate an audit snapshot, see [APIs for retrieving License Service data in IBM Documentation](https://www.ibm.com/docs/en/cpfs?topic=data-per-cluster-from-license-service).

## Obtaining a status page

You can obtain the status page that contains the most important information about your deployments and their license usage. You can use this data for analysis or troubleshooting.

This feature is available from IBM License Service version 1.4.0.

The status page is an html page that summarizes the most important information about your deployments on the cluster. The status page always shows the information that are collected and valid at the moment of page retrieval.

To retrieve the status page, complete the following actions:

1\. Obtain the License Service URL.

- For Openshift: [Obtain a License Service URL from the route](https://www.ibm.com/docs/en/cpfs?topic=pcfls-apis#ls_url)
- For Kubernetes: The URL depands on ingress configuration. For more information, see [Configuring ingress](Configuration.md#configuring-ingress)

2\. Get the authentication token. For more information, see [API authentication for License Service](https://www.ibm.com/docs/en/cpfs?topic=service-api-authentication).

3\. Open the License Service URL in your browser.

4\. Select the status page.

5\. Provide your authentication token.

For more information, see [Obtaining status page](https://www.ibm.com/docs/en/cpfs?topic=service-obtaining-status-page).

## Tracking license usage in multicluster environment

You can use the data that is collected by License Service from individual clusters to track the cumulative license usage in a multicluster environment. The cumulative report is not required for audit, but it might give you a full overview of your license usage in the multicluster environment.
For more information, see the overview in [Tracking license usage in multicluster environment](https://www.ibm.com/docs/en/cpfs?topic=operator-tracking-license-usage-in-multicluster-environment).

You can use the non-automated procedure to create a cumulative report for your environment. For more information, see [Manually tracking license usage in multicluster environment in IBM Documentation](https://www.ibm.com/docs/en/cpfs?topic=environment-manually-tracking-license-usage-in-multicluster).

### License Service Reporter

**Note:** License Service has an extension for tracking license usage in multicluster environment called License Service Reporter. However, at this point License Service Reporter is only supported with License Service instance that is shipped with IBM Cloud Pak foundational services and integrated with IBM Cloud Pak solutions.

If you have an IBM Cloud Pak deployed in your environment and you deployed License Service Reporter, you can configure an OpenShift or non-OpenShift cluster to deliver licensing data  from the License Service instance to License Service Reporter.
For more information, about how to configure your Kubernetes cluster as License Service Reporter data source, see [Configuring data sources](https://www.ibm.com/docs/en/cpfs?topic=reporter-configuring-data-sources).

For more information about License Service Reporter, see [Tracking license usage in multicluster environment with License Service Reporter](https://www.ibm.com/docs/en/cpfs?topic=tluime-tracking-license-usage-in-multicluster-environment-license-service-reporter).

<b>Related links</b>
- [Go back to home page](../License_Service_main.md#documentation)
