package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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

	r.Run(":8080")
}

func getPVCUsage(c *gin.Context) {
	var req PVCUsageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use the default kubeconfig path
	kubeconfig := "/home/ubuntu/.kube/configdtc"

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build kubeconfig"})
		return
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Kubernetes client"})
		return
	}

	// Get PVC
	pvc, err := clientset.CoreV1().PersistentVolumeClaims(req.Namespace).Get(context.TODO(), req.PVCName, metav1.GetOptions{})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PVC not found"})
		return
	}

	// Get all pods in the namespace
	pods, err := clientset.CoreV1().Pods(req.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list pods"})
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
