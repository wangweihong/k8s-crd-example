package main

import (
	"flag"
	"fmt"

	"github.com/golang/glog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	"testcontroller/crd"
	examplecomclientset "testcontroller/pkg/client/clientset/versioned"

	apix "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
)

var (
	kuberconfig = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	master      = flag.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
)

func main() {
	flag.Parse()

	cfg, err := clientcmd.BuildConfigFromFlags(*master, *kuberconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %v", err)
	}

	clientset, err := apix.NewForConfig(cfg)
	if err != nil {
		panic(err)
	}
	err = crd.CreateCRD(clientset)
	if err != nil {
		panic(err)
	}

	exampleClient, err := examplecomclientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building example clientset: %v", err)
	}

	list, err := exampleClient.XfleetV1().Foos("default").List(metav1.ListOptions{})
	if err != nil {
		glog.Fatalf("Error listing all foos: %v", err)
	}

	for _, db := range list.Items {
		fmt.Printf("database %s with user %q\n", db.Name, db.Spec.DeploymentName)
	}
}
