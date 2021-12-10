# Offline installation

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Creating an IBM Licensing instance](#creating-an-ibm-licensing-instance)

## Prerequisites

- A private Docker image registry where you can push the images using `Docker` and from where your cluster can pull images. For more information, see [Docker registry in Docker product documentation](https://docs.docker.com/registry/).
- Complete the offline installation on a host that meets the following criteria:
    - Has Linux or macOS operating system (or Windows with Linux Bash Shell for example from WSL).
    - Has Docker and Kubernetes CLI installed.
    - Has internet access.
    - Has access to your offline cluster via Kubernetes config.

## Installation

1\. Prepare Docker images.

a.  Run the following command to prepare your Docker images.

```bash
export my_docker_registry=<YOUR PRIVATE REGISTRY IMAGE PREFIX HERE; for example: "my.registry:5000" or "my.private.registry.example.com">
export operator_version=1.3.3
export operand_version=1.3.3
```

b. Pull the required images with the following command.

```bash
docker pull quay.io/opencloudio/ibm-licensing-operator:${operator_version}
docker pull quay.io/opencloudio/ibm-licensing:${operand_version}
```

c. Before pushing the images to your private registry, make sure that you are logged in. Use the following command.

```bash
docker login ${my_docker_registry}
```

d. Tag the images with your registry prefix and push with the following commands.

```bash
docker tag quay.io/opencloudio/ibm-licensing-operator:${operator_version} ${my_docker_registry}/ibm-licensing-operator:${operator_version}
docker push ${my_docker_registry}/ibm-licensing-operator:${operator_version}

docker tag quay.io/opencloudio/ibm-licensing:${operand_version} ${my_docker_registry}/ibm-licensing:${operand_version}
docker push ${my_docker_registry}/ibm-licensing:${operand_version}
```

2\. Create the required resources.

a. Run the following command on machine where you have access to your cluster and can use `kubectl`.

```bash
export my_docker_registry=<SAME REGISTRY AS BEFORE>
```

b. Run the following command to create the `ibm-common-services` namespace where you will later install the operator.

```bash
kubectl create namespace ibm-common-services
```

c. If your cluster needs the access token to your private Docker registry, create the secret in the `ibm-common-services` namespace:

```bash
kubectl create secret -n ibm-common-services docker-registry my-registry-token --docker-server=${my_docker_registry} --docker-username=<YOUR_REGISTRY_USERNAME> --docker-password=<YOUR_REGISTRY_TOKEN> --docker-email=<YOUR_REGISTRY_EMAIL, probably can be same as username>
```

d. Set the context so that the resources are made in the `ibm-common-services` namespace:

```bash
kubectl config set-context --current --namespace=ibm-common-services
```

e. Use `git clone`:

```bash
export operator_release_version=v1.3.3
git clone -b ${operator_release_version} https://github.com/IBM/ibm-licensing-operator.git
cd ibm-licensing-operator/
```

**Note:** If You cannot use `git clone` on machine with `kubectl` (for example, when you do not have the Internet connection), use the solution described in the troubleshooting section. See [Preparing resources for offline installation without git](Troubleshooting.md#preparing-resources-for-offline-installation-without-git). Then, see the Results underneath this step.

f. Apply RBAC roles and CRD:

```bash
# add CRD:
kubectl apply -f deploy/crds/operator.ibm.com_ibmlicensings_crd.yaml
kubectl apply -f deploy/crds/operator.ibm.com_ibmlicenseservicereporters_crd.yaml
# add RBAC:
kubectl apply -f deploy/role.yaml
kubectl apply -f deploy/service_account.yaml
kubectl apply -f deploy/role_binding.yaml
```

g. Modify the `operator.yaml` image so that your private registry is used:

- For **LINUX** users:

```bash
export operator_version=1.3.3
export operand_version=1.3.3
ESCAPED_REPLACE=$(echo ${my_docker_registry} | sed -e 's/[\/&]/\\&/g')
sed -i 's/quay\.io\/opencloudio/'"${ESCAPED_REPLACE}"'/g' deploy/operator.yaml
sed -i 's/operator@sha256.*/operator:'"${operator_version}"'/g' deploy/operator.yaml
sed -i 's/@sha256.*/:'"${operand_version}"'/g' deploy/operator.yaml
kubectl apply -f deploy/operator.yaml
```

- For **MAC** users:

```bash
export operator_version=1.3.3
export operand_version=1.3.3
ESCAPED_REPLACE=$(echo ${my_docker_registry} | sed -e 's/[\/&]/\\&/g')
sed -i "" 's/quay.io\/opencloudio/'"${ESCAPED_REPLACE}"'/g' deploy/operator.yaml
sed -i "" 's/operator@sha256.*/operator:'"${operator_version}"'/g' deploy/operator.yaml
sed -i "" 's/@sha256.*/:'"${operand_version}"'/g' deploy/operator.yaml
kubectl apply -f deploy/operator.yaml
```

**Results:**
You have created the **Operator** for **IBM Licensing Service**. The **Operator** is only responsible for watching over the configuration and managing resources used by **IBM Licensing Service**.

**What's Next:**
Create the IBM Licensing instance.

## Creating an IBM Licensing instance

1\. To create the the IBM Licensing instance, run the following command:

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

2\. If you created the secret that is needed to access the images, add it to the configuration.

```yaml
apiVersion: operator.ibm.com/v1alpha1
kind: IBMLicensing
metadata:
  name: instance
spec:
...
  imagePullSecrets:     # <-- this needs to be added
    - my-registry-token # <-- this needs to be added with your secret name
...
```

For example:

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

**Results:**
Installation is complete and **License Service** is running in your cluster.

<b>Related links</b>

- [Go back to home page](../License_Service_main.md#documentation)
- [Configuration](Configuration.md)
- [Retrieving license usage data from the cluster](Retrieving_data.md)
