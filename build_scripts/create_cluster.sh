#!/bin/bash
set -euo pipefail


unset KUBERNETES_PORT
unset KUBERNETES_PORT_443_TCP
unset KUBERNETES_PORT_443_TCP_ADDR
unset KUBERNETES_PORT_443_TCP_PORT
unset KUBERNETES_PORT_443_TCP_PROTO
unset KUBERNETES_SERVICE_HOST
unset KUBERNETES_SERVICE_PORT
unset KUBERNETES_SERVICE_PORT_HTTPS

echo "Starting kind-in-dind image"

# Clean up existing container if it exists
if docker ps -a --format '{{.Names}}' | grep -q '^kind-dind$'; then
  echo "Removing existing kind-dind container..."
  docker rm -f kind-dind || true
  sleep 15
fi

# Registry base
registry="docker-na-public.artifactory.swg-devops.com"
repo_path="hyc-cloud-private-scratch-docker-local/ibmcom"

# Image tags
kind_tag="licensing-kind:latest"
kindest_tag="licensing-kindest/node:v1.32.2"
registry_tag="licensing-registry:2"

# Full image paths
kind_image="$registry/$repo_path/$kind_tag"
kindest_image="$registry/$repo_path/$kindest_tag"
registry_image="$registry/$repo_path/$registry_tag"

# Docker run
docker run \
  -e CLUSTER_IMAGE="$kindest_image" \
  -e REGISTRY_IMAGE="$registry_image" \
  -e REGISTRY="$registry" \
  -e ARTIFACTORY_TOKEN="$ARTIFACTORY_TOKEN" \
  -e ARTIFACTORY_USERNAME="$ARTIFACTORY_USERNAME" \
  --privileged \
  -d \
  --name kind-dind \
  -p 61616:61616 \
  -p 5001:5001 \
  -p 2112:2112 \
  "$kind_image"
echo '::endgroup::'

#this sleep is needed as SPS is very slow in such env setups
sleep 90

MAX_RETRIES=50
RETRIES=0
until [ "$RETRIES" -ge "$MAX_RETRIES" ] || [[ "$(docker inspect --format='{{json .State.Health}}' kind-dind | jq -r '.Status')" == "healthy" ]]
do
  echo "Attempt $((++RETRIES)): kind in dind container not yet ready...$(docker inspect --format='{{json .State.Health}}' kind-dind | jq -r '.Status')"
  sleep 5
done

if [ "$RETRIES" -ge "$MAX_RETRIES" ]; then
  echo "kind in dind container failed to become ready, failing build"
  docker inspect --format='{{json .State.Health}}' kind-dind
  docker logs kind-dind
  exit 1
else
  echo "kind in dind container is reporting healthy status"
fi

echo "Cluster is now healthy, extracting config and testing connection"
mkdir -p ~/.kube
until eval "docker cp kind-dind:/root/.kube/config ~/.kube/config"
do
  sleep 1
done

kubectl version
kubectl cluster-info --context kind-kind

docker exec kind-dind ls -l

echo "=========================="
echo "Kubeconfig before any tests:"
cat ~/.kube/config
echo "Current context:"
kubectl config current-context
echo "Current namespace:"
kubectl config view --minify -o jsonpath='{..namespace}'
echo "=========================="

echo "Verifying cluster access..."

kubectl cluster-info --context kind-kind
kubectl get nodes

# INGRESS
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
kubectl rollout status deployment ingress-nginx-controller -n ingress-nginx

echo "Creating certificates inside Kind nodes..."
# Get the Kind node name
NODE_NAME=$(docker exec kind-dind kubectl get nodes -o jsonpath='{.items[0].metadata.name}')
# Create certificate directory inside the Kind node
docker exec kind-dind docker exec "$NODE_NAME" mkdir -p /var/folders/licensing/certs
# Generate certificates inside the Kind node
docker exec kind-dind docker exec "$NODE_NAME" sh -c "
  openssl req -nodes -new -x509 -subj '/C=/ST=/L=/O=/CN=' \
    -keyout /var/folders/licensing/certs/tls.key \
    -out /var/folders/licensing/certs/tls.crt
  chmod 664 /var/folders/licensing/certs/*
"
echo "Certificates created in Kind node filesystem"

echo "Cluster verification complete"

kubectl rollout status deployment ingress-nginx-controller  -n ingress-nginx
