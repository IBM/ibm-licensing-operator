
# Automatic installation using Operator Lifecycle Manager (OLM)

You can use an automatic script to install License Service on the cluster, and automatically create the License Service instance.

- [Supported configurations](#supported-configurations)
- [Installation](#installation)

## Supported configurations

The script is supported on the following platforms:

- Linux x86 architecture with or without Operator Lifecycle Manager (OLM)
- Any other cluster that already has Operator Lifecycle Manager (OLM) installed

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
