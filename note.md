# 1. Without API
## Execute without docker container
go run main.go -kubeconfig=/path/to/kubeconfig -namespace=namespaceName -pvc=pvcName

## Execute with docker
docker run --rm -v /path/to/kubeconfig:/root/.kube/config imageName -namespace=namespaceName -pvc=pvcName

# 2. With API
## Execute without docker container
go run test/pvc_usage.go

## Execute with docker
docker run --rm -p 8080:8080 -v /path/to/kubeconfig:/root/.kube/config -e KUBECONFIG=/root/.kube/config balamaru/pvcusageby:3.0

 atau

docker run --rm -p 8080:8080 -v /path/to/kubeconfig:/root/.kube/config -e KUBECONFIG=/root/.kube/config -e KUBERNETES_API_SERVER=https://<kubernetesMasterIp>:6443 balamaru/pvcusageby:3.0


POST http://ipDockerServer:8080/pvc-usage-by

payload :
```json
{
  "namespace": "namespaceName",
  "pvcName": "pvcName"
}
```