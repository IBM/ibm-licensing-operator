# IBM Licensing Operator

**IBM Licensing Operator** installs **IBM License Service**.

**IBM License Service** is a tool that collects information about license usage of IBM products, such as IBM Cloud Paks and stand-alone containerized IBM software, on a cluster where it is deployed. 
Using an API call, you can retrieve the license usage of your products and generate an audit snapshot on demand.

**IBM License Service**, as a part of IBM Cloud Platform Common Services, is integrated in IBM Cloud Paks. The Operator, however, allows you to install the service with Operator Lifecycle Manager (OLM0 separately for stand-alone containerized IBM software.

## Operator versions and supported platforms

List the platforms and operation systems on which the operator is supported.

|Operator version|Release Date|Supported operating systems|Supported platforms|Details|
|---|---|---|---|---|
|1.0.0| 03/2020|AMD64|<ul><li>[OpenShift Container Platform](https://www.openshift.com/) 4.2 or higher</li><li>Kubernetes 1.11.3 or higher</li></ul>|First release |

## Documentation

To learn more about License Service, see the [IBM Cloud Platform Common Services documentation](http://ibm.biz/cpcsdocs).

## Developer guide

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Installing IBM Licensing Operator as a part of IBM Cloud Platform Common Services on OpenShift](#installing-ibm-licensing-operator-as-a-part-of-ibm-cloud-platform-common-services-on-openshift)
- [Installing IBM Licensing Operator with stand-alone containerized IBM products using `Operator Lifecycle Manager`(OLM)](#installing-ibm-licensing-operator-with-stand-alone-containerized-ibm-products-using-operator-lifecycle-managerolm)
    - [Install the IBM Licensing Operator On OCP 4.2+](#install-the-ibm-licensing-operator-on-ocp-42)
    - [Install the IBM Licensing Operator on Kubernetes from scratch with `kubectl`](#install-the-ibm-licensing-operator-on-kubernetes-from-scratch-with-kubectl)
- [Post-installation steps](#post-installation-steps)
    - [Create instance on OCP 4.2+](#create-instance-on-ocp-42)
    - [Creating an instance from console](#creating-an-instance-from-console)
    - [Check Components](#check-components)
- [Using IBM License Service to retrieve license usage information](#using-ibm-license-service-to-retrieve-license-usage-information)
- [Uninstalling from any kubernetes cluster](#uninstalling-from-any-kubernetes-cluster)
- [Troubleshoot](#troubleshoot)
    - [CreateContainerConfigError Marketplace Operator error](#createcontainerconfigerror-marketplace-operator-error)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

### Installing IBM Licensing Operator as a part of IBM Cloud Platform Common Services on OpenShift

**Prerequisites**

To use License Service, aditionally install the following operators that IBM Licensing Operator depends on:
- ibm-cert-manager-operator
- ibm-licensing-operator
- ibm-metering-operator
- ibm-mongodb-operator

**Installation**

For the installation steps, see [Installing IBM Cloud Platform Services in your OpenShift Container Platform cluster](https://www.ibm.com/support/knowledgecenter/SSHKN6/installer/1.1.0/install_operator.html).

### Installing IBM Licensing Operator with stand-alone containerized IBM products using Operator Lifecycle Manager(OLM)

#### Installing the IBM Licensing Operator on OCP 4.2+

**Prerequisites**
- Administrator permissions for the cluster

1\. **Create `OperatorSource`**

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


To add the OpenSource:
    1. Log in to OpenShift Console
    2. Click the plus button on the right hand site of the header
    3. Copy the above operator source into the editor.

2\. **Create the `ibm-common-services` namespace**

    1. From the hamburger menu in the OpenShift Console, go to **Operators>Operator Hub**.
    2. Select **Create Project** and type **ibm-common-services** as a name.
    3. Click **Create**
    
![Create Project](images/create-project.png)

3\. **Install IBM Licensing Operator package in OperatorHub**

    a. Open `OperatorHub` and search for `IBM Licensing Operator`.

    b. Select `IBM Licensing Operator` and click **Install**.
![Operator Hub IBM Licensing](images/operator-hub-licensing.png)

    c. As **A specific namespace on the cluster** select `ibm-common-services` that you created in the previous step, and click subscribe.
![Subscribe to IBM Licensing OLM](images/subscribe-licensing.png)

Waite for about 1 minute, and click **Installed operators**. You should see IBM Licensing Operator in the `InstallSucceeded` status.
![IBM Licensing Installed](images/installed.png)

#### Install the IBM Licensing Operator on Kubernetes from scratch with `kubectl`

**Prerequisites**
- Administrator permissions for the cluster
- 'kubectl` 1.11.3 or higher
- Linux or iOS
    **Note**: To install on Windows, change the commands to fit the Windows standard.

1\. **Install the Operator Lifecycle Manager**

Make sure that you are connected to your cluster. You can, for example, run the following command:

```bash
kubectl get node
```

You should get a list of your nodes in response.

Download OLM Release from [here](https://github.com/operator-framework/operator-lifecycle-manager/releases)

**Note:** For versions newer than 13.0, the process might differ. Use the following script to install OLM v13.0:

```bash
curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.13.0/install.sh | bash -s 0.13.0
```

If you get an error, you might have the old version of Kubernetes. You can try either upgrading your Kubernetes server version or using older version of OLM.

2\. **Install the Operator Marketplace**

Clone the following github repo:

```bash
git clone --single-branch --branch release-4.6 https://github.com/operator-framework/operator-marketplace.git
```

Change the `marketplace` namespace to `olm` to be able to create subscriptions to the operatorsource/catalogsource from different namespace.

You can check you `global catalog namespace` at OLM `packageserver` pod yaml somewhere in your cluster, using `grep           global-namespace`.
If the cluster's `global catalog namespace` is different than `olm`, complete the following steps to change it:

    1. Run the following command:
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

    2. Install the Operator Marketplace into the cluster in the `$GLOBAL_CATALOG_NAMESPACE` namespace:

```bash
kubectl apply -f operator-marketplace/deploy/upstream
```
    3. Optional: If you get the `unknown field "preserveUnknownFields"` error, try to delete `preserveUnknownFields` from yaml files inside `operator-marketplace/deploy/upstream/` catalog or consider upgrading Kubernetes server version:

For **Linux** users:

```bash
sed -i '/.*preserveUnknownFields.*/d' operator-marketplace/deploy/upstream/*
kubectl apply -f operator-marketplace/deploy/upstream
```

For **MAC** users:

```bash
sed -i "" '/.*preserveUnknownFields.*/d' operator-marketplace/deploy/upstream/*
kubectl apply -f operator-marketplace/deploy/upstream
```

3\. **Create the OperatorSource**

An `OperatorSource` object is used to define the external datastore that is used to store operator bundles. For more information including examples, see the documentation included in the `operator-marketplace` [repository](https://github.com/operator-framework/operator-marketplace#operatorsource).

Create `operator source` to get operator bundles from `quay.io`.

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

The `operator-marketplace` controller should successfully process this object. See if the Status is `Succeeded`:

```console
$ kubectl get operatorsource opencloud-operators -n $GLOBAL_CATALOG_NAMESPACE
NAME                  TYPE          ENDPOINT              REGISTRY      DISPLAYNAME       PUBLISHER   STATUS      MESSAGE                                       AGE
opencloud-operators   appregistry   https://quay.io/cnr   opencloudio   IBMCS Operators   IBM         Succeeded   The object has been successfully reconciled   1m32s
```

Additionally, a `CatalogSource` is created in the `$GLOBAL_CATALOG_NAMESPACE` namespace:

```console
$ kubectl get catalogsource -n $GLOBAL_CATALOG_NAMESPACE
NAME                           DISPLAY                        TYPE   PUBLISHER   AGE
opencloud-operators            IBMCS Operators                grpc   IBM         20m
[...]
```

If everything goes well, you should see the following pods:

```console
$ kubectl get pod -n $GLOBAL_CATALOG_NAMESPACE
NAME                                            READY   STATUS    RESTARTS   AGE
marketplace-operator-6576b4ddc8-dqcgr           1/1     Running   0          84s
opencloud-operators-66df4d97ff-4rhjj            1/1     Running   0          80s
upstream-community-operators-7ffb6b674b-7qlvx   1/1     Running   0          80s
[...]
```

If something seems wrong check the [troubleshooting section](#createcontainerconfigerror-marketplace-operator-error)

4\. **View Available Operators**

Once the `OperatorSource` and `CatalogSource` are deployed, the following command can be used to list the available operators including ibm-licensing-operator-app.
**Note:** The command assumes that the of the `OperatorSource` object is `opencloud-operators`. Adjust if needed.

```console
$ kubectl get opsrc opencloud-operators -o=custom-columns=NAME:.metadata.name,PACKAGES:.status.packages -n $GLOBAL_CATALOG_NAMESPACE
NAME                  PACKAGES
opencloud-operators   ibm-meta-operator-bridge-app,ibm-commonui-operator-app,ibm-catalog-ui-operator-app,ibm-metering-operator-app,ibm-helm-repo-operator-app,ibm-iam-operator-app,ibm-elastic-stack-operator-app,ibm-monitoring-exporters-operator-app,ibm-monitoring-prometheusext-operator-app,cp4foobar-operator-app,ibm-healthcheck-operator-app,ibm-platform-api-operator-app,ibm-management-ingress-operator-app,ibm-helm-api-operator-app,ibm-licensing-operator-app,ibm-ingress-nginx-operator-app,ibm-monitoring-grafana-operator-app,ibm-auditlogging-operator-app,operand-deployment-lifecycle-manager-app,ibm-mgmt-repo-operator-app,ibm-mongodb-operator-app,ibm-cert-manager-operator-app
```

5\. **Create an OperatorGroup**

An `OperatorGroup` is used to denote which namespaces your Operator should be watching.
It must exist in the namespace where your operator is deployed, for example, `ibm-common-services`.

    1. Create a namespace for `IBM Licensing Operator`:

```bash
kubectl create namespace ibm-common-services
```

    2. Deploy the `OperatorGroup` resource:

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: operators.coreos.com/v1alpha2
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

 The last piece ties together all of the previous steps. A `Subscription` is created foro the operator that upgrades `IBM Licensing Operator` when needed.
 Create the `Subscription` using the following command:

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

Watch `IBM Licensing Operator` being deployed by `OLM` from the `catalog source` created by `Operator Marketplace` with the following command:

```console
$ kubectl get clusterserviceversion -n ibm-common-services
NAME                            DISPLAY                  VERSION   REPLACES                        PHASE
ibm-licensing-operator.v1.0.0   IBM Licensing Operator   1.0.0     ibm-licensing-operator.v0.0.0   Succeeded
```

**Note:** The above command assumes that you have created the `Subscription` in the `ibm-common-services` namespace.
If your Operator deployment (CSV) shows `Succeeded` in the `InstallPhase` status, your Operator is deployed successfully. Otherwise, check the `ClusterServiceVersion` objects status for details.

Optional: You can also check your Operator's deployment:

```bash
kubectl get deployment -n ibm-common-services | grep ibm-licensing-operator
```

### Post-installation steps

After you successfully install IBM Licensing Operator, you can create IBMLicensing instance that will make IBM Licensing Service run on cluster.

#### Create instance on OCP 4.2+

If you have OCP 4.2+ you can create the instance from the OCP UI: **Installed Operators->IBM Licensing Operator->IBM Licensing tab->Create IBMLicensing**

![OCP click Create IBM Licensing](images/ocp_create_instance.png)

By clicking `Create IBMLicensing`, you can edit your parameters.

**Note:** Make sure to change datasource to `datacollector`.

[OCP instance datacollector](images/ocp_instance_datacollector.png)

For more information about the parameters, see [IBMLicensingOperatorParameters](images/IBMLicensingOperatorParameters.csv).

If you are sure about your parameters click **Create**.
To edit your instance in the future, go to:
**OCP UI->Administration->Custom Resource Definitions->select IBMLicensing->instances->Edit IBMLicensing**

[OCP Edit Instance](images/ocp_edit_instance.png)

**Note**: If the instance is not updated properly, try deleting the instance and creating new one with new parameters.

#### Creating an instance from console

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

**Configuring ingress**
You might want to configure ingress. Here is an example of how you can do it:
1. Get the nginx ingress controller, for example from here: [https://kubernetes.github.io/ingress-nginx/deploy](https://kubernetes.github.io/ingress-nginx/deploy)

2. Apply this IBMLicensing instance to your cluster:

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
      "nginx.ingress.kubernetes.io/rewrite-target": "/$2"
    path: /ibm-licensing-service-instance/(/|$)(.*)
EOF
```

3. Access the inctanse at your ingress host with the following path: `/ibm-licensing-service-instance`.

**Note:** For HTTPS,set `spec.httpsEnable` to `true`, and edit `ingressOptions`. Read more about the options here:
[IBMLicensingOperatorParameters](images/IBMLicensingOperatorParameters.csv)

#### Check Components

1. Check whether the pod is created.
To see the logs go to **OCP UI->Workloads->Pods** and search for `licensing` in `ibm-common-services` project:

![OCP Pod](images/ocp_pod.png)

2. Check if the pod is running. To investigate further select `licensing` and check logs, and events.
You can also run the following command from the console:

```bash
podName=`kubectl get pod -n ibm-common-services -o jsonpath="{range .items[*]}{.metadata.name}{'\n'}" | grep ibm-licensing-service-instance`
kubectl logs $podName -n ibm-common-services
kubectl describe pod $podName -n ibm-common-services
```

3. Check Route or Ingress settings depanding on your parameter settings.

You can check the Route or Ingress Settings in **OCP UI->Networking->Service**. Or you can check if using the console command, for example: 

```bash
kubectl get ingress -n ibm-common-services -o yaml
```

### Using IBM License Service to retrieve license usage information

For more information about how to use License Service to retrieve license usage data, se [IBM Cloud Platform Common Services documentation](https://www.ibm.com/support/knowledgecenter/SSHKN6/license-service/1.0.0/retrieving.html).

### Uninstalling License Service from a Kubernetes cluster

**Note:** The following procedesure assumes that you have deployed IBM License Service in the `ibm-common-services` namespace

1. **Delete the `IBMLicensing custom` resource**

Delete the instance and the operator cleand its resources.
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

2. **Delete the operator subscription**

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
3. **Delete Cluster Service Version (CSV)**

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

4. **Delete Custom Resource Definition (CRD)**

Delete the custom resource definition with the following command:

```bash
kubectl delete CustomResourceDefinition ibmlicensings.operator.ibm.com
```

5. **Delete Operator Group**

**Note:** If you have other subscriptions that are tied with that operatorGroup do not delete it.
IBM Licensing Operator is now uninstalled.You can also clean up the operatorgroup that you created for subscription by using the following command:

```bash
licensingNamespace=ibm-common-services
operatorGroupName=operatorgroup
kubectl delete OperatorGroup ${operatorGroupName} -n ${licensingNamespace}
```

6. **Delete OperatorSource**

**Note:** If you have other services that use the opencloudio catalog source do not delete the OperatorSource.
Otherwise, you can delete the OperatorSource with the following command:

```bash
export GLOBAL_CATALOG_NAMESPACE=olm
opencloudioSourceName=opencloud-operators
kubectl delete OperatorSource ${opencloudioSourceName} -n ${GLOBAL_CATALOG_NAMESPACE}
```

7. **Delete OperatorMarketplace**

**Important:** Do not delete the OperatorMarketplace if it use it with a different operator.

You can delete the OperatorMarketplace with the following command:

```bash
export GLOBAL_CATALOG_NAMESPACE=olm
kubectl delete -f operator-marketplace/deploy/upstream
```

8. **Uninstall OLM**

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

In case of poblems  during the installation of operator-marketplace, check the pods in the marketplace and their status.

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

In case the following error apprears in the pod status: `container has runAsNonRoot and image has non-numeric user (marketplace-operator), cannot verify user is non-root`, fix it by adding securityContext to operator-marketplace/deploy/upstream:

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
