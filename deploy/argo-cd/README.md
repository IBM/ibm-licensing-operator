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

1. Install *Red Hat OpenShift GitOps* from the *OperatorHub* (see 
[RedHat documentation](https://docs.openshift.com/gitops/1.14/installing_gitops/installing-openshift-gitops.html)
for more information):

![install-red-hat-openshift-gitops-step-1](docs/images/install-red-hat-openshift-gitops-step-1.png)
![install-red-hat-openshift-gitops-step-2](docs/images/install-red-hat-openshift-gitops-step-2.png)

2. Access *ArgoCD* UI:

![argo-cd-ui-step-1.png](docs/images/argo-cd-ui-step-1.png)
![argo-cd-ui-step-2.png](docs/images/argo-cd-ui-step-2.png)

3. Log in via *OpenShift* and check the Applications screen is accessible:

![applications-screen.png](docs/images/applications-screen.png)

### Apply prerequisites

There are multiple ways to apply prerequisites in your cluster. We recommend that the cluster admins review and apply
required modifications manually, however, this can also be automated.

#### Apply the yaml files

You can apply (assuming you are logged in to the cluster) all prerequisites required for IBM Licensing components
with a simple command executed on the `prerequisites` directory:

```commandline
oc apply -f prerequisites --recursive
```

Note that some values (such as namespaces or annotations) may need adjustment depending on your desired results.

#### Include prerequisites as part of your ArgoCD deployment

To automate prerequisites deployment, you can include yaml files from the `prerequisites` directory in your ArgoCD
applications' paths. To make sure they are applied before the IBM Licensing components are installed, you can use
[sync waves](https://argo-cd.readthedocs.io/en/latest/user-guide/sync-waves/). For example, through annotating required
resources with the `PreSync` phase.

## Installation

To install all components, execute the following command (assuming you are logged in to your cluster):
```commandline
oc project openshift-gitops && oc apply -f applications
```

If you wish to configure the components (e.g. connect *IBM License Service* with *IBM License Service Reporter*), please
refer to the official documentation for each component and modify the `spec` section in the `values.yaml` files within
the `components` directory.

Please note that *IBM License Service Scanner* is not yet officially documented - contact us to learn more about it.

![components.png](docs/images/components.png)

To install selected components separately, for example to install *IBM License Service* only, execute this command:
```commandline
oc project openshift-gitops && oc apply -f applications/license-service.yaml
```

Installing components separately is recommended for example when you want to install *IBM License Service Reporter*
on a different cluster.

The steps in such scenario would be as follows:
- Apply `applications/reporter.yaml` to your cluster
- Follow official *IBM License Service* docs to prepare connection secret and CR configuration
- Modify `components/license-service/values.yaml` to perform the connection
- Apply `applications/license-service.yaml` to your cluster and check both components are working and connected

## Configuration

We recommend that you adjust the `Application` yaml files to configure the components' `helm` charts. Please check
the [ArgoCD user guide](https://argo-cd.readthedocs.io/en/latest/user-guide/helm/) on `helm` for more details.

Alternatively, you may want to adjust the yaml files within the `components` directory itself, before deploying
an `Application` targeting them. For example, you could fork this repository and adjust some custom resource
configuration directly in the relevant file.

For your convenience, below are some common scenarios with examples on how to resolve provided, sample issues.

### With helm

Since the YAML files provided as part of the `components` directory are templated with `helm`, you can add the following
section to the `Application` files, to modify some templated field:

```yaml
source:
  helm:
    valuesObject:
      key: new-value
```

Naturally, you can also fork/copy this repository and apply the changes yourself to `values.yaml` files.

#### Configure the CR

To configure licensing components through custom resources, please modify the `spec` section. For example, to accept
the license terms:

```yaml
source:
  helm:
    valuesObject:
      spec:
        license:
          accept: true
```

Please refer to the components' official documentation to learn more about the supported configuration options.

#### Change target namespace

By default, IBM Licensing components are installed in three different namespaces, to separate the resources, and to
group them up by the component. If you want to install a specific component in a different namespace:

```yaml
source:
  helm:
    valuesObject:
      namespace: my-custom-namespace
```

#### Apply custom metadata

To apply custom labels and annotations please refer to the official documentation for each component and apply the
changes to the `spec` section:

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

Note that these labels and annotations are added in addition of the default ones, and will not override them.