package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "configdtc"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	namespace := flag.String("namespace", "default", "namespace of the PVC")
	pvcName := flag.String("pvc", "", "name of the PVC")
	flag.Parse()

	if *pvcName == "" {
		fmt.Println("Please provide a PVC name using the -pvc flag")
		os.Exit(1)
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Get PVC
	pvc, err := clientset.CoreV1().PersistentVolumeClaims(*namespace).Get(context.TODO(), *pvcName, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Searching for Deployments using PVC: %s\n", pvc.Name)

	// Get all pods in the namespace
	pods, err := clientset.CoreV1().Pods(*namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// Find pods using the PVC
	for _, pod := range pods.Items {
		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == pvc.Name {
				fmt.Printf("Pod %s is using the PVC\n", pod.Name)

				// Check if the pod is owned by a ReplicaSet
				for _, owner := range pod.OwnerReferences {
					if owner.Kind == "ReplicaSet" {
						rs, err := clientset.AppsV1().ReplicaSets(*namespace).Get(context.TODO(), owner.Name, metav1.GetOptions{})
						if err != nil {
							fmt.Printf("Error getting ReplicaSet: %v\n", err)
							continue
						}

						// Check if the ReplicaSet is owned by a Deployment
						for _, rsOwner := range rs.OwnerReferences {
							if rsOwner.Kind == "Deployment" {
								fmt.Printf("Deployment %s is using the PVC via Pod %s\n", rsOwner.Name, pod.Name)
							}
						}
					}
				}
			}
		}
	}
}
