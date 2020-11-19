
# Automatic installation using Operator Lifecycle Manager (OLM)

You can use an automatic script to install License Service on the cluster, and automatically create the License Service instance.

- [Supported configurations](#supported-configurations)
- [Installation](#installation)

## Supported configurations

The script is supported on the following platforms:

- Linux x86 architecture with or without Operator Lifecycle Manager (OLM)
- Any other cluster that already has Operator Lifecycle Manager (OLM) installed

## Prerequisites

### IBM Cloud Kubernetes Services (IKS)

To install License Service on IBM Cloud Kubernetes Services (IKS), make sure that OLM is installed, or that the script can installed OLM for you. Make sure that you meet one of the following criteria.

- OLM is properly installed.
- OLM is not installed. Check whether the OLM Custom Resource Definition (CRD) exists on the cluster.
  - The OLM CRD does not exist. Proceed with the installation of License Service. The script installs OLM for you.
  - The OLM CRD exists but OLM is not installed. Remove the CRD. Proceed with the installation of License Service. The script installs OLM for you.
  - The OLM CRD exists and OLM is not installed, however, you cannot remove the CRD because you use it for other purposes. Install OLM on top of your existing CRD and proceed with installation of License Service. In this scenario, the scrip cannot install OLM for you.

- OLM is properly installed.
- OLM is not installed. Check whether the OLM Custom Resource Definition (CRD) exists on the cluster.

|Scenario|Actions|
|---|---|
|The OLM CRD does not exist.|Proceed with the installation of License Service. The script installs OLM for you.
|The OLM CRD exists but OLM is not installed.|<ul><li>Remove the CRD. Proceed with the installation of License Service. The script installs OLM for you.</li><li>If you cannot remove the CRD because you use it for other purposes, install OLM on top of your existing CRD and proceed with installation of License Service. In this scenario, the scrip cannot install OLM for you.
  - The OLM CRD does not exist. Proceed with the installation of License Service. The script installs OLM for you.

## Installation

The script installs License Service, creates an instance and validates the installation steps.

1\. Download the script from the following location in the repository:
[common/scripts/ibm_licensing_operator_install.sh](/common/scripts/ibm_licensing_operator_install.sh).

2\. Run the script.

**Results:**
Installation is complete and **License Service** is running in your cluster. To check if License Service components are properly installed, and perform extra configuration, see [Configuration](Configuration.md).

<b>Related links</b>

- [Go back to home page](../License_Service_main.md#documentation)
- [Configuration](Configuration.md)
- [Retrieving license usage data from the cluster](Retrieving_data.md)
