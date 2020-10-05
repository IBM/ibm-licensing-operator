# Backup and upgrade

* [Backup](#backup)
* [Upgrades](#upgrades)

## Backup

The license usage data that is collected by License Service is stored in the persistent cluster memory and is not affected when you kill or restart a pod.

Nonetheless, it is a good practice to generate an audit snapshot periodically for backup purposes and store it in a safe location. You do not need to perform any other backup.

**Note:** Before decommissioning a cluster, record the license usage of the products that are deployed on this cluster by generating an audit snapshot until the day of decommissioning.

## Upgrades

* For online environments, License Service is automatically upgraded with each new operator release.
* For offline environments, to upgrade License Service to a new version, first uninstall License Service from the cluster and redeploy it.
    
    **Note:** The license usage data is stored in the persistent cluster memory and should not be affected by reinstallation of License Service. However, it is a good practice to create an audit snapshot before reinstalling License Service as a safety precaution. 

**Related links**

* [Go back to home page](../License_Service_main.md#documentation)
* [Retrieving license usage data from the cluster](Retrieving_data.md)
* [Uninstalling](Uninstalling.md)
* [Offline installation](Install_offline.md)
