package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type PVCUsageRequest struct {
	Namespace string `json:"namespace" binding:"required"`
	PVCName   string `json:"pvcName" binding:"required"`
}

type DeploymentInfo struct {
	Name    string `json:"name"`
	PodName string `json:"podName"`
}

func main() {
	r := gin.Default()

	r.POST("/pvc-usage-by", getPVCUsage)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Starting server on port %s\n", port)
	r.Run(":" + port)
}

func getPVCUsage(c *gin.Context) {
	var req PVCUsageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clientset, err := getKubernetesClientset()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create Kubernetes client: %v", err)})
		return
	}

	// Get PVC
	pvc, err := clientset.CoreV1().PersistentVolumeClaims(req.Namespace).Get(context.TODO(), req.PVCName, metav1.GetOptions{})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("PVC not found: %v", err)})
		return
	}

	// Get all pods in the namespace
	pods, err := clientset.CoreV1().Pods(req.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to list pods: %v", err)})
		return
	}

	deployments := []DeploymentInfo{}

	// Find pods using the PVC
	for _, pod := range pods.Items {
		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == pvc.Name {
				// Check if the pod is owned by a ReplicaSet
				for _, owner := range pod.OwnerReferences {
					if owner.Kind == "ReplicaSet" {
						rs, err := clientset.AppsV1().ReplicaSets(req.Namespace).Get(context.TODO(), owner.Name, metav1.GetOptions{})
						if err != nil {
							fmt.Printf("Error getting ReplicaSet: %v\n", err)
							continue
						}

						// Check if the ReplicaSet is owned by a Deployment
						for _, rsOwner := range rs.OwnerReferences {
							if rsOwner.Kind == "Deployment" {
								deployments = append(deployments, DeploymentInfo{
									Name:    rsOwner.Name,
									PodName: pod.Name,
								})
							}
						}
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"pvc":         req.PVCName,
		"namespace":   req.Namespace,
		"deployments": deployments,
	})
}

func getKubernetesClientset() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		// We're running inside a cluster, use the in-cluster config
		fmt.Println("Using in-cluster configuration")
		config, err = rest.InClusterConfig()
	} else {
		// We're running outside the cluster, use the kubeconfig file
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = "/root/.kube/config"
		}
		fmt.Printf("Using kubeconfig: %s\n", kubeconfig)
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create config: %v", err)
	}

	fmt.Printf("Config: %+v\n", config)

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %v", err)
	}

	return clientset, nil
}
