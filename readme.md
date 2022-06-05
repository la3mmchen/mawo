# mawo

mawo is just a little app that filters kubernetes pods by labels and print the configured resource requests and limits as json.

mawo can used ad cli or run inside a kubernetes cluster to be ask over http.

mawo consumes it's only setting via environment value:

* **MAWO_PORT** : set the port at which mawo will bind. if not set mawo takes port 80. mawo does bind to 0.0.0.0 and has no option to chose another interface.

## mawo-server

you can either build the binary directly on your system or build a container image.

### native

```bash
$ make build-server
all modules verified
wrote server to bin/mawo
MAWO_PORT=1245 $ ./bin/mawo
Starting server at port 1234
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


## mawo-cli

in case you do not want to run mawo in your cluster a littl cli can be built:

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