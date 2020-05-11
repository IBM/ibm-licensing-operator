# ibm-licensing-operator

> **Important:** <ul><li>If you install this operator as a part of the IBM Cloud Platform Common Services, do not install this operator directly. Install this operator using the IBM Common Services Operator. For more information about installing this operator and other Common Services operators, see [Installer documentation](http://ibm.biz/cpcs_opinstall).</li><li>If you are using this operator as part of an IBM Cloud Pak, see the documentation for that IBM Cloud Pak to learn more about how to install and use the operator service. For more information about IBM Cloud Paks, see [IBM Cloud Paks that use Common Services](http://ibm.biz/cpcs_cloudpaks).</li><li>If you are using this operator outside of IBM Cloud Platform Common Services or an IBM Cloud Pak, see [Installing IBM Licensing Operator with stand-alone containerized IBM products](#installing-ibm-licensing-operator-with-stand-alone-containerized-ibm-products-using-operator-lifecycle-managerolm).</li></ul>

**ibm-licensing operator** installs **IBM License Service**.

**IBM License Service** is a tool that collects information about license usage of IBM products, such as IBM Cloud Paks and stand-alone containerized IBM software, on a cluster where it is deployed.
Using an API call, you can retrieve the license usage of your products and generate an audit snapshot on demand.

**IBM License Service**, as a part of IBM Cloud Platform Common Services, is integrated in IBM Cloud Paks. The operator, however, allows you to install the service with Operator Lifecycle Manager (OLM) separately for stand-alone containerized IBM software.

For more information about the available IBM Cloud Platform Common Services, see the [IBM Knowledge Center](http://ibm.biz/cpcsdocs).

## Supported platforms
```
<!--
List the platforms and operating systems on which the operator is supported. If needed, such as when complex mapping of OpenShift versions to architecture platforms exists, use a table to identify the supported platforms.

Example supported platforms:

Red Hat OpenShift Container Platform 4.2 or newer installed on one of the following platforms:

   - Linux x86_64
   - Linux on Power (ppc64le)
   - Linux on IBM Z and LinuxONE
-->
```

- [OpenShift Container Platform](https://www.openshift.com/) 4.2 or higher
- Kubernetes 1.11.3 or higher

## Operator versions

-1.0.0
    - First release
 
 ## Prerequisites

Before you install this operator, you need to first install the operator dependencies and prerequisites:

- For the list of operator dependencies, see the IBM Knowledge Center [Common Services dependencies documentation](http://ibm.biz/cpcs_opdependencies).

- For the list of prerequisites for installing the operator, see the IBM Knowledge Center [Preparing to install services documentation](http://ibm.biz/cpcs_opinstprereq).

## Documentation

To install the operator with the IBM Common Services Operator follow the the installation and configuration instructions within the IBM Knowledge Center.

- If you are using the operator as part of an IBM Cloud Pak, see the documentation for that IBM Cloud Pak. For a list of IBM Cloud Paks, see [IBM Cloud Paks that use Common Services](http://ibm.biz/cpcs_cloudpaks).
- If you are using the operator with an IBM Containerized Software on OpenShift, see the IBM Cloud Platform Common Services Knowledge Center [Installer documentation](http://ibm.biz/cpcs_opinstall).
- If you are using the operator an IBM Containerized Software outside of OpenShoift, see the [Developer guide](#devguide)

## SecurityContextConstraints Requirements

License Service supports running with the OpenShift Container Platform 4.3 default restricted Security Context Constraints (SCCs).

For more information about the OpenShift Container Platform Security Context Constraints, see [Managing Security Context Constraints](https://docs.openshift.com/container-platform/4.3/authentication/managing-security-context-constraints.html).

## Developer guide
{ :#devguide}

If, as a developer, you are looking to build and test this operator to try out and learn more about the operator and its capabilities, you can use the following developer guide. This guide provides commands for a quick install and initial validation for running the operator.

> **Important:** The following developer guide is provided as-is and only for trial and education purposes. IBM and IBM Support does not provide any support for the usage of the operator with this developer guide. For the official supported install and usage guide for the operator, see the the IBM Knowledge Center documentation for your IBM Cloud Pak or for IBM Cloud Platform Common Services.

### Quick start guide

Use the following quick start instructions for building and testing the operator.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Installing IBM Licensing Operator as a part of IBM Cloud Platform Common Services on OpenShift](#installing-ibm-licensing-operator-as-a-part-of-ibm-cloud-platform-common-services-on-openshift)
- [Installing IBM Licensing Operator with stand-alone containerized IBM products using Operator Lifecycle Manager(OLM)](#installing-ibm-licensing-operator-with-stand-alone-containerized-ibm-products-using-operator-lifecycle-managerolm)
    - [Installing the IBM Licensing Operator on OCP 4.2+](#installing-the-ibm-licensing-operator-on-ocp-42)
    - [Install the IBM Licensing Operator on Kubernetes from scratch with `kubectl`](#install-the-ibm-licensing-operator-on-kubernetes-from-scratch-with-kubectl)
    - [Offline installation](#offline-installation)
- [Post-installation steps](#post-installation-steps)
    - [Create instance on OpenShift Console 4.2+](#create-instance-on-openshift-console-42)
    - [Creating an instance from console](#creating-an-instance-from-console)
    - [Check Components](#check-components)
- [Using IBM License Service to retrieve license usage information](#using-ibm-license-service-to-retrieve-license-usage-information)
- [Uninstalling License Service from a Kubernetes cluster](#uninstalling-license-service-from-a-kubernetes-cluster)
- [Troubleshooting](#troubleshooting)
    - [CreateContainerConfigError Marketplace Operator error](#createcontainerconfigerror-marketplace-operator-error)
    - [Prepare resources for offline installation without git](#prepare-resources-for-offline-installation-without-git)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

#### Installing IBM Licensing Operator as a part of IBM Cloud Platform Common Services on OpenShift

<b>Installation</b>

For the installation steps, see [Installing IBM Cloud Platform Services in your OpenShift Container Platform cluster](https://www.ibm.com/support/knowledgecenter/SSHKN6/installer/1.1.0/install_operator.html).

<!-- Section below is mentioned by READINESS WIKI that is used by other IBM teams, so after changing this section name, readiness wiki also needs to be changed -->

#### Automatically installing ibm-licensing operator with a stand-alone IBM Containerized Software using Operator Lifecycle Manager(OLM)

There is automatic script that installs IBM Licensing Operator, creates instance and validates the steps. It was tested to work on `OpenShift Container Platform 4.2+`, `vanilla Kubernetes` cluster, and is available at:
[common/scripts/ibm_licensing_operator_install.sh](common/scripts/ibm_licensing_operator_install.sh)

If you want to do it by hand, you can also find steps to do it below.

##### Installing the IBM Licensing Operator on OCP 4.2+

<b>Prerequisites</b>
- Administrator permissions for the cluster

1\. **Create OperatorSource**

Before you install IBM Licensing Operator, the following operator source should be created to get operator bundles from `quay.io`.

```yaml
apiVersion: operators.coreos.com/v1
kind: OperatorSource
metadata:
  name: opencloud-operators
  namespace: openshift-marketplace
spec:
  authorizationToken: {}
  displayName: IBMCS Operators
  endpoint: https://quay.io/cnr
  publisher: IBM
  registryNamespace: opencloudio
  type: appregistry
```

To add the OperatorSource:

a. Log in to OpenShift Console

b. Click the plus button on the right hand site of the header

c. Copy the above operator source into the editor.

2\. **Create the `ibm-common-services` namespace**

a. From the hamburger menu in the OpenShift Console, go to **Operators>Operator Hub**.

b. Select **Create Project** and type **ibm-common-services** as a name.

c. Click **Create**

![Create Project](images/create-project.png)

3\. **Install IBM Licensing Operator package in OperatorHub**

a. Go to **OperatorHub** and search for **IBM Licensing Operator**.

b. Select **IBM Licensing Operator** and click **Install**.

![Operator Hub IBM Licensing](images/operator-hub-licensing.png)

4\. As **A specific namespace on the cluster** select **ibm-common-services** that you created in the previous step, and click **Subscribe**.

![Subscribe to IBM Licensing OLM](images/subscribe-licensing.png)

5\. To check if the installation is successful, wait for about 1 minute, and click **Installed operators**. You should see IBM Licensing Operator in the **InstallSucceeded** status.

![IBM Licensing Installed](images/installed.png)

##### Install the IBM Licensing Operator on Kubernetes from scratch with `kubectl`

**Prerequisites**
- Administrator permissions for the cluster
- 'kubectl` 1.11.3 or higher
- Linux or iOS

    **Note**: To install the IBM Licensing Operator on Windows, adjust the commands to fit the Windows standard.

1\. **Install the Operator Lifecycle Manager (OLM)**

a. Make sure that you are connected to your cluster. You can, for example, run the following command:

```bash
kubectl get node
```

The response includes a list of your nodes

b. Download OLM release from [the OLM GitHub repository](https://github.com/operator-framework/operator-lifecycle-manager/releases).

   **Note:** For versions newer than 13.0, the process might differ.

c. Use the following script to install OLM v13.0:

```bash
curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.13.0/install.sh | bash -s 0.13.0
```

   **Troubleshooting:** If you get an error, you might have the old version of Kubernetes. You can try either upgrading your Kubernetes server version or using the older version of OLM.

2\. **Install the Operator Marketplace**

1\) Clone the Operator Marketplace GitHub repo with the following command:

```bash
git clone --single-branch --branch release-4.6 https://github.com/operator-framework/operator-marketplace.git
```

2\) Change the `marketplace` namespace to `olm` to be able to create subscriptions to the operatorsource/catalogsource from a different namespace.

3\) Check your `global catalog namespace` at OLM `packageserver` pod yaml somewhere in your cluster, using `grep           global-namespace`.
If the cluster's `global catalog namespace` is different than `olm`, complete the following steps to change it:

a. Run the following command:

- For **Linux** users

```bash
# change this value if this would not be olm
export GLOBAL_CATALOG_NAMESPACE=olm
# change all resources namespace to olm
sed -i 's/namespace: .*/namespace: '"${GLOBAL_CATALOG_NAMESPACE}"'/g' operator-marketplace/deploy/upstream/*
# change namespace to olm
sed -i 's/name: .*/name: '"${GLOBAL_CATALOG_NAMESPACE}"'/g' operator-marketplace/deploy/upstream/01_namespace.yaml
```

- For **MAC** users:

```bash
export GLOBAL_CATALOG_NAMESPACE=olm
sed -i "" 's/namespace: .*/namespace: '"${GLOBAL_CATALOG_NAMESPACE}"'/g' operator-marketplace/deploy/upstream/*
sed -i "" 's/name: .*/name: '"${GLOBAL_CATALOG_NAMESPACE}"'/g' operator-marketplace/deploy/upstream/01_namespace.yaml
```

b. Install the Operator Marketplace into the cluster in the `$GLOBAL_CATALOG_NAMESPACE` namespace:

```bash
kubectl apply -f operator-marketplace/deploy/upstream
```

c. **Optional**: If you get the `unknown field "preserveUnknownFields"` error, try to delete `preserveUnknownFields` from yaml files inside `operator-marketplace/deploy/upstream/` catalog or consider upgrading Kubernetes server version by running the following command:

- For **Linux** users:

```bash
sed -i '/.*preserveUnknownFields.*/d' operator-marketplace/deploy/upstream/*
kubectl apply -f operator-marketplace/deploy/upstream
```

- For **MAC** users:

```bash
sed -i "" '/.*preserveUnknownFields.*/d' operator-marketplace/deploy/upstream/*
kubectl apply -f operator-marketplace/deploy/upstream
```

3\. **Create the OperatorSource**

An `OperatorSource` object is used to define the external datastore that is used to store operator bundles. For more information including examples, see the documentation in the `operator-marketplace` [repository](https://github.com/operator-framework/operator-marketplace#operatorsource).

<b>Create `operator source` to get operator bundles from `quay.io`.</b>

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: operators.coreos.com/v1
kind: OperatorSource
metadata:
  name: opencloud-operators
  namespace: $GLOBAL_CATALOG_NAMESPACE
spec:
  authorizationToken: {}
  displayName: IBMCS Operators
  endpoint: https://quay.io/cnr
  publisher: IBM
  registryNamespace: opencloudio
  type: appregistry
EOF
```

<b>Check results:</b>

- The `operator-marketplace` controller should successfully process this object. See if the Status is `Succeeded`:

```console
$ kubectl get operatorsource opencloud-operators -n $GLOBAL_CATALOG_NAMESPACE
NAME                  TYPE          ENDPOINT              REGISTRY      DISPLAYNAME       PUBLISHER   STATUS      MESSAGE                                       AGE
opencloud-operators   appregistry   https://quay.io/cnr   opencloudio   IBMCS Operators   IBM         Succeeded   The object has been successfully reconciled   1m32s
```

- Additionally, a `CatalogSource` is created in the `$GLOBAL_CATALOG_NAMESPACE` namespace:

```console
$ kubectl get catalogsource -n $GLOBAL_CATALOG_NAMESPACE
NAME                           DISPLAY                        TYPE   PUBLISHER   AGE
opencloud-operators            IBMCS Operators                grpc   IBM         20m
[...]
```

- If everything goes well, you should see the following pods:

```console
$ kubectl get pod -n $GLOBAL_CATALOG_NAMESPACE
NAME                                            READY   STATUS    RESTARTS   AGE
marketplace-operator-6576b4ddc8-dqcgr           1/1     Running   0          84s
opencloud-operators-66df4d97ff-4rhjj            1/1     Running   0          80s
upstream-community-operators-7ffb6b674b-7qlvx   1/1     Running   0          80s
[...]
```

   **Troubleshooting:** In case of any problems, check the [troubleshooting section](#createcontainerconfigerror-marketplace-operator-error).

4\. **View Available Operators**

Once the `OperatorSource` and `CatalogSource` are deployed, the following command can be used to list the available operators including ibm-licensing-operator-app.
**Note:** The command assumes that the of the `OperatorSource` object is `opencloud-operators`. Adjust if needed.

```console
$ kubectl get opsrc opencloud-operators -o=custom-columns=NAME:.metadata.name,PACKAGES:.status.packages -n $GLOBAL_CATALOG_NAMESPACE
NAME                  PACKAGES
opencloud-operators   ibm-meta-operator-bridge-app,ibm-commonui-operator-app,ibm-catalog-ui-operator-app,ibm-metering-operator-app,ibm-helm-repo-operator-app,ibm-iam-operator-app,ibm-elastic-stack-operator-app,ibm-monitoring-exporters-operator-app,ibm-monitoring-prometheusext-operator-app,cp4foobar-operator-app,ibm-healthcheck-operator-app,ibm-platform-api-operator-app,ibm-management-ingress-operator-app,ibm-helm-api-operator-app,ibm-licensing-operator-app,ibm-ingress-nginx-operator-app,ibm-monitoring-grafana-operator-app,ibm-auditlogging-operator-app,operand-deployment-lifecycle-manager-app,ibm-mgmt-repo-operator-app,ibm-mongodb-operator-app,ibm-cert-manager-operator-app
```

5\. **Create an OperatorGroup**

An `OperatorGroup` is used to denote which namespaces your Operator should watch.
It must exist in the namespace where your operator is deployed, for example, `ibm-common-services`.

1\) Create a namespace for IBM Licensing Operator with the following command:

```bash
kubectl create namespace ibm-common-services
```

2\) Use the following command to deploy the `OperatorGroup` resource:

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: operatorgroup
  namespace: ibm-common-services
spec:
  targetNamespaces:
  - ibm-common-services
EOF
```

6\. **Create a Subscription**

 The last piece that ties together all of the previous steps is creating a subscription for the Operator. A subscription is created for the operator that upgrades IBM Licensing Operator when needed.

 Create the **Subscription** using the following command:

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: ibm-licensing-operator-app
  namespace: ibm-common-services
spec:
  channel: stable-v1
  name: ibm-licensing-operator-app
  source: opencloud-operators
  sourceNamespace: $GLOBAL_CATALOG_NAMESPACE
EOF
```

7\. **Verify Operator health**

1\) See if the IBM Licensing Operator is deployed by OLM from the `catalog source` created by `Operator Marketplace` with the following command:

```console
$ kubectl get clusterserviceversion -n ibm-common-services
NAME                            DISPLAY                  VERSION   REPLACES                        PHASE
ibm-licensing-operator.v1.0.0   IBM Licensing Operator   1.0.0     ibm-licensing-operator.v0.0.0   Succeeded
```

**Note:** The above command assumes that you have created the Subscription in the `ibm-common-services` namespace.
If your Operator deployment (CSV) shows `Succeeded` in the `InstallPhase` status, your Operator is deployed successfully. Otherwise, check the `ClusterServiceVersion` objects status for details.

2\) **Optional**: You can also check your Operator's deployment:

```bash
kubectl get deployment -n ibm-common-services | grep ibm-licensing-operator
```

##### Offline installation

<b>Prerequisites</b>

You need to have private docker image registry, where you can push images using `docker` and from where your cluster can pull images. Also you need machine with access to your cluster with `kubectl` command.

1\. **Prepare docker images**

First you need to push images to your registry:

```bash
# on machine with access to internet
export my_docker_registry=<YOUR REGISTRY IMAGE PREFIX HERE f.e.: "my.registry:5000" or "quay.io/opencloudio">

# pull needed images
docker pull quay.io/opencloudio/ibm-licensing-operator:1.1.0
docker pull quay.io/opencloudio/ibm-licensing:1.1.0

# tag them with your registry prefix and push
docker tag quay.io/opencloudio/ibm-licensing-operator:1.1.0 ${my_docker_registry}/ibm-licensing-operator:1.1.0
docker push ${my_docker_registry}/ibm-licensing-operator:1.1.0
docker tag quay.io/opencloudio/ibm-licensing:1.1.0 ${my_docker_registry}/ibm-licensing:1.1.0
docker push ${my_docker_registry}/ibm-licensing:1.1.0
```

2\. **Create needed resources**

Now do everything on machine where you have access to your cluster and can use `kubectl` to access it.

```bash
# on machine with access to cluster
export my_docker_registry=<SAME REGISTRY AS BEFORE>
```

We will install our operator in `ibm-common-services` namespace. So first we need to create it:

```bash
kubectl create namespace ibm-common-services
```

If your cluster need access token to your private docker registry then create this secret in created namespace:

```bash
kubectl create secret -n ibm-common-services docker-registry my-registry-token --docker-server=${my_docker_registry} --docker-username=<YOUR_REGISTRY_USERNAME> --docker-password=<YOUR_REGISTRY_TOKEN> --docker-email=<YOUR_REGISTRY_EMAIL, probably can be same as username>
```

Set context so that resources will be made `ibm-common-services` namespace:

```bash
kubectl config set-context --current --namespace=ibm-common-services
```

If You cannot use `git clone` on machine with `kubectl` (f.e. when You don't have internet connection there) then use the solution at troubleshooting section at the end of the readme named
['Prepare resources for offline installation without git'](#prepare-resources-for-offline-installation-without-git). Then continue to step 3 below.

If You can use `git clone`:

```bash
git clone -b v1.0.0-cambridge https://github.com/IBM/ibm-licensing-operator.git
cd ibm-licensing-operator/
```

Apply RBAC roles and CRD:

```bash
# add CRD:
kubectl apply -f deploy/crds/operator.ibm.com_ibmlicensings_crd.yaml
# add RBAC:
kubectl apply -f deploy/role.yaml
kubectl apply -f deploy/service_account.yaml
kubectl apply -f deploy/role_binding.yaml
```

Modify operator.yaml image so your private registry is used:

- For **LINUX** users:

```bash
ESCAPED_REPLACE=$(echo ${my_docker_registry} | sed -e 's/[\/&]/\\&/g')
sed -i 's/quay\.io\/opencloudio/'"${ESCAPED_REPLACE}"'/g' deploy/operator.yaml
kubectl apply -f deploy/operator.yaml
```

- For **MAC** users:

```bash
ESCAPED_REPLACE=$(echo ${my_docker_registry} | sed -e 's/[\/&]/\\&/g')
sed -i "" 's/quay.io\/opencloudio/'"${ESCAPED_REPLACE}"'/g' deploy/operator.yaml
kubectl apply -f deploy/operator.yaml
```

3\. **Steps to include when creating operator instance**

Create the instance just like in [online installation](#creating-an-instance-from-console), but if you created secret needed for access to images then also add it, like that:

```yaml
apiVersion: operator.ibm.com/v1alpha1
kind: IBMLicensing
metadata:
  name: instance
spec:
...
  imagePullSecrets:
    - my-registry-token
...
```

Example:

```yaml
apiVersion: operator.ibm.com/v1alpha1
kind: IBMLicensing
metadata:
  name: instance
spec:
  apiSecretToken: ibm-licensing-token
  datasource: datacollector
  httpsEnable: false
  instanceNamespace: ibm-common-services
  imagePullSecrets:
    - my-registry-token
```

#### Post-installation steps

After you successfully install IBM Licensing Operator, you can create IBMLicensing instance that will make IBM License Service run on cluster.

##### Create instance on OpenShift Console 4.2+

If you have OpenShift 4.2+ you can create the instance from the Console.

1\. Go to **Installed Operators>IBM Licensing Operator>IBM Licensing tab>Create IBMLicensing**

![OCP click Create IBM Licensing](images/ocp_create_instance.png)

2\. Click **Create IBMLicensing** to edit your parameters.

   **Note:** Make sure to change datasource to `datacollector`. For more information about the parameters, see [IBMLicensingOperatorParameters](images/IBMLicensingOperatorParameters.csv).

![OCP instance datacollector](images/ocp_instance_datacollector.png)

3\. Click **Create**.

   **Note:** To edit your instance in the future, in OpenShift Console go to **Administration>Custom Resource Definitions>select IBMLicensing>instances>Edit IBMLicensing**

![OCP Edit Instance](images/ocp_edit_instance.png)

**Troubleshooting**: If the instance is not updated properly, try deleting the instance and creating new one with new parameters.

##### Creating an instance from console

Minimal setup requires applying this IBMLicensing instance:

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: operator.ibm.com/v1alpha1
kind: IBMLicensing
metadata:
  name: instance
spec:
  apiSecretToken: ibm-licensing-token
  datasource: datacollector
  httpsEnable: true
  instanceNamespace: ibm-common-services
EOF
```

<b>Configuring ingress</b>

You might want to configure ingress. Here is an example of how you can do it:

1\. Get the nginx ingress controller You might get it, for example, from here: [https://kubernetes.github.io/ingress-nginx/deploy](https://kubernetes.github.io/ingress-nginx/deploy)

2\. Apply this IBMLicensing instance to your cluster:

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: operator.ibm.com/v1alpha1
kind: IBMLicensing
metadata:
  name: instance
spec:
  apiSecretToken: ibm-licensing-token
  datasource: datacollector
  httpsEnable: false
  instanceNamespace: ibm-common-services
  ingressEnabled: true
  ingressOptions:
    annotations:
      "nginx.ingress.kubernetes.io/rewrite-target": "/\$2"
    path: /ibm-licensing-service-instance(/|$)(.*)
EOF
```

3\. Access the instance at your ingress host with the following path: `/ibm-licensing-service-instance`.

**Note:** For HTTPS, set `spec.httpsEnable` to `true`, and edit `ingressOptions`. Read more about the options here:
[IBMLicensingOperatorParameters](images/IBMLicensingOperatorParameters.csv)

**Troubleshooting**: If the instance is not updated properly (for example after updating ingressOptions), try deleting the instance and creating new one with new parameters.

##### Check Components

1\. Check whether the pod is created.
To see the logs go to **OCP UI->Workloads->Pods** and search for **licensing** in the `ibm-common-services` project:

![OCP Pod](images/ocp_pod.png)

2\. Check if the pod is running. To investigate further select `licensing` and check logs, and events.
You can also run the following command from the console:

```bash
podName=`kubectl get pod -n ibm-common-services -o jsonpath="{range .items[*]}{.metadata.name}{'\n'}" | grep ibm-licensing-service-instance`
kubectl logs $podName -n ibm-common-services
kubectl describe pod $podName -n ibm-common-services
```

3\. Check Route or Ingress settings depending on your parameter settings.

You can check the Route or Ingress Settings in **OCP UI->Networking->Service**. Or you can check if using the console command, for example:

```bash
kubectl get ingress -n ibm-common-services -o yaml
```

#### Using IBM License Service to retrieve license usage information

For more information about how to use License Service to retrieve license usage data, se [IBM Cloud Platform Common Services documentation](https://www.ibm.com/support/knowledgecenter/SSHKN6/license-service/1.0.0/retrieving.html).

#### Uninstalling License Service from a Kubernetes cluster

**Note:** The following procedure assumes that you have deployed IBM License Service in the `ibm-common-services` namespace

1\. **Delete the `IBMLicensing custom` resource**

Delete the instance and the operator will clean its resources.
First, check what `ibmlicensing` instances you have by running the following command:

```bash
licensingNamespace=ibm-common-services
kubectl get ibmlicensing -n ${licensingNamespace} -o jsonpath="{range .items[*]}{.metadata.name}{'\n'}"
```

The command should return one instance. Delete this instance, if it exists with the following command:

```bash
licensingNamespace=ibm-common-services
instanceName=`kubectl get ibmlicensing -n ${licensingNamespace} -o jsonpath="{range .items[*]}{.metadata.name}{'\n'}"`
kubectl delete ibmlicensing ${instanceName} -n ${licensingNamespace}
```

2\. **Delete the operator subscription**

Run the following command to see your subscriptions:

```bash
licensingNamespace=ibm-common-services
kubectl get subscription -n ${licensingNamespace} -o jsonpath="{range .items[*]}{.metadata.name}{'\n'}"
```

Delete the `ibm-licensing-operator-app` subscription by using the following command:

```bash
licensingNamespace=ibm-common-services
subName=ibm-licensing-operator-app
kubectl delete subscription ${subName} -n ${licensingNamespace}
```

3\. **Delete Cluster Service Version (CSV)**

Delete CSV that manages the Operator image.
Run the following command to get your CSV name, look for `ibm-licensing-operator`:

```bash
licensingNamespace=ibm-common-services
kubectl get clusterserviceversion -n ${licensingNamespace} -o jsonpath="{range .items[*]}{.metadata.name}{'\n'}"
kubectl get clusterserviceversion -n ${licensingNamespace} -o jsonpath="{range .items[*]}{.metadata.name}{'\n'}" | grep ibm-licensing-operator
```

Delete it by using the following command:

```bash
licensingNamespace=ibm-common-services
csvName=`kubectl get clusterserviceversion -n ${licensingNamespace} -o jsonpath="{range .items[*]}{.metadata.name}{'\n'}" | grep ibm-licensing-operator`
kubectl delete clusterserviceversion ${csvName} -n ${licensingNamespace}
```

4\. **Delete Custom Resource Definition (CRD)**

Delete the custom resource definition with the following command:

```bash
kubectl delete CustomResourceDefinition ibmlicensings.operator.ibm.com
```

5\. **Delete Operator Group**

**Note:** If you have other subscriptions that are tied with that operatorGroup do not delete it.
IBM Licensing Operator is now uninstalled.You can also clean up the operatorgroup that you created for subscription by using the following command:

```bash
licensingNamespace=ibm-common-services
operatorGroupName=operatorgroup
kubectl delete OperatorGroup ${operatorGroupName} -n ${licensingNamespace}
```

6\. **Delete OperatorSource**

**Note:** If you have other services that use the opencloudio catalog source do not delete the OperatorSource.
Otherwise, you can delete the OperatorSource with the following command:

```bash
export GLOBAL_CATALOG_NAMESPACE=olm
opencloudioSourceName=opencloud-operators
kubectl delete OperatorSource ${opencloudioSourceName} -n ${GLOBAL_CATALOG_NAMESPACE}
```

7\. **Delete OperatorMarketplace**

**Important:** Do not delete the OperatorMarketplace if it use it with a different operator.

You can delete the OperatorMarketplace with the following command:

```bash
export GLOBAL_CATALOG_NAMESPACE=olm
kubectl delete -f operator-marketplace/deploy/upstream
```

8\. **Uninstall OLM**

**Important:** Before uninstalling OLM, make sure that you do not use it with other operators.

Uninstall OLM with the following command:

```bash
export GLOBAL_CATALOG_NAMESPACE=olm
kubectl delete crd clusterserviceversions.operators.coreos.com \
installplans.operators.coreos.com \
subscriptions.operators.coreos.com \
catalogsources.operators.coreos.com \
operatorgroups.operators.coreos.com
kubectl delete namespace ${GLOBAL_CATALOG_NAMESPACE}
```

### Troubleshooting

#### CreateContainerConfigError Marketplace Operator error

In case of problems during the installation of operator-marketplace, check the pods in the marketplace and their status.

If you see the following error, note down the pod name:

```console
$ kubectl get pod -n $GLOBAL_CATALOG_NAMESPACE
NAME                                    READY   STATUS                       RESTARTS   AGE
marketplace-operator-7d4c5bdb5-mxsj6    0/1     CreateContainerConfigError   0          1m36s
```

Then, check what the problem is by using the yaml where you provide a pod name:

```bash
kubectl get pod marketplace-operator-7d4c5bdb5-mxsj6 -n $GLOBAL_CATALOG_NAMESPACE -o yaml
```

In case the following error appears in the pod status: `container has runAsNonRoot and image has non-numeric user (marketplace-operator), cannot verify user is non-root`, fix it by adding securityContext to operator-marketplace/deploy/upstream:

```console
vim operator-marketplace/deploy/upstream/08_operator.yaml
```

Next, append the following lines:

```yaml
...
      containers:
        - name: marketplace-operator
          securityContext: # <- this
            runAsUser: 65534 # <- and this
          image: quay.io/openshift/origin-operator-marketplace:latest
...
```

And apply the configuration:

```bash
kubectl apply -f operator-marketplace/deploy/upstream/08_operator.yaml
```

#### Prepare resources for offline installation without git

Apply RBAC roles and CRD:

```bash
# copy the yaml from here:
https://github.com/IBM/ibm-licensing-operator/releases/download/v1.0.0-cambridge/rbac_and_crd.yaml
```

Then apply the copied yaml:

```bash
cat <<EOF | kubectl apply -f -
# PASTE IT HERE
EOF
```

Make sure `${my_docker_registry}` variable has your private registry and apply the operator:

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ibm-licensing-operator
  labels:
    app.kubernetes.io/instance: "ibm-licensing-operator"
    app.kubernetes.io/managed-by: "ibm-licensing-operator"
    app.kubernetes.io/name: "ibm-licensing"
spec:
  replicas: 1
  selector:
    matchLabels:
      name: ibm-licensing-operator
  template:
    metadata:
      labels:
        name: ibm-licensing-operator
        app.kubernetes.io/instance: "ibm-licensing-operator"
        app.kubernetes.io/managed-by: "ibm-licensing-operator"
        app.kubernetes.io/name: "ibm-licensing"
      annotations:
        productName: IBM Cloud Platform Common Services
        productID: "068a62892a1e4db39641342e592daa25"
        productVersion: "3.3.0"
        productMetric: FREE
    spec:
      serviceAccountName: ibm-licensing-operator
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
              - matchExpressions:
                  - key: beta.kubernetes.io/arch
                    operator: In
                    values:
                      - amd64
      hostIPC: false
      hostNetwork: false
      hostPID: false
      containers:
        - name: ibm-licensing-operator
          image: ${my_docker_registry}/ibm-licensing-operator:1.1.0
          command:
            - ibm-licensing-operator
          imagePullPolicy: Always
          env:
            - name: WATCH_NAMESPACE
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: OPERATOR_NAME
              value: "ibm-licensing-operator"
            - name: OPERAND_LICENSING_IMAGE
              value: "${my_docker_registry}/ibm-licensing:1.1.0"
            - name: SA_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.serviceAccountName
          resources:
            limits:
              cpu: 20m
              memory: 200Mi
            requests:
              cpu: 10m
              memory: 100Mi
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
            privileged: false
            readOnlyRootFilesystem: true
            runAsNonRoot: true
EOF
```

### Debugging guide

Use the following commands to debug the operator:

```
<!--- Include any required quick start commands -->
```

### End-to-End testing

For more instructions on how to run end-to-end testing with the Operand Deployment Lifecycle Manager, see [ODLM guide](https://github.com/IBM/operand-deployment-lifecycle-manager/blob/master/docs/install/common-service-integration.md#end-to-end-test).
