[![Go Report Card](https://goreportcard.com/badge/github.com/IBM/ibm-licensing-operator)](https://goreportcard.com/report/github.com/IBM/ibm-licensing-operator)
![kind](https://github.com/IBM/ibm-licensing-operator/workflows/kind/badge.svg)
![ocp](https://github.com/IBM/ibm-licensing-operator/workflows/ocp/badge.svg)
<img alt="Uptime Robot status" src="https://img.shields.io/uptimerobot/status/m786127186-a86f251061d6fd7958c67707?label=OCP%20test%20cluster">
[![Code Coverage](https://codecov.io/gh/IBM/ibm-licensing-operator/branch/master/graphs/badge.svg?branch=master)](https://codecov.io/gh/IBM/ibm-licensing-operator?branch=master)

You can install License Service with ibm-licensing-operator to collect license usage information in two scenarios:

- [License Service as a part of an IBM Cloud Pak (included in IBM Cloud Platform Common Services)](#ibm-licensing-operator)
- [License Service without an IBM Cloud Pak](#ibm-licensing-operator-for-deploying-license-service-without-an-ibm-cloud-pak)

# ibm-licensing-operator

<b>Scenario: License Service as a part of an IBM Cloud Pak (included in IBM Cloud Platform Common Services)</b>

> **Important:** Do not install this operator directly. Only install this operator using the IBM Common Services Operator. For more information about installing this operator and other Common Services operators, see [Installer documentation](http://ibm.biz/cpcs_opinstall). If you are using this operator as part of an IBM Cloud Pak, see the documentation for that IBM Cloud Pak to learn more about how to install and use the operator service. For more information about IBM Cloud Paks, see [IBM Cloud Paks that use Common Services](http://ibm.biz/cpcs_cloudpaks).

You can use the ibm-licensing-operator to install License Service as a part of IBM Cloud Platform Common Services or an IBM Cloud Pak. You can use License Service to collect information about license usage of IBM containerized products and IBM Cloud Paks per cluster. You can retrieve license usage data through a dedicated API call and generate an audit snapshot on demand.

For more information about the available IBM Cloud Platform Common Services, see the [IBM Knowledge Center](http://ibm.biz/cpcsdocs).

## Supported platforms

Red Hat OpenShift Container Platform 4.2 or newer installed on Linux x86_64, Linux on Power (ppc64le), Linux on IBM Z and LinuxONE.

> **Note:** On Red Hat OpenShift Container Platform 4.2

## Operator versions

- 1.0.0, 1.1.0, 1.1.1, 1.1.2, 1.1.3, 1.2.2, 1.2.3, 1.3.1, 1.3.2

## Prerequisites

Before you install this operator, you need to first install the operator dependencies and prerequisites:

- For the list of operator dependencies, see the IBM Knowledge Center [Common Services dependencies documentation](http://ibm.biz/cpcs_opdependencies).
- For the list of prerequisites for installing the operator, see the IBM Knowledge Center [Preparing to install services documentation](http://ibm.biz/cpcs_opinstprereq).

> **Important:** If you installed License Service with the stand-alone IBM containerized software and you want to install an IBM Cloud Pak, it is recommended to first uninstall License Service from every cluster. Before uninstalling, the best practice is to retrieve an audit snapshot to ensure no data is lost. The Cloud Pak will install a new instance of License Service. This is a temporary action that we would like to automate in the future.

## Documentation

To install the operator with the IBM Common Services Operator follow the the installation and configuration instructions within the IBM Knowledge Center.

- If you are using the operator as part of an IBM Cloud Pak, see the documentation for that IBM Cloud Pak. For a list of IBM Cloud Paks, see [IBM Cloud Paks that use Common Services](http://ibm.biz/cpcs_cloudpaks).
- If you are using the operator with an IBM Containerized Software as a part of IBM Cloud Platform Common Services, see the [Installer documentation](http://ibm.biz/cpcs_opinstall) in Knowledge Center.

## SecurityContextConstraints Requirements

License Service supports running with the OpenShift Container Platform 4.3 default restricted Security Context Constraints (SCCs).

For more information about the OpenShift Container Platform Security Context Constraints, see [Managing Security Context Constraints](https://docs.openshift.com/container-platform/4.3/authentication/managing-security-context-constraints.html).

# ibm-licensing-operator for deploying License Service without an IBM Cloud Pak

<!--- This documentation is linked under the following short link: https://ibm.biz/license_service4containers. If content is moved update the link through the: Hybrid Cloud ID Team
--->

<b>Scenario: Learn how to deploy License Service on Kubernetes clusters witout an IBM CLoud Pak</b>

You can use the `ibm-licensing-operator` to install License Service on any Kubernetes cluster without an IBM CLoud Pak. License Service collects information about license usage of IBM Containerized Products. You can retrieve license usage data through a dedicated API call and generate an audit snapshot on demand.

## Product documentation

For the overview and documentation, see [License Service deployment without an IBM Cloud Pak](docs/License_Service_main.md).

<!---
- [Preparing for installation](docs/Preparing_for_installation.md)
  - [Supported platforms](docs/Preparing_for_installation.md#supported-platforms)
  - [Operator versions](docs/Preparing_for_installation.md#operator-versions)
- [Installing License Service](docs/Installation_scenarios.md)
    - [Automatically installing ibm-licensing-operator with a stand-alone IBM Containerized Software using Operator Lifecycle Manager (OLM)](docs/Automatic_installation.md)
    - [Manually installing License Service on OCP 4.2+](docs/Install_on_OCP.md)
    - [Manually installing License Service on Kubernetes from scratch with `kubectl`](docs/Install_from_scratch.md)
    - [Offline installation](docs/Install_offline.md)
- [Configuration](docs/Configuration.md)
  - [Configuring ingress](docs/Configuration.md#configuring-ingress)
  - [Checking License Service components](docs/Configuration.md#checking-license-service-components)
  - [Using custom certificates](docs/Configuration.md#using-custom-certificates)
  - [Cleaning existing License Service dependencies](docs/Configuration.md#cleaning-existing-license-service-dependencies)
- [Retrieving license usage data from the cluster](docs/Retrieving_data.md)
  - [Available APIs](docs/Retrieving_data.md#available-apis)
  - [Tracking license usage in multicluster environment](docs/Retrieving_data.md#tracking-license-usage-in-multicluster-environment)
- [Uninstalling License Service from a Kubernetes cluster](docs/Uninstalling.md)
- [Troubleshooting](docs/Troubleshooting.md)
  - [Preparing resources for offline installation without git](docs/Troubleshooting.md#prepareing-resources-for-offline-installation-without-git)
--->



