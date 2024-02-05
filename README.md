# Ekspose
Ekspose is a custom controller that creates a service and ingress resource corresponding to every deployment that is created in Kubernetes.

## Steps to run the project

Clone the repository
```bash
git clone https://github.com/Devaansh-Kumar/ekspose
```

Build the docker image
```bash
docker build -t devaanshk840/ekspose .
```

Start kind cluster using the given config and load the docker image
```bash
kind create cluster --config=kind-config.yaml
kind load docker-image devaanshk840/ekspose -n ekspose-cluster
```

Install ingress-nginx for ingress controller implementation in kind
```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
```

Deploy the controller
```bash
kubectl create namespace ekspose
kubectl create -f manifests/deployment.yaml
kubectl create -f manifests/service-account.yaml
kubectl create -f manifests/cluster-role.yaml
kubectl create -f manifests/cluster-role-binding.yaml
```

Now on creation of any deployment, a service and ingress should be created and deleted on deletion of deployment.