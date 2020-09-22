
# Automatic installation using Operator Lifecycle Manager (OLM)

You can use an automatic script to install License Service on the cluster, and automatically create the License Service instance.

## Supported configurations

The script is supported on the following platforms: 
- Linux x86 architecture,
- Linux on Power (ppc64le), Linux on IBM Z and LinuxONE on Red Hat OpenShift Container Platform 3.11, 4.1, 4.2, 4.3 or newer, or on any other cluster that already has Operator Lifecycle Manager (OLM) installed.
The script was tested on `OpenShift Container Platform 4.2+`, `ICP cluster: v1.12.4+icp-ee`, `vanilla Kubernetes custer`

## Installation

The script installs License Service, creates an instance and validates the installation steps. 

1. Download the script from the following location in the repository:
[common/scripts/ibm_licensing_operator_install.sh](../common/scripts/ibm_licensing_operator_install.sh). 

2. Run the script.

**Results:** 
Installation is complete and **License Service** is running in your cluster.

**Related links**

- [Go back to home page](../License_Service_main.md#documentation)
- [Configuration](Configuration.md)
- [Retrieving license usage data from the cluster](Retrieving_data.md)
