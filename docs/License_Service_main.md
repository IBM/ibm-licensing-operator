
# License Service for stand-alone IBM Containerized Software without IBM Cloud Paks

<b>Scenario: Learn how to deploy License Service on any Kubernetes cluster without an IBM Cloud Pak. See how to use License Service to record and track license usage of IBM Containerized Software.</b>

You can use the `ibm-licensing-operator` to install License Service on any Kubernetes cluster. License Service collects information about license usage of IBM containerized products. You can retrieve license usage data through a dedicated API call and generate an audit snapshot on demand.

**Note:** License Service is integrated into IBM Cloud Pak solutions. You do not have to deploy it to clusters where IBM Cloud Pak solutions are deployed. License Service should already be there and collect usage data for the IBM containerized products that are enabled for reporting.

Use the installation scenario that is outlined in this documentation to deploy License Service to a cluster where IBM Cloud Pak solutions are not deployed.

## About License Service

License Service is required for monitoring and measuring license usage of the IBM Containerized software in accord with the pricing rule for containerized environments. Manual license measurements are not allowed.

License Service collects and measures the license usage of your products at the cluster level. You can retrieve this data upon request for monitoring and compliance. You can also retrieve an audit snapshot of the data that is audit evidence.

License Service

- Collects and measures the license usage of Virtual Processor Core (VPC) and Processor Value Unit (PVU) metrics at the cluster level of IBM Containerized products that are enabled for reporting. To learn if your product is enabled for reporting, contact the product support.
- Currently, License Service refreshes the data every 5 minutes. However, this might be subject to change in the future. With this frequency, you can capture changes in a dynamic cloud environment.
- Provides the API that you can use to retrieve data that outlines the highest license usage on the cluster.
- Provides the API that you can use to retrieve an audit snapshot that lists the highest license usage values for the requested period for products that are deployed on a cluster.

## Using License Service for container licensing

Currently, supported core-based metrics for container licensing are Processor Value Unit (PVU) and Virtual Processor Core (VPC). For core license metrics, you are obliged to use License Service and periodically generate an audit snapshots to fulfill container licensing requirements.

For more information about core and non-core metrics that are collected by License Service, see [Reported metrics](https://www.ibm.com/support/knowledgecenter/SSHKN6/license-service/1.x.x/reported_metrics.html).

License Service collects data that is required for compliance and audit purposes. With License Service, you can retrieve an audit snapshot per cluster without any configuration.

At this point, it is required to generate an audit snapshot at least once a quarter for the last 90 days, and to store it for 2 years in a location from which it could be retrieved and delivered to auditors.

Note, that the requirements might change over time. You should always make sure to follow the latest requirements that are posted on Passport Advantage.

For more information, see the following resources:

- [IBM Container Licenses on Passport Advantage](https://www.ibm.com/software/passportadvantage/containerlicenses.html)
- [Container licensing FAQs](https://www.ibm.com/software/passportadvantage/containerfaqov.html)
- [How to: Retrieving an audit snapshot](https://www.ibm.com/support/knowledgecenter/SSHKN6/license-service/1.x.x/APIs.html#auditSnapshot)

<b>Best practices</b>

- It is recommended to generate an audit snapshot report monthly as a safety precaution.
- Before decommissioning a cluster, record the license usage of the products that are deployed on this cluster by generating an audit snapshot until the day of decommissioning.
- Plan your storage to contain regular audit snapshots. The size of an audit snapshot .zip package might vary and depends on the number of products and range of the reporting period. On average, the size of the package for a small environment is around 10 KB, and for medium and large environments - around 100 KB.

## Documentation

- [Preparing for installation](Content/Preparing_for_installation.md)
    - [Supported platforms](Content/Preparing_for_installation.md#supported-platforms)
    - [Required resources](Content/Preparing_for_installation.md#required-resources)
- [Installing License Service](Content/Installation_scenarios.md)
    - [Automatically installing ibm-licensing-operator with a stand-alone IBM Containerized Software using Operator Lifecycle Manager (OLM)](Content/Automatic_installation.md)
    - [Manually installing License Service on OCP 4.2+](Content/Install_on_OCP.md)
    - [Manual installation without the Operator Lifecycle Manager (OLM)](Content/Install_without_OLM.md)
    - [Manually installing License Service on Kubernetes from scratch with `kubectl`](Content/Install_from_scratch.md)
    - [Offline installation](Content/Install_offline.md)
- [Configuration](Content/Configuration.md)
    - [Configuring ingress](Content/Configuration.md#configuring-ingress)
    - [Checking License Service components](Content/Configuration.md#checking-license-service-components)
    - [Using custom certificates](Content/Configuration.md#using-custom-certificates)
    - [Cleaning existing License Service dependencies](Content/Configuration.md#cleaning-existing-license-service-dependencies)
    - [Modifying the application deployment resources](Content/Configuration.md#modifying-the-application-deployment-resources)
- [Retrieving license usage data from the cluster](Content/Retrieving_data.md)
    - [Available APIs](Content/Retrieving_data.md#available-apis)
    - [Retrieving an audit snapshot](https://www.ibm.com/support/knowledgecenter/SSHKN6/license-service/1.x.x/APIs.html#auditSnapshot)
    - [Tracking license usage in multicluster environment](Content/Retrieving_data.md#tracking-license-usage-in-multicluster-environment)
- [Uninstalling License Service from a Kubernetes cluster](Content/Uninstalling.md)
- [Backup and upgrade](Content/Backup_and_upgrade.md)
- [Troubleshooting](Content/Troubleshooting.md)
    - [Verifying completeness of license usage data](Content/Troubleshooting.md#verifying-completeness-of-license-usage-data)
    - [Preparing resources for offline installation without git](Content/Troubleshooting.md#prepareing-resources-for-offline-installation-without-git)
    - [License Service pods are crashing and License Service cannot run](Content/Troubleshooting.md#license-service-pods-are-crashing-and-license-service-cannot-run)
