**IMPORTANT:** The `master` branch contains the currently developed version of License Service and its content should not be used. Switch to another branch to view the content for the already-released version of License Service, for example `release-<version>` branch.

You can install License Service with ibm-licensing-operator to collect license usage information in two scenarios:

- [License Service as a part of an IBM Cloud Pak (included in IBM Cloud Pak foundational services)](#ibm-licensing-operator)
- [License Service without an IBM Cloud Pak](#ibm-licensing-operator-for-deploying-license-service-without-an-ibm-cloud-pak)

# ibm-licensing-operator

<b>Scenario: License Service as a part of an IBM Cloud Pak (included in IBM Cloud Pak foundational services)</b>

> **Important:** Do not install this operator directly. Only install this operator using the IBM IBM Cloud Pak foundational services operator. For more information about installing this operator and other foundational services operators, see [Installer documentation](http://ibm.biz/cpcs_opinstall). If you are using this operator as part of an IBM Cloud Pak, see the documentation for that IBM Cloud Pak to learn more about how to install and use the operator service. For more information about IBM Cloud Paks, see [IBM Cloud Paks that use IBM Cloud Pak foundational services](http://ibm.biz/cpcs_cloudpaks).

You can use the `ibm-licensing-operator` to install License Service as a part of IBM Cloud Pak foundational services or an IBM Cloud Pak. You can use License Service to collect information about license usage of IBM containerized products and IBM Cloud Paks per cluster. You can retrieve license usage data through a dedicated API call and generate an audit snapshot on demand.

For more information about the available IBM Cloud Pak foundational services, see the [IBM Documentation](http://ibm.biz/cpcsdocs).

## Supported platforms

Red Hat OpenShift Container Platform 4.2 or newer installed on Linux x86_64, Linux on Power (ppc64le), Linux on IBM Z and LinuxONE.

## Operator versions

- 1.0.0, 1.1.0, 1.1.1, 1.1.2, 1.1.3, 1.2.2, 1.2.3, 1.3.1, 1.4.1, 1.5.0, 1.6.0, 1.7.0, 1.8.0, 1.9.0, 1.10.0, 1.11.0, 1.12.0, 1.13.0, 1.14.0, 1.15.0, 1.16.0, 1.17.0, 1.18.0, 1.19.0, 1.20.0, 4.0.0, 4.1.0, 4.2.0, 4.2.1, 4.2.2, 4.2.3, 4.2.4, 4.2.5, 4.2.6, 4.2.7, 4.2.8, 4.2.9, 4.2.10, 4.2.11, 4.2.12

## Prerequisites

Before you install this operator, you need to first install the operator dependencies and prerequisites:

- For the list of operator dependencies, see [Dependencies of the IBM Cloud Pak foundational services](http://ibm.biz/cpcs_opdependencies) in the IBM Documentation.
- For the list of prerequisites for installing the operator, see [Preparing to install services documentation](http://ibm.biz/cpcs_opinstprereq) in the IBM Documentation.

> **Important:** If you installed License Service with the stand-alone IBM containerized software and you want to install an IBM Cloud Pak, it is recommended to first uninstall License Service from every cluster. Before uninstalling, the best practice is to retrieve an audit snapshot to ensure no data is lost. The Cloud Pak will install a new instance of License Service. This is a temporary action that we would like to automate in the future.

## Documentation

To install the operator with the IBM Cloud Pak foundational services Operator follow the installation and configuration instructions within the IBM Documentation.

- If you are using the operator as part of IBM Cloud Pak, see the documentation for that IBM Cloud Pak. For a list of IBM Cloud Paks, see [IBM Cloud Paks that use IBM Cloud Pak foundational services](http://ibm.biz/cpcs_cloudpaks).
- If you are using the operator with an IBM Containerized Software as a part of IBM Cloud Pak foundational services, see the [Installer documentation](http://ibm.biz/cpcs_opinstall) in IBM Documentation.

## SecurityContextConstraints Requirements

License Service supports running with the OpenShift Container Platform 4.3 default restricted Security Context Constraints (SCCs).

For more information about the OpenShift Container Platform Security Context Constraints, see [Managing Security Context Constraints](https://docs.openshift.com/container-platform/4.3/authentication/managing-security-context-constraints.html).

# ibm-licensing-operator for deploying License Service without an IBM Cloud Pak

<!--- This documentation is linked under the following short link: https://ibm.biz/license_service4containers. If content is moved update the link through the: Hybrid Cloud ID Team
--->

<b>Scenario: Learn how to deploy License Service on Kubernetes clusters without an IBM CLoud Pak</b>

You can use the `ibm-licensing-operator` to install License Service on any Kubernetes cluster without an IBM Cloud Pak. License Service collects information about license usage of IBM Containerized Products. You can retrieve license usage data through a dedicated API call and generate an audit snapshot on demand.

## Product documentation

For the overview and documentation, see [License Service deployment without an IBM Cloud for IBM stand-alone IBM Containerized Software](https://ibm.biz/license_service4containers).

<!--- The short link: https://ibm.biz/license_service4containers contains documentation for License Service stand-alond intended for IBM stand alone COntainerized Software. To have this link updated contact the Foundational services ID team, manager: Dan Hawkins.
--->

