# Offline installation

## Prerequisites
- A private Docker image registry where you can push the images using `Docker` and from where your cluster can pull images.
- Machine with access to your cluster with `kubectl` command.

## Installation

1\. **Prepare Docker images**

Prepare your Docker images:

```bash
# on machine with access to internet
export my_docker_registry=<YOUR REGISTRY IMAGE PREFIX HERE e.g.: "my.registry:5000" or "quay.io/opencloudio">
export operator_version=1.1.3
export operand_version=1.1.2

# pull needed images
docker pull quay.io/opencloudio/ibm-licensing-operator:${operator_version}
docker pull quay.io/opencloudio/ibm-licensing:${operand_version}

# tag them with your registry prefix and push
docker tag quay.io/opencloudio/ibm-licensing-operator:${operator_version} ${my_docker_registry}/ibm-licensing-operator:${operator_version}
docker push ${my_docker_registry}/ibm-licensing-operator:${operator_version}

docker tag quay.io/opencloudio/ibm-licensing:${operand_version} ${my_docker_registry}/ibm-licensing:${operand_version}
docker push ${my_docker_registry}/ibm-licensing:${operand_version}
```

2\. **Create needed resources**

a. Run the following command on machine where you have access to your cluster and can use `kubectl`.

```bash
# on machine with access to cluster
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
export operator_release_version=v1.1.3-durham
git clone -b ${operator_release_version} https://github.com/IBM/ibm-licensing-operator.git
cd ibm-licensing-operator/
```

**Note:** If You cannot use `git clone` on machine with `kubectl` (for example, when you do not have the Internet connection), use the solution described in the troubleshooting section. See [Preparing resources for offline installation without git](Troubleshooting.md#preparing-resources-for-offline-installation-without-git). Then, see the Results underneath this step.

f. Apply RBAC roles and CRD:

```bash
# add CRD:
kubectl apply -f deploy/crds/operator.ibm.com_ibmlicensings_crd.yaml
# add RBAC:
kubectl apply -f deploy/role.yaml
kubectl apply -f deploy/service_account.yaml
kubectl apply -f deploy/role_binding.yaml
```

g. Modify the `operator.yaml` image so that your private registry is used:

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

**Results:**
You have created the **Operator** for **IBM Licensing Service**. The **Operator** is only responsible for watching over the configuration and managing resources used by **IBM Licensing Service**.

**What's Next:**
Configure the IBM Licensing instance.

## Creating an IBM Licensing instance

**Important:** The minimal setup requires applying this IBMLicensing instance. However, before applying the instance, get familiar with the entire configuration process.

1. To create the the IBM Licensing instance, run the following command:

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

2. If you created the secret that is needed to access the images, add it to the configuration.

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

**Related links**

- [Go back to home page](../License_Service_main.md#documentation)
- [Configuration](Configuration.md)
- [Retrieving license usage data from the cluster](Retrieving_data.md)