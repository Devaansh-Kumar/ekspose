package main

import (
	_ "context"
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "location to your kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "location to your kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("error %s building config from flags", err.Error())
		config, err = rest.InClusterConfig()
		if err != nil {
			fmt.Printf("error %s getting in cluster config", err.Error())
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("error %s creating clientset", err.Error())
	}

	informer := informers.NewSharedInformerFactory(clientset, time.Minute*10)

	ch := make(chan struct{})
	c := newController(clientset, informer.Apps().V1().Deployments())
	informer.Start(ch)
	c.run(ch)

	fmt.Println(informer)
}
