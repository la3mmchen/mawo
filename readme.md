# mawo

mawo is just a little app that filters kubernetes pods by labels and print the configured resource requests and limits as json.

mawo can used as cli or run inside a kubernetes cluster to be ask over network.

mawo consumes it's only setting via an environment value:

* **MAWO_PORT** : set the port at which mawo will bind. if not set mawo takes port 80. mawo does bind to 0.0.0.0 and has no option to chose another interface.

## mawo-server

you can either build the binary directly on your system or build a container image.

### native

```bash
$ make build-server
all modules verified
wrote server to bin/mawo
MAWO_PORT=1245 $ ./bin/mawo
(..)
```

### container image

```bash
$ make build-image
docker build --build-arg app_version=a3e0e89 -t la3mmchen/mawo:a3e0e89 .
Sending build context to Docker daemon  117.7MB
Step 1/13 : FROM golang:alpine as builder
 ---> 155ead2e66ca
Step 2/13 : ARG app_version
 ---> Using cache
 ---> 4456744fcb0e
(...)
 ---> 0a5d5dc4c40e
Step 13/13 : CMD [""]
 ---> Running in a2a24e238850
Removing intermediate container a2a24e238850
 ---> 2dd8adf9cef8
Successfully built 2dd8adf9cef8
Successfully tagged la3mmchen/mawo:a3e0e89

docker tag la3mmchen/mawo:a3e0e89 la3mmchen/mawo:latest
```

if you're lazy just use the images on (ghcr.io)[https://github.com/la3mmchen/mawo/pkgs/container/mawo]

```bash
$ docker pull ghcr.io/la3mmchen/mawo:latest
latest: Pulling from la3mmchen/mawo
2408cc74d12b: Already exists
2cf7c15bfcf4: Pull complete
25a2a6b6e23b: Pull complete
Digest: sha256:ef5cca12ef71b9b71bd23e5fae25010f9fb2888d3731f8fabe4609abf64739f5
Status: Downloaded newer image for ghcr.io/la3mmchen/mawo:latest
ghcr.io/la3mmchen/mawo:latest

# will not run as i'm a macbook not a k8s cluster
$ docker run --rm -it  ghcr.io/la3mmchen/mawo:latest
2022/06/05 19:04:31 unable to load in-cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined
```

### helm it

there is a helm chart embedded that can be applied, maybe through predefined make target.

The helm chart contains a sound values.yaml, a deployment, a serviceaccount that is tied to a clusterrole that enables it to list pods, a service that is preset to type loadbalancer.

```bash
$ make deploy-latest
set -euxo pipefail &&\
        cd mawo-helm &&\
        helm lint . &&\
        helm template . &&\
        sed -i "" "s/0.0.0/latest/" "./Chart.yaml" &&\
        helm upgrade --namespace=mawo --create-namespace --install -f values.yaml mawo . &&\
        kubectl --namespace=mawo rollout restart deployment mawo &&\
        git checkout -- Chart.yaml
+ cd mawo-helm
+ helm lint .
==> Linting .
[INFO] Chart.yaml: icon is recommended

1 chart(s) linted, 0 chart(s) failed
+ helm template .
---
# Source: mawo/templates/serviceaccount.yaml
(...)
+ sed -i '' s/0.0.0/latest/ ./Chart.yaml
+ helm upgrade --namespace=mawo --create-namespace --install -f values.yaml mawo .
Release "mawo" has been upgraded. Happy Helming!
NAME: mawo
LAST DEPLOYED: ....
NAMESPACE: mawo
STATUS: deployed
REVISION: 12
NOTES:
la3mmchen/mawo
+ kubectl --namespace=mawo rollout restart deployment mawo
deployment.apps/mawo restarted
+ git checkout -- Chart.yaml
```

depending on your cluster setup you should now be able to talk to mawo, e.g.

```bash
λ ~/repos/github.com/la3mmchen/mawo/ main* curl -s "mawo.mawo.kube-cluster.local/container-resources?pod-label=tier%3Dcontrol-plane" | jq .
[
  {
    "container_name": "etcd",
    "cpu_limit": "",
    "cpu_req": "100m",
    "mem_limit": "",
    "mem_req": "100Mi",
    "namespace": "kube-system",
    "pod_name": "etcd-controlplane01-kube-cluster.local"
  },
  {
    "container_name": "kube-apiserver",
    "cpu_limit": "",
    "cpu_req": "250m",
    "mem_limit": "",
    "mem_req": "",
    "namespace": "kube-system",
    "pod_name": "kube-apiserver-controlplane01-kube-cluster.local"
  },
  {
    "container_name": "kube-controller-manager",
    "cpu_limit": "",
    "cpu_req": "200m",
    "mem_limit": "",
    "mem_req": "",
    "namespace": "kube-system",
    "pod_name": "kube-controller-manager-controlplane01-kube-cluster.local"
  },
  {
    "container_name": "kube-scheduler",
    "cpu_limit": "",
    "cpu_req": "100m",
    "mem_limit": "",
    "mem_req": "",
    "namespace": "kube-system",
    "pod_name": "kube-scheduler-controlplane01-kube-cluster.local"
  }
]
``

## mawo-cli

in case you do not want to run mawo in your cluster there is a little cli that can be built:

```bash
$ make build-cli
all modules verified
wrote binary to bin/mawo-cli
$ ./bin/mawo-cli tier=control-plane | jq .
[
  {
    "container_name": "etcd",
    "cpu_limit": "",
    "cpu_req": "100m",
    "mem_limit": "",
    "mem_req": "100Mi",
    "namespace": "kube-system",
    "pod_name": "etcd-minikube"
  },
  {
    "container_name": "kube-apiserver",
    "cpu_limit": "",
    "cpu_req": "250m",
    "mem_limit": "",
    "mem_req": "",
    "namespace": "kube-system",
    "pod_name": "kube-apiserver-minikube"
  },
  {
    "container_name": "kube-controller-manager",
    "cpu_limit": "",
    "cpu_req": "200m",
    "mem_limit": "",
    "mem_req": "",
    "namespace": "kube-system",
    "pod_name": "kube-controller-manager-minikube"
  },
  {
    "container_name": "kube-scheduler",
    "cpu_limit": "",
    "cpu_req": "100m",
    "mem_limit": "",
    "mem_req": "",
    "namespace": "kube-system",
    "pod_name": "kube-scheduler-minikube"
  }
]
```

enjoy.