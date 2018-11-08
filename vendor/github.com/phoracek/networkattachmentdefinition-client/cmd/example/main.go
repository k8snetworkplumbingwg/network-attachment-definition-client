package main

import (
	"flag"
	"fmt"

	"github.com/golang/glog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	clientset "github.com/phoracek/networkattachmentdefinition-client/pkg/client/clientset/versioned"
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

	exampleClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building example clientset: %v", err)
	}

	// TODO: make it so it uses something like K8sCniCncfIo
	// TODO: maybe by nesting API package
	list, err := exampleClient.K8sV1().NetworkAttachmentDefinitions("default").List(metav1.ListOptions{})
	if err != nil {
		glog.Fatalf("Error listing all databases: %v", err)
	}

	for _, db := range list.Items {
		fmt.Printf("database %s with user %q\n", db.Name, db.Spec.Config)
	}
}
