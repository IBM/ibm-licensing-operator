# Documentation change request — Helm chart installation split

**Target page:** https://www.ibm.com/docs/en/cloud-paks/foundational-services/4.19.x?topic=service-installing-license-helm-charts

**Reason for change:**  
The single `ibm-licensing-cluster-scoped` Helm chart previously contained both cluster-scoped
and namespace-scoped resources. It has been split into two charts:

| Chart | Scope | Resources |
|---|---|---|
| `ibm-licensing-cluster-scoped` | Cluster | CRDs, ClusterRoles, ClusterRoleBindings |
| `ibm-licensing` *(new)* | Namespace | Operator Deployment, ServiceAccounts, Roles, RoleBindings, IBMLicensing CR |

Additionally the installation method changes from `helm template … | kubectl apply -f -`
to `helm upgrade --install`, which is the standard Helm release-tracked approach and avoids
the double-run workaround.

---

## Section: "Installing"

### Current content

> 1. Download the latest version of the ibm-licensing-cluster-scoped Helm chart from the
>    official IBM Helm charts repository or save its raw GitHub URL.
>
> 2. Create the ibm-licensing namespace.
>
> 3. Optional: To configure your installation, see Configuring the installation.
>
> 4. To install License Service, run the Helm template TWICE with the downloaded files
>    or the raw URL.
>
>    ```
>    helm template ibm-licensing-cluster-scoped <ibm-licensing-cluster-scoped-chart> | kubectl apply -f -
>    helm template ibm-licensing-cluster-scoped <ibm-licensing-cluster-scoped-chart> | kubectl apply -f -
>    ```
>
>    Where `<ibm-licensing-cluster-scoped-chart>` is a path to the downloaded Helm chart
>    or the raw GitHub URL.
>
>    For example:
>
>    ```
>    helm template ibm-licensing-cluster-scoped https://github.com/IBM/charts/raw/refs/heads/master/repo/ibm-helm/ibm-licensing-cluster-scoped-4.2.16+20250606.101044.0.tgz | kubectl apply -f -
>    helm template ibm-licensing-cluster-scoped https://github.com/IBM/charts/raw/refs/heads/master/repo/ibm-helm/ibm-licensing-cluster-scoped-4.2.16+20250606.101044.0.tgz | kubectl apply -f -
>    ```

### New content

> 1. Download both Helm chart packages from the official IBM Helm charts repository:
>    - `ibm-licensing-cluster-scoped-<version>.tgz`
>    - `ibm-licensing-<version>.tgz`
>
> 2. Create the target namespace:
>
>    ```
>    kubectl create namespace <namespace>
>    ```
>
> 3. [OPTIONAL] If your registry requires authentication, create an image pull secret in the
>    namespace. The secret must exist before the operator pod starts:
>
>    ```
>    kubectl create secret docker-registry <pull-secret-name> \
>      --docker-server=<registry-server> \
>      --docker-username=<username> \
>      --docker-password=<password> \
>      --namespace=<namespace>
>    ```
>   
>   setup the global.imagePullSecret in values.yaml
> 
>    ```
>    global:
>      imagePullSecret: artifactory-pull-secret
>    ```
> 4. Optional: To configure your installation, see Configuring the installation.
>    Prepare a single values file — it will be passed to both charts.
>
> 5. Install the cluster-scoped chart first (CRDs, ClusterRoles, ClusterRoleBindings):
>
>    ```
>    helm upgrade --install ibm-licensing-cluster-scoped \
>      <ibm-licensing-cluster-scoped-chart> \
>      --namespace <namespace> \
>      -f <custom-values>.yaml
>    ```  
> 
> 6. Wait for the CRDs to be fully established before continuing:
>
>    ```
>    kubectl wait crd \
>      ibmlicensings.operator.ibm.com \
>      ibmlicensingquerysources.operator.ibm.com \
>      ibmlicensingmetadatas.operator.ibm.com \
>      ibmlicensingdefinitions.operator.ibm.com \
>      --for=condition=Established \
>      --timeout=60s
>    ```
>
> 7. Install the namespace-scoped chart (Operator Deployment, ServiceAccounts,
>    Roles, RoleBindings, IBMLicensing CR):
>
>    ```
>    helm upgrade --install ibm-licensing \
>      <ibm-licensing-chart> \
>      --namespace <namespace> \
>      -f <custom-values>.yaml
>    ```
>
>    Where:
>    - `<ibm-licensing-cluster-scoped-chart>` is a path to the downloaded cluster-scoped chart.
>    - `<ibm-licensing-chart>` is a path to the downloaded namespace-scoped chart.
>    - `<custom-values>.yaml` is the file with your custom settings — **the same file must be
>      passed to both charts**.
>
> **Important:** Both `helm upgrade --install` commands must receive the **same**
> `-f <custom-values>.yaml`. The cluster-scoped chart embeds the operator namespace in its
> ClusterRoleBindings. If the two charts are installed with different namespace values, the
> operator will fail at startup with a `forbidden` error when listing its own custom resources.

