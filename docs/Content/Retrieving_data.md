# Retrieving license usage data from the cluster

- [Available APIs](#available-apis)
- [Obtaining a status page](#obtaining-a-status-page)
- [Tracking license usage in multicluster environment](#tracking-license-usage-in-multicluster-environment)

## Available APIs

License Service collects and measures information about license usage of your IBM containerized software per cluster.

You can use a set of dedicated APIs to retrieve the following information:

- Retrieving an audit snapshot for the cluster
- Retrieving license usage of products on a cluster level and retrieving license usage of product deployed in user-defined groups (chargeback)
- Retrieving license usage of bundled products on the cluster level
- Retrieving information about License Service version
- Retrieving information about License Service health

To learn how to use the APIs to retrieve data and generate an audit snapshot, see [APIs for retrieving License Service data in IBM Documentation](https://www.ibm.com/docs/en/cpfs?topic=data-per-cluster-from-license-service).

## Obtaining a status page

You can obtain the status page that contains the most important information about your deployments and their license usage. You can use this data for analysis or troubleshooting.

This feature is available from IBM License Service version 1.4.0.

The status page is an html page that summarizes the most important information about your deployments on the cluster. The status page always shows the information that are collected and valid at the moment of page retrieval.

To retrieve the status page, complete the following actions:
1. Obtain the License Service URL.

  - For Openshift: [Obtain a License Service URL from the route](https://www.ibm.com/docs/en/cpfs?topic=pcfls-apis#ls_url)
  - For Kubernetes: The URL depands on ingress configuration. For more information, see [Configuring ingress](Configuration.md#configuring-ingress)

2. Get the authentication token.
3. Open the License Service URL in your browser.
4. Select the status page.
5. Provide your authentication token.

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
