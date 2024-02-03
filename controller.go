package main

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			AddFunc:    c.handleAdd,
			DeleteFunc: c.handleDelete,
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
	for c.processItem() {

	}
}

func (c *contoller) processItem() bool {
	item, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}
	defer c.workqueue.Forget(item)

	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		fmt.Printf("error getting key from cache: %s\n", err.Error())
	}

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		// TODO: implement retry logic here
		fmt.Printf("error splitting key into namespace and name: %s\n", err.Error())
	}

	err = c.syncDeployment(ns, name)
	if err != nil {
		fmt.Printf("error syncing deployments\n: %s", err.Error())
		return false
	}

	return true
}

// Used to create service and ingress for a deployment
func (c *contoller) syncDeployment(ns, name string) error {
	ctx := context.Background()

	dep, err := c.deploymentLister.Deployments(ns).Get(name)
	if err != nil {
		fmt.Printf("error getting deployment name: %s\n", err.Error())
	}

	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dep.Name,
			Namespace: ns,
		},
		Spec: corev1.ServiceSpec{
			Selector: depLabels(*dep),
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name: "http",
					Port: 80,
				},
			},
		},
	}
	_, err = c.clientset.CoreV1().Services(ns).Create(ctx, &svc, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("error creating service: %s\n", err.Error())
	}
	return nil
}

func (c *contoller) handleAdd(obj interface{}) {
	fmt.Println("handleAdd")
	c.workqueue.Add(obj)
}

func (c *contoller) handleDelete(obj interface{}) {
	fmt.Println("handleDelete")
}

func depLabels(dep appsv1.Deployment) map[string]string {
	return dep.Spec.Template.Labels
}
