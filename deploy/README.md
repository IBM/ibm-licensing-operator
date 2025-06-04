# Deploying with Helm

## Installation

Download the latest IBM License Service Helm Chart from the
[official IBM Helm Charts repository](https://github.com/IBM/charts/tree/master/repo/ibm-helm) or save its `raw` GitHub URL.

If you want to configure your installation, see the [Configuration](#configuration) section.

If you want to install IBM License Service with the default configuration, run `helm template` **TWICE** with the downloaded files or the `raw` URL. For example:

```shell
helm template ibm-licensing-cluster-scoped https://github.com/IBM/charts/raw/refs/heads/master/repo/ibm-helm/<ibm-licensing-cluster-scoped-tgz-file> | kubectl apply -f -
helm template ibm-licensing-cluster-scoped https://github.com/IBM/charts/raw/refs/heads/master/repo/ibm-helm/<ibm-licensing-cluster-scoped-tgz-file> | kubectl apply -f -
```

**Note:** Running the command twice is a temporary workaround and will be remediated in the future.

## Configuration

You can use the `-f` flag when calling `helm install` to override the default `values.yaml` file:
```shell
helm install ibm-licensing-cluster-scoped -f <new-values-yaml> <ibm-licensing-cluster-scoped-chart>
```

You can also use `--set key=value` to override them directly in the command:
```shell
helm install ibm-licensing-cluster-scoped -set <key>=<value> <ibm-licensing-cluster-scoped-chart>
```

### Namespace

By default, IBM License Service is installed in its recommended `ibm-licensing` namespace. If you want to install it in a different namespace, set the following parameter:

```yaml
ibmLicensing:
  namespace: <your-custom-namespace>
```

### Custom Resource

To configure License Service custom resource, modify the `spec` section. For example, to enable hyper-threading, set the following parameter:

```yaml
ibmLicensing:
  spec:
    features:
      hyperThreading:
        threadsPerCore: <number of threads>
```

To learn more about the supported configuration options, see
[the official documentation](https://www.ibm.com/docs/en/cloud-paks/foundational-services/4.12.0?topic=service-configuring).

### Metadata

Custom labels and annotations are additions to the default ones, and they do not override them.

- To apply custom labels and annotations to the operator-managed resources, set the following parameters:

```yaml
ibmLicensing:
  spec:
    labels:
      <your-custom-label>: <some-label>
    annotations:
      <your-custom-annotation>: <some-annotation>
```

- To apply custom labels and annotations to the operator deployment, set the following parameters:

```yaml
ibmLicensing:
  operator:
    labels:
      <your-custom-label>: <some-label>
    annotations:
      <your-custom-annotation>: <some-annotation>
```

### Specify image registry and image registry namespace

To specify a different image registry, set the following parameter:

```yaml
global:
  imagePullPrefix: <your-custom-registry>
```

As a result, the operator and operand image registries are overwritten. For example, the image of `ibm-licensing-operator` becomes `<your-registry>/cpopen/ibm-licensing-operator@digest`.

To additionally modify the image registry namespace of either the operator or the operand, change the value of
`ibmLicensing.imageRegistryNamespaceOperator`, `ibmLicensing.imageRegistryNamespaceOperand`, or both.

```yaml
ibmLicensing:
  imageRegistryNamespaceOperator: <your-operator-image-registry-namespace>
  imageRegistryNamespaceOperand: <your-operand-image-registry-namespace>
```

As a result, the operator and operand image registry namespaces are overwritten. For example, the image of `ibm-licensing-operator` becomes `icr.io/<your-operator-image-registry-namespace>/ibm-licensing-operator@digest`.

**Note:** `global.imagePullPrefix`, `ibmLicensing.imageRegistryNamespaceOperator` and `ibmLicensing.imageRegistryNamespaceOperand` take precedence over any values that you provided in the CR configuration, for example, through `ibmLicensing.spec.imageRegistry`.

### Specify image pull secrets

To specify which image pull secret should be used to pull from the registry, change the value of `global.imagePullSecret`:

```yaml
global:
  imagePullSecret: <your-custom-pull-secret>
```

As a result, the `imagePullSecrets` field of the operator and the operand include the specified secret. This secret is used when pulling the images from the registry.

**Note:** `global.imagePullSecret` is added to the list of secrets provided in the CR configuration, for example, through `ibmLicensing.spec.imagePullSecrets`.

### Accept license

To accept the license terms for the particular IBM product for which you are deploying this component (ibm.biz/lsvc-lic), update the `global.licenseAccept` section.

```yaml
global:
  licenseAccept: true
```

**Note:** `global.licenseAccept` takes precedence over values that you provided in the CR configuration through `ibmLicensing.spec.license.accept`.

### Watch namespaces

By default, IBM License Service watches for `OperandRequest`-s in its own namespace only. To extend this functionality, you should set the following parameter:

```yaml
ibmLicensing:
  watchNamespace: <your-custom-namespace>,<your-operand-request-namespace>
```

For example, for IBM License Service installed in its default namespace, to watch for operand requests in `service-namespace-one` and `service-namespace-two`, set the parameter to:

```yaml
ibmLicensing:
  watchNamespace: ibm-licensing,service-namespace-one,service-namespace-two
```
