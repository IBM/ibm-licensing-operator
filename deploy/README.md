# Deploying with Helm

## Installation

Download the latest License Service Helm Charts from the
[official IBM Helm Charts repository](https://github.com/IBM/charts/tree/master/repo/ibm-helm).

You can download them locally and install with [`helm install`](https://helm.sh/docs/helm/helm_install/) or use
the `raw` GitHub URLs. For example:

```shell
helm install license-service-cluster-scoped https://github.com/IBM/charts/raw/refs/heads/master/repo/ibm-helm/ibm-license-service-cluster-scoped-4.2.15+20250506.101113.0.tgz
helm install license-service https://github.com/IBM/charts/raw/refs/heads/master/repo/ibm-helm/ibm-license-service-4.2.15+20250506.101113.0.tgz
```

## Configuration

You can use the `-f myvalues.yaml` argument when calling `helm install` to override the default `values.yaml` file. You can also use `--set key=value` to override them directly in the command.

### Namespace

By default, IBM License Service is installed in its recommended `ibm-licensing` namespace. If you want to install it in a different namespace, set the following parameters:

```shell
helm install license-service-cluster-scoped --set global.operatorNamespace=<custom-namespace> (...)
helm install license-service --set global.operatorNamespace=<custom-namespace> (...)
```

In general, when you change the `namespace` value, you should also modify the value of the `watchNamespace`:

```shell
helm install license-service --set ibmLicenseService.watchNamespace=<custom-namespace> (...)
```

### Custom Resource

To configure License Service custom resource, modify the `spec` section. For example, to enable hyper-threading, set the following parameter:

```shell
helm install license-service --set ibmLicenseService.spec.features.hyperThreading.threadsPerCore=<number of threads> (...)
```

To learn more about the supported configuration options, see
[the official documentation](https://www.ibm.com/docs/en/cloud-paks/foundational-services/4.12.0?topic=service-configuring).

### Metadata

Custom labels and annotations are additions to the default ones, and they do not override them.

- To apply custom labels and annotations to the operator-managed resources, set the following parameters:

```shell
helm install license-service --set ibmLicenseService.spec.labels.appName=LicenseService --set ibmLicenseService.spec.annotations.companyName=IBM (...)
```

- To apply custom labels and annotations to the operator deployment, set the following parameters:

```shell
helm install license-service --set ibmLicenseService.operator.labels.appName=LicenseService --set ibmLicenseService.operator.annotations.companyName=IBM (...)
```

### Specify image registry and image registry namespace

To specify a different image registry, set the following parameter:

```shell
helm install license-service --set global.imagePullPrefix=<your-registry> (...)
```

As a result, the operator and operand image registries are overwritten. For example, the image of `ibm-licensing-operator` becomes `<your-registry>/cpopen/ibm-licensing-operator@digest`.

To additionally modify the image registry namespace of either the operator or the operand, change the value of
`ibmLicenseService.imageRegistryNamespaceOperator`, `ibmLicenseService.imageRegistryNamespaceOperand`, or both.

```shell
helm install license-service --set ibmLicenseService.imageRegistryNamespaceOperator=<your-operator-image-registry-namespace> (...)
helm install license-service --set ibmLicenseService.imageRegistryNamespaceOperand=<your-operand-image-registry-namespace> (...)
```

As a result, the operator and operand image registry namespaces are overwritten. For example, the image of `ibm-licensing-operator` becomes `icr.io/<your-operator-image-registry-namespace>/ibm-licensing-operator@digest`.

**Note:** `global.imagePullPrefix`, `ibmLicenseService.imageRegistryNamespaceOperator` and `ibmLicenseService.imageRegistryNamespaceOperand` take precedence over any values that you provided in the CR configuration, for example, through `ibmLicenseService.spec.imageRegistry`.

### Specify image pull secrets

To specify which image pull secret should be used to pull from the registry, change the value of `global.imagePullSecret`:

```shell
helm install license-service --set global.imagePullSecret=<your-secret> (...)
```

As a result, the `imagePullSecrets` field of the operator and the operand include the specified secret. This secret is used when pulling the images from the registry.

**Note:** `global.imagePullSecret` is added to the list of secrets provided in the CR configuration, for example, through `ibmLicenseService.spec.imagePullSecrets`.

### Accept license

To accept the license, update the `global.licenseAccept` section.

```shell
helm install license-service --set global.licenseAccept=true (...)
```

**Note:** `global.licenseAccept` takes precedence over values that you provided in the CR configuration through `ibmLicenseService.spec.license.accept`.
