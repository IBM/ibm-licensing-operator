# IBM Licensing components as ArgoCD Applications

Provided functionalities:
- Installation of IBM License Service (ILS), ILS Reporter, and ILS Scanner
- Configuration of the components so that the whole licensing suite is functional

## Prerequisites

There is a number of steps involved before it's possible to deploy ArgoCD applications:
- There must be a cluster with ArgoCD installed
- ArgoCD application controller must have all required permissions OR
- The prerequisites to install the applications must be met

Below are instructions on how to provision and configure a cluster for IBM Licensing components.

### Install ArgoCD on an Openshift cluster

- Install *Red Hat OpenShift GitOps* from the *OperatorHub* (see
[RedHat documentation](https://docs.openshift.com/gitops/1.14/installing_gitops/installing-openshift-gitops.html)
for more information):
    ![install-red-hat-openshift-gitops-step-1](docs/images/install-red-hat-openshift-gitops-step-1.png)
    ![install-red-hat-openshift-gitops-step-2](docs/images/install-red-hat-openshift-gitops-step-2.png)

- Access *ArgoCD* UI:
    ![argo-cd-ui-step-1.png](docs/images/argo-cd-ui-step-1.png)
    ![argo-cd-ui-step-2.png](docs/images/argo-cd-ui-step-2.png)

- Log in via *OpenShift* and check the Applications screen is accessible:
    ![applications-screen.png](docs/images/applications-screen.png)

### Apply prerequisites

There are multiple ways to apply prerequisites in your cluster. We recommend that the cluster admins review and apply
required modifications manually, however, this can also be automated.

#### Apply the yaml files

You can apply (assuming you are logged in to the cluster) all prerequisites required for IBM Licensing components
with a simple command executed on the `prerequisites` directory:

```shell
oc apply -f prerequisites --recursive
```

Note that some values (such as namespaces or annotations) may need adjustment depending on your desired results.

#### Include prerequisites as part of your ArgoCD deployment

To automate prerequisites deployment, you can include yaml files from the `prerequisites` directory in your ArgoCD
applications' paths. To make sure they are applied before the IBM Licensing components are installed, you can use
[sync waves](https://argo-cd.readthedocs.io/en/latest/user-guide/sync-waves/). For example, through annotating required
resources with the `PreSync` phase.

## Configuration

We recommend that you adjust the `Application` yaml files to configure the components' `helm` charts. Please check
the [ArgoCD user guide](https://argo-cd.readthedocs.io/en/latest/user-guide/helm/) on `helm` for more details.
In general, the modifications will be introduced through this structure:

```yaml
source:
  helm:
    valuesObject:
      key: new-value
```

Alternatively, you may want to adjust the yaml files within the `components` directory itself, before deploying
an `Application` targeting them. For example, you could fork this repository and adjust some custom resource
configuration directly in the relevant file.

For your convenience, below are some common scenarios with examples on how to resolve provided, sample issues.

### Configure the CR

To configure licensing components through custom resources, please modify the `spec` section. For example, to enable
hyper-threading in license service:

```yaml
source:
  helm:
    valuesObject:
      spec:
        features:
          hyperThreading:
            threadsPerCore: <number of threads>
```

Please refer to the components' official documentation to learn more about the supported configuration options. You may
find relevant sections under the following links:
- [*License Service*](https://www.ibm.com/docs/en/cloud-paks/foundational-services/4.6?topic=service-configuration)
- [*License Service Reporter*](https://www.ibm.com/docs/en/cloud-paks/foundational-services/4.6?topic=reporter-installing-configuring-license-service)
- *License Service Scanner* -> the official documentation is not yet available publicly, please contact us to learn more

### Change target namespace

By default, IBM Licensing components are installed in three different namespaces, to separate the resources, and to
group them up by the component. If you want to install a specific component in a different namespace:

```yaml
source:
  helm:
    valuesObject:
      namespace: my-custom-namespace
```

### Apply custom metadata

To apply custom labels and annotations to the operator-managed resources:

```yaml
source:
  helm:
    valuesObject:
      spec:
        labels:
          appName: LicenseService
        annotations:
          companyName: IBM
```

To apply custom labels and annotations to the operator deployment:

```yaml
source:
  helm:
    valuesObject:
      operator:
        labels:
          appName: LicenseService
        annotations:
          companyName: IBM
```

Note that these labels and annotations are in addition to the default ones, and will not override them.

### Enable auto-connect

To ensure that *IBM License Service Scanner* can connect with *IBM License Service* automatically, please add the
scanner's namespace to the `watchNamespace` field of `license-service.yaml`, for example:

```yaml
source:
  helm:
    valuesObject:
      watchNamespace: ibm-licensing,ibm-licensing-scanner
```

Without this change, the following `INFO` log will appear on the License Service operator side, after the application
of `Application` from `scanner.yaml`:
```text
INFO operandrequest-discovery OperandRequest for ibm-licensing-operator detected. IBMLicensing OperatorGroup will be extended {"OperandRequest": "ibm-licensing-scanner-ls-operand-request", "Namespace": "ibm-licensing-scanner"}
INFO operandrequest-discovery OperatorGroup for IBMLicensing operator not found {"Namespace": "ibm-licensing"}
```

Furthermore, you must make sure that the `licenseServiceNamespace` field in `scanner.yaml` is matching your
configuration. By default, the following namespace is expected:

```yaml
source:
  helm:
    valuesObject:
      licenseServiceNamespace: ibm-licensing
```

Otherwise, License Service operator will log errors related to missing RBAC permissions.

## Installation

To install all components, execute the following command (assuming you are logged in to your cluster):

```shell
oc project openshift-gitops && oc apply -f applications
```

![components.png](docs/images/components.png)

To install selected components separately, for example to install *IBM License Service* only, execute this command:

```shell
oc project openshift-gitops && oc apply -f applications/license-service.yaml
```

Remember to `sync` after the applications are applied, or add the `auto-sync` option to your setup.

### Separate installation scenario

Installing components separately is recommended for example when you want to install *IBM License Service Reporter*
on a different cluster.

The steps in such scenario would be as follows:
- Apply `applications/reporter.yaml` to your cluster
- Follow official *IBM License Service* docs to prepare connection secret and CR configuration
- Add the values to `applications/license-service.yaml` to configure the connection
- Apply `applications/license-service.yaml` to your cluster and check both components are working and connected
