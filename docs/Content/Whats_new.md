
# What's New in License Service

## License Service version 1.16.1

- Certification of License Service on VMware Tanzu Kubernetes Grid.

## License Service version 1.16.x

- For the products that have three-layer reporting structure that includes services that are grouped under bundled products, and are enabled for reporting, such as IBM Cloud Pak for Data, you can view the contribution of these services with License Service. Use the `/services` API to see how services contribute to the usage of the bundled products.

## License Service version 1.15.x

- License Service is enabled to report product metrics that come from Prometheus queries.

## License Service version 1.14.x and 1.13.x updates

No major updates.

## License Service version 1.12.x updates

- License Service is enabled to report the breakdown of the license usage of an IBM Cloud Pak by services. The information about the breakdown will be reported by License Service only when the IBM Cloud Pak implements the proper mechanism for reporting services.

- License Service supports an advanced and customizable authorization method that is based on a service account token, which is provided by the request header field. Service account token authentication provides flexible and customizable way to manage access to License Service APIs which is based on a role-based access control (RBAC).

## License Service version 1.11.x, 1.10.x and 1.9.x updates

No major updates.

## License Service version 1.8.x updates

- License Service supports hyperthreading on worker nodes also referred to as Simultaneous multithreading (SMT) or Hyperthreading (HT). If your IBM software is deployed on a cluster that has SMT or HT enabled, hyperthreading might have great impact on your licensing costs.

## License Service version 1.7.x updates

- License Service can report open source products that are  managed and supported by IBM, for example, WebSphere Liberty that is managed through IBM Cloud Foundry Migration Runtime (CFMR).

## License Service version 1.6.x updates

No major updates.

## License Service version 1.5.x updates

- The `bundled_products` file in the audit snapshot and the `/bundled_products` API additionally include information about the license metric unit of an IBM Cloud Pak to which the bundled product contributes.

## License Service version 1.4.x updates

- You can retrieve the status page that contains the most important information about your deployments and their license usage.

- The `/products` API is extended to enable retrieving information about the products that are deployed in namespaces that belong to user-defined groups. Thanks to grouping, you can see how the internal organizations or business divisions in your company contribute to the overall license usage. This information can help you with establishing the potential chargeback.

## License Service version 1.3.x updates

No major updates.

### License Service version 1.2.x updates

- You can retrieve information about License Service version and health by using the new APIs.

### License service version 1.1.0 (installer version 3.4.0)

- License Service supports license usage measurements for stand-alone IBM Container Software on all Kubernetes-orchestrated clouds.

- License Service collects and measures the license usage of IBM Cloud Paks and their bundled products that are enabled for reporting and licensed with the Managed Virtual Server (MVS) license metric.

<b>Related links</b>

- [Go back to home page](../License_Service_main.md#documentation)