---

## Section: "Configuring the installation"

### Current content — method 1 (values file)

> Create a YAML file with your custom settings and use the `-f` flag during the Helm install.
> It will override default settings that are present in the `values.yaml` file.
>
> ```
> helm template install ibm-licensing-cluster-scoped <ibm-licensing-cluster-scoped-chart> -f <new-values-yaml> | kubectl apply -f -
> ```
>
> Where:
> - `<ibm-licensing-cluster-scoped-chart>` is a path to the downloaded Helm chart or the raw GitHub URL.
> - `<new-values-yaml>` is the file with your custom settings.

### Current content — method 2 (--set)

> Use the `--set <key>=<value>` to override parameters directly in the command:
>
> ```
> helm install ibm-licensing-cluster-scoped <ibm-licensing-cluster-scoped-chart> --set <key>=<value> | kubectl apply -f -
> ```
>
> Where:
> - `<ibm-licensing-cluster-scoped-chart>` is a path to the downloaded Helm chart or the raw GitHub URL.
> - `<key>=<value>` is your custom setting. For example: `ibmLicensing.operator.labels.namespace=mynamespace`.

### New content

> You can configure the installation by creating a YAML file with your custom settings and
> passing it to both `helm upgrade --install` commands with the `-f` flag.
>
> ```
> helm upgrade --install ibm-licensing-cluster-scoped <ibm-licensing-cluster-scoped-chart> \
>   --namespace <namespace> -f <custom-values>.yaml
>
> helm upgrade --install ibm-licensing <ibm-licensing-chart> \
>   --namespace <namespace> -f <custom-values>.yaml
> ```
>
> Alternatively, use `--set <key>=<value>` to override parameters directly in the command.
> When using `--set`, apply the same override to both charts where applicable.
>
> **Note:** The `--set` approach is not recommended for the `ibmLicensing.namespace` parameter
> — always use a values file to ensure both charts receive the same namespace value.

The configuration sub-sections (Namespace, Custom resource, Metadata, Specifying image
registry, Specifying image pull secrets, Accepting license, Watch namespaces) remain
**unchanged** — they describe values file parameters that apply to both charts without
modification.

---

## Section: "Uninstalling" *(new section — does not currently exist)*

### New content

> To uninstall License Service, remove both Helm releases and manually delete the CRDs.
> Helm does not remove CRDs automatically to protect existing custom resource data.
>
> 1. Uninstall the namespace-scoped release. This removes the Operator Deployment,
>    ServiceAccounts, Roles, RoleBindings, and the IBMLicensing CR:
>
>    ```
>    helm uninstall ibm-licensing --namespace <namespace>
>    ```
>
> 2. Uninstall the cluster-scoped release. This removes the ClusterRoles and
>    ClusterRoleBindings:
>
>    ```
>    helm uninstall ibm-licensing-cluster-scoped --namespace <namespace>
>    ```
>
> 3. [OPTIONAL] Delete the CRDs manually:
>
>    ```
>    kubectl delete crd \
>      ibmlicensings.operator.ibm.com \
>      ibmlicensingquerysources.operator.ibm.com \
>      ibmlicensingmetadatas.operator.ibm.com \
>      ibmlicensingdefinitions.operator.ibm.com \
>      --ignore-not-found=true
>    ```
>
> 4. Delete the namespace:
>
>    ```
>    kubectl delete namespace <namespace>
>    ```
>
> **Note:** If any resources remain after the Helm uninstall (for example, resources created
> by the operator at runtime), delete them by label:
>
> ```
> kubectl delete all,serviceaccount,configmap,secret,pvc,ingress \
>   --namespace <namespace> \
>   -l app.kubernetes.io/managed-by=ibm-licensing-operator \
>   --ignore-not-found=true
> ```

---

## Summary of all changes

| Section | Type | Description |
|---|---|---|
| Installing — step 1 | Changed | One chart to download → two charts (`ibm-licensing-cluster-scoped` + `ibm-licensing`) |
| Installing — step 2 | Unchanged | Create namespace |
| Installing — step 3 | New | Create image pull secret if registry requires authentication |
| Installing — step 4 | Changed | `helm template … \| kubectl apply -f -` (run twice) → `helm upgrade --install` (standard release-tracked install) |
| Installing — step 4 | New | CRD establishment wait (`kubectl wait crd … --for=condition=Established`) between the two chart installs |
| Installing — warning | New | Both charts must receive the same `-f values` file |
| Configuring — commands | Changed | `helm template … \| kubectl apply` → `helm upgrade --install` in all examples |
| Configuring — scope note | New | Clarify that the same values file is passed to both charts; `--set` caveats for namespace |
| Configuring — sub-sections | Unchanged | Namespace, Custom resource, Metadata, Image registry, Pull secrets, License, Watch namespaces |
| Uninstalling | New section | `helm uninstall` for both releases + manual CRD deletion + namespace deletion |
