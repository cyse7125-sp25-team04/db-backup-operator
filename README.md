# test-crd
// TODO(user): Add simple overview of use/purpose

## Description
// TODO(user): An in-depth paragraph about your project and overview of use

## Getting Started

### Prerequisites
- go version v1.23.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### Building docker image for operator : 

#### Note:  RUN apk add --no-cache mysql-client docker command fails when building on linux machine. So removed the linux/arm64 platform build in Jenkinsfile. When building locally you can use below docker buildx command. 

```sh
docker buildx create --use
docker buildx build --platform linux/amd64,linux/arm64 -t csye712504/db-backup-operator:latest --push .
```

### Decrypt: 
```sh
sops -d key-denc.json > key.json

sops -d secrets-enc.yaml > secrets-denc.yaml
```

### Deploy GCP key as k8s secret to grant permissions to pod to interact with GCS bucket:

``` sh
kubectl create secret generic gcp-key --from-file=key.json -n backup
```


### NOTE: 
 - Helm first installs the CRDs, then will deploy the manifests