# Installing IBM Licensing components as Argo CD applications

Learn how to install IBM Licensing components as Argo CD applications. This procedure guides you through the following steps:

- [Prerequisites](#prerequisites): Preparing for the installation.
- [(Optional) Configuring](#configuring): Configuring the components to ensure that the Licensing suite is functional.
- [(Optional) Downloading](#downloading-helm-charts-and-images-for-air-gapped-environments): Downloading Helm charts and images for air-gapped environments.
- [Installing](#installing): Installation of IBM License Service and IBM License Service Reporter.

## Prerequisites

Before you can deploy Argo CD applications:

- You must have a cluster with Argo CD installed.
- ArgoCD application controller must have all required permissions. See the [prerequisites directory](prerequisites).

Perform the following steps to provision and configure a cluster for IBM Licensing components.

### 1. Install Argo CD

#### On an Openshift cluster

1.1. Log in to the OpenShift console and go to **Operators > OperatorHub** or **Ecosystem > Software Catalog** depending on the version of your console. Install *Red Hat OpenShift GitOps*. For more information, see [RedHat documentation](https://docs.openshift.com/gitops/1.14/installing_gitops/installing-openshift-gitops.html).

1.2. To access the *Argo CD* user interface, click the Application Launcher icon in the top-right corner, and click **Cluster Argo CD**:
    ![argo-cd-ui-step-1.png](docs/images/argo-cd-ui-step-1.png)

1.3. Log in via *OpenShift* and check whether the Applications screen is accessible.
    ![applications-screen.png](docs/images/applications-screen.png)

#### On EKS

1.1. Install *Argo CD* by following the official [AWS documentation](https://docs.aws.amazon.com/eks/latest/userguide/argocd.html).

1.2. Access *Argo CD* user interface:
    ![argo-cd-ui-eks-step-1.png](docs/images/argo-cd-ui-eks-step-1.png)

1.3. Log in with the IAM Identity Center user and check whether the Applications screen is accessible.
    ![applications-screen-eks.png](docs/images/applications-screen-eks.png)

### 2. Apply prerequisites

You can apply prerequisites in multiple ways. It is recommended for the cluster administrators to review and apply the required modifications manually. However, it can also be automated. To apply teh prerequisites, first clone or download this repository.

    ```bash
    git clone --single-branch --branch latest-4.x https://github.com/IBM/ibm-licensing-operator.git
    ```

#### Apply the .yaml files

Log in to the cluster and run the following command on the `prerequisites` directory to apply prerequisites for the IBM Licensing components.

```shell
kubectl apply -f <path-to-cloned-repo>/deploy/argo-cd/prerequisites --recursive
```

**Note:** Some values, such as namespaces or annotations, might need adjustment depending on your desired results.

#### Include prerequisites as a part of your Argo CD deployment

To automate the deployment of prerequisites, include the .yaml files from the `prerequisites` directory in your
Argo CD applications' paths. To make sure that they are applied before the IBM Licensing components are installed,
you can use [sync waves](https://argo-cd.readthedocs.io/en/latest/user-guide/sync-waves/). For example, through
annotating the required resources with the `PreSync` phase.

## Configuring

### Configuring by adjusting .yaml files

It is recommended to adjust the `Application` .yaml files to configure the `helm` charts of the components. For more
information, see the [Argo CD user guide](https://argo-cd.readthedocs.io/en/latest/user-guide/helm/) on `helm`.

In general, for applications with multiple sources, the modifications are introduced with the following structure:

```yaml
sources:
  - helm:
      valuesObject:
        key: new-value
```

Alternatively, you can adjust the .yaml files within the `components` directory itself or the `values.yaml`
files, before deploying an `Application` targeting them. For example, you can fork this repository and adjust
some custom resource configuration directly in the relevant file.

### Configuring through CR

To configure the Licensing components through custom resources, modify the `spec` section. For example, to enable
hyper-threading in License Service, add the following lines:

```yaml
helm:
  valuesObject:
    spec:
      features:
        hyperThreading:
          threadsPerCore: <number of threads>
```

To learn more about the supported configuration options, see the official documentation for License Service and License Service Reporter. See the relevant sections under the following links:

- [*License Service*](https://www.ibm.com/docs/en/cloud-paks/foundational-services/4.x_cd?topic=service-configuring)
- [*License Service Reporter*](https://www.ibm.com/docs/en/cloud-paks/foundational-services/4.x_cd?topic=reporter-installing-configuring-license-service)

### Configurable properties

The following are some common scenarios with examples on how to resolve the provided sample issues.

**Note:** Apply the following examples to the following sources:

- For License Service, apply the examples to the `helm-cluster-scoped` source that supports full CR configuration.
- For License Service Reporter, apply the examples to the sources that targets the path with `helm`, not `helm-cluster-scoped`, as
the cluster-scoped charts only support the `namespace` parameter

#### Changing the target namespace

<details>
<summary>Click to expand</summary>

By default, IBM Licensing components are installed in three different namespaces to separate the resources, and
group them by the component. If you want to install a specific component in a different namespace, change the following lines:

```yaml
helm:
  valuesObject:
    namespace: my-custom-namespace
```

**Note:** When you change the `namespace` value in `applications/license-service.yaml`, in general, you should also modify the value of the `watchNamespace` parameter.

</details>

#### Applying custom metadata

<details>
<summary>Click to expand</summary>

- To apply custom labels and annotations to the operator-managed resources, change the following lines:

  ```yaml
  helm:
    valuesObject:
      spec:
        labels:
          appName: LicenseService
        annotations:
          companyName: IBM
  ```

- To apply custom labels and annotations to the operator deployment, change the following lines:

  ```yaml
  helm:
    valuesObject:
      operator:
        labels:
          appName: LicenseService
        annotations:
          companyName: IBM
  ```

**Note:** Custom labels and annotations are additions to the default ones, and they do not override them.

</details>

#### Specifying image registry and image registry namespace

<details>
<summary>Click to expand</summary>

To specify a different image registry for the installation of the components, change the value of `global.imagePullPrefix` in the relevant `Application.yaml` file:

```yaml
helm:
  valuesObject:
    global:
      imagePullPrefix: <your-registry>
```

As a result, the operator and operand image registries are overwritten. For example, after applying the above changes to the `applications/license-service.yaml` file, the image of the `ibm-licensing-operator`
becomes `<your-registry>/cpopen/ibm-licensing-operator:4.2.21`.

To additionally modify the image registry namespace of either the operator or the operand, change the value of `cpfs.imageRegistryNamespaceOperator` or `cpfs.imageRegistryNamespaceOperand`, or both, in the relevant `Application.yaml` file:

```yaml
helm:
  valuesObject:
    cpfs:
      imageRegistryNamespaceOperator: <your-operator-image-registry-namespace>
      imageRegistryNamespaceOperand: <your-operand-image-registry-namespace>
```

As a result, the operator and operand image registry namespaces are overwritten. For example, after applying the above
changes to the `applications/license-service.yaml` file, the image of the `ibm-licensing-operator` becomes
`icr.io/<your-operator-image-registry-namespace>/ibm-licensing-operator:4.2.21`.

</details>

#### Specifying image pull secrets

<details>
<summary>Click to expand</summary>

To specify which image pull secret should be used to pull from the registry, change the value of `global.imagePullSecret` in the relevant `Application.yaml` file:

```yaml
helm:
  valuesObject:
    global:
      imagePullSecret: <your-secret>
```

As a result, the `imagePullSecrets` field of the operator and the operand include the specified secret, and this secret is used when pulling the images from the registry.

**Note:** `global.imagePullSecret` is added to the list of secrets provided in the CR configuration, for example, through `spec.imagePullSecrets`.

</details>

## Downloading Helm charts and images for air-gapped environments

To install License Service or License Service Reporter as Argo CD applications in an air-gapped environment, you need download Helm charts and container images and store them in your local repository.


1. Clone or download this repository.

    ```bash
    git clone --single-branch --branch latest-4.x https://github.com/IBM/ibm-licensing-operator.git
    ```

2. Store this repository in your local repository.

    ```bash
    git remote add local-repo <custom_repository>
    git push local-repo latest-4.x
    ```

3. Adjust `repoURL` in the .yaml file for every component in ./argo-cd/applications/*component*.yaml

    For example, if you are installing License Service, modify the `license-service.yaml` file.

    **Before**

    ```yaml
    apiVersion: argoproj.io/v1alpha1
    kind: Application
    metadata:
      name: ibm-license-service
      finalizers:
        - resources-finalizer.argocd.argoproj.io
    spec:
      project: default
      destination:
        server: https://kubernetes.default.svc
      sources:
        - repoURL: "https://github.com/IBM/ibm-licensing-operator"
          targetRevision: "latest-4.x"
          path: deploy/argo-cd/components/license-service/helm-cluster-scoped
    ```

    **After**

    ```yaml
    apiVersion: argoproj.io/v1alpha1
    kind: Application
    metadata:
      name: ibm-license-service
      finalizers:
        - resources-finalizer.argocd.argoproj.io
    spec:
      project: default
      destination:
        server: https://kubernetes.default.svc
      sources:
        - repoURL: "https://github.com/custom-company/custom-repo" # Link to your custom repository
          targetRevision: "latest-4.x"
          path: deploy/argo-cd/components/license-service/helm-cluster-scoped
    ```

    As a result, Helm charts will be pulled from your local repository during the installation.

4. Download container images and store them in your local repository.

    - For License Service, download the following images.

        ```
        icr.io/cpopen/ibm-licensing-operator:<version>
        icr.io/cpopen/cpfs/ibm-licensing:<version>
        ```

    - For License Service Reporter, download the following images.

        ```
        icr.io/cpopen/cpfs/ibm-postgresql:<version>
        icr.io/cpopen/cpfs/ibm-license-service-reporter:<version>
        icr.io/cpopen/cpfs/ibm-license-service-reporter-ui:<version>
        icr.io/cpopen/cpfs/ibm-license-service-reporter-oauth2-proxy:<version>
        icr.io/cpopen/ibm-license-service-reporter-operator:<version>
        ```

    <details>
    <summary>Tip: You can run the following command to get the list of the required images.</summary>

    ```bash
    helm template ./deploy/argo-cd/components/license-service/helm-cluster-scoped | grep icr.io
    ```

    The output of the command will contain similar information:

    ```
    value: icr.io/cpopen/cpfs/ibm-licensing:4.2.21
    image: icr.io/cpopen/ibm-licensing-operator:4.2.21
    ```

    For License Service Reporter, run the command against the standard Helm charts, not the cluster-scoped charts.


    ```bash
    helm template ./deploy/argo-cd/components/reporter/helm | grep icr.io
    helm template ./deploy/argo-cd/components/scanner/helm | grep icr.io
    ```

    </details>

5. To pull the required images, run the following command.

    ```bash
    docker pull icr.io/cpopen/ibm-licensing-operator:<version>
    docker pull icr.io/cpopen/cpfs/ibm-licensing:<version>
    ```

6. Before you push the images to your private registry, ensure that you are logged in. Use the following command.

    ```bash
    docker login <docker_registry>
    ```

7. Tag the images with your registry prefix and push with the following commands.

    ```bash
    docker tag icr.io/cpopen/ibm-licensing-operator:<version> <docker_registry>/ibm-licensing-operator:<version>
    docker push <docker_registry>/ibm-licensing-operator:<version>

    docker tag icr.io/cpopen/cpfs/ibm-licensing:<version> <docker_registry>/ibm-licensing:<version>
    docker push <docker_registry>/ibm-licensing:<version>
    ```

    For example, for the License Service image version 4.2.21, run the following commands.

    ```bash
    docker pull icr.io/cpopen/ibm-licensing-operator:4.2.21
    docker tag icr.io/cpopen/ibm-licensing-operator:4.2.21 custom_repo.com/custom_namespace/ibm-licensing-operator:4.2.21
    docker push custom_repo.com/custom_namespace/ibm-licensing-operator:4.2.21
    ```

8. Update the image registry configuration. For more information, see [Specifying image registry and image registry namespace](#specifying-image-registry-and-image-registry-namespace).

    For example:

    ```bash 
    imagePullPrefix: custm_repo.com
    imageRegistryNamespaceOperator: operator_custom_namespace #namespace that you used on images that contained `operator` in name
    imageRegistryNamespaceOperand: operand_custom_namespace #namespace that you tagged on any other image
    ```


## Installing

### Installing all components

To install all components, perform the following steps.

1. Log in to your cluster, and open the following namespace.

- For OpenShift, by default `openshift-gitops`
- For EKS, by default `argocd`

  ```shell
  kubectl config set-context --current --namespace=openshift-gitops
  ```

2. Run the following command.

    ```shell
    kubectl apply -f <path-to-cloned-repo>/deploy/argo-cd/applications
    ```

**Note** Remember to `sync` after the applications are applied, or add the `auto-sync` option to your setup.

You should see the apps in Argo CD.
![components.png](docs/images/components.png)

### Installing selected components

To install selected components separately, for example to install *IBM License Service* only, perform the following steps.

1. Log in to your cluster, and open the following namespace.

- For OpenShift, by default `openshift-gitops`
- For EKS, by default `argocd`

  ```shell
  kubectl config set-context --current --namespace=openshift-gitops
  ```

2. Run the following command.

    ```shell
    kubectl apply -f<path-to-cloned-repo>/deploy/argo-cd/applications/license-service.yaml
    ```

**Note:** Remember to `sync` after the applications are applied, or add the `auto-sync` option to your setup.

### Installing on EKS clusters

You must register your cluster and modify the `server` field of your `Application`, because the default local cluster
destination is not supported. Follow official [AWS documentation](<https://docs.aws.amazon.com/eks/latest/userguide/argocd-register-clusters.html>) to
register your cluster.

You will also need to configure the right roles and permissions so that your ArgoCD instance can sync the application.

### Separate installation scenario

Installing components separately is recommended, for example, when you want to install *IBM License Service Reporter* on a different cluster.

In such scenario, complete the following steps:

1. Apply `applications/reporter.yaml` to your cluster.
2. Follow the official *IBM License Service* documentation to prepare the connection secret and CR configuration.
3. Add the values to `applications/license-service.yaml` to configure the connection.
4. Apply `applications/license-service.yaml` to your cluster and check whether both components are working and are
connected.

### Installation with helm

Helm installation support is in its alpha stage. To install the Licensing components with helm, run the following commands:

- IBM License Service:

  **Note:** License Service supports only cluster-scoped installation and only has a `helm-cluster-scoped` chart.

  ```commandline
  helm install license-service ./components/license-service/helm-cluster-scoped
  ```

- IBM License Service Reporter:

  ```commandline
  helm install reporter-cluster-scoped ./components/reporter/helm-cluster-scoped
  helm install reporter ./components/reporter/helm
  ```

Commands such as `helm upgrade` should be functional. However, due to the alpha stage of the development, they can
result in an unexpected state. Therefore, it is recommended to perform installation on a clean-state cluster.

If you already have any Licensing components installed, use the `--take-ownership` flag, which is introduced in
`helm` version `3.17.0`, when running the `install` commands.
