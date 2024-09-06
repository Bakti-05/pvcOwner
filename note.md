# 1. Without API
# Execute without docker container
go run main.go -kubeconfig=/path/to/kubeconfig -namespace=namespaceName -pvc=pvcName

# Execute with docker
docker run --rm -v /path/to/kubeconfig:/root/.kube/config imageName -namespace=namespaceName -pvc=pvcName

# 2. With API
go run docs/pvc_usage.go

docker run --rm -p 8080:8080 -v /home/ubuntu/.kube/config:/root/.kube/config -e KUBECONFIG=/root/.kube/config balamaru/pvcusageby:3.0

 atau

docker run --rm -p 8080:8080 -v /home/ubuntu/.kube/config:/root/.kube/config -e KUBECONFIG=/root/.kube/config -e KUBERNETES_API_SERVER=https://103.179.33.244:6443 balamaru/pvcusageby:3.0


POST http://ipDockerServer:8080/pvc-usage-by

payload :
```json
{
  "namespace": "namespaceName",
  "pvcName": "pvcName"
}
```