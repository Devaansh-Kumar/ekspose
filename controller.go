package main

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	appsinformer "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type contoller struct {
	clientset kubernetes.Interface

	// for listing deployments
	deploymentLister appslisters.DeploymentLister

	deploymentCacheSynced cache.InformerSynced
	workqueue             workqueue.RateLimitingInterface
}

func newController(clientset kubernetes.Interface, depInformer appsinformer.DeploymentInformer) *contoller {
	c := &contoller{
		clientset:             clientset,
		deploymentLister:      depInformer.Lister(),
		deploymentCacheSynced: depInformer.Informer().HasSynced,
		workqueue:             workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ekspose"),
	}

	depInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    handleAdd,
			DeleteFunc: handleDelete,
		},
	)

	return c
}

func (c *contoller) run(ch <-chan struct{}) {
	fmt.Println("Starting controller")
	// We need to wait for the informer cache to be synced or at least be initialized for the
	// first time.
	if !cache.WaitForCacheSync(ch, c.deploymentCacheSynced) {
		fmt.Printf("Waiting for cache to be synced:\n")
	}

	go wait.Until(c.worker, time.Second*1, ch)

	<-ch
}

func (c *contoller) worker() {

}

func handleAdd(obj interface{}) {
	fmt.Println("handleAdd")
}

func handleDelete(obj interface{}) {
	fmt.Println("handleDelete")
}
