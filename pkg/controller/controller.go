package controller

import (
	"time"
	"fmt"
	"strings"
	"context"
	store "watch-dns/pkg/cache"
	"k8s.io/client-go/informers"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)


const  maxRetries = 5

// Event indicate the informerEvent
type Event struct {
	key          string
	eventType    string
	namespace    string
}

type Controller struct {
	logger *logrus.Entry
	clientSet kubernetes.Interface
	queue workqueue.RateLimitingInterface
	informer cache.SharedIndexInformer
	m *store.MapStore
	clusterDomain string
}


// query exit ingress with store
func FirstQueryIngressesResource(c *Controller, clusterDomain string)  {
	list, err := c.clientSet.ExtensionsV1beta1().Ingresses(v1.NamespaceAll).List(context.TODO(),metav1.ListOptions{})
	if err != nil{
		panic(err)
	}
	for _, data := range list.Items{
		hostname := data.Spec.Rules[0].Host
		if ! strings.HasSuffix(hostname, clusterDomain){
			c.m.Add(hostname, data.Status.LoadBalancer.Ingress[0].IP)
		}
	}
	//c.logger.Println("current cluster has host: ", c.m.List())
}

func NewController(clientset kubernetes.Interface, informer cache.SharedIndexInformer,cd string) *Controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	m := store.NewStore()
	var event Event
	informer.AddEventHandler(
		cache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool {
				ingObj, ok := obj.(*v1beta1.Ingress)
				if ! ok{
					return false
				}
				host := ingObj.Spec.Rules[0].Host
				if strings.HasSuffix(host,cd){
					return false
				}
				name := ingObj.Name
				if strings.Contains(name,"cm-acme"){
					return false
				}
				return true
			},
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					key, err := cache.MetaNamespaceKeyFunc(obj)
					event.key = key
					event.eventType = "create"
					if err == nil{
						queue.Add(event)
					}
				},
				UpdateFunc: func(oldObj, newObj interface{}) {
					key, err := cache.MetaNamespaceKeyFunc(newObj)
					event.key = key
					event.eventType = "update"

					if err == nil{
						queue.Add(event)
					}
				},
				DeleteFunc: func(obj interface{}) {
					logrus.Info("this version no delete action")
					//key, err := cache.MetaNamespaceKeyFunc(obj)
					//event.key = key
					//event.eventType = "delete"
					//if err == nil{
					//	queue.Add(event)
					//}
				},
			},
		},

	)
	return &Controller{
		logger: logrus.WithField("pkg","watch-dns"),
		clientSet: clientset,
		queue: queue,
		informer: informer,
		m: m,
		clusterDomain: cd,
	}
}

func (c *Controller) Run (stopCh chan struct{})  {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()
	c.logger.Info("start watch dns controller...")
	go c.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf(" timed out waiting for caches to sync ..."))
		return
	}
	go wait.Until(c.runWorker, time.Second, stopCh)
	<- stopCh
	c.logger.Info("Stop watch dns controller")
}

func (c *Controller)processNextItem() bool {

	obj, quit := c.queue.Get()
	if quit{
		return  false
	}
	defer c.queue.Done(obj)
	err := c.processItem(obj.(Event))
	if err == nil {
		c.queue.Forget(obj)
	}else if c.queue.NumRequeues(obj) < maxRetries {
		c.logger.Info("Error processing %s (will retry): %v", obj.(Event).key, err)
		c.queue.AddRateLimited(obj)
	}else {
		c.logger.Errorf("Error processing %s (giving up): %v", obj.(Event).key, err)
		c.queue.Forget(obj)
		runtime.HandleError(err)
	}
	return true

}

func (c *Controller) processItem (event Event) error {

	obj, exists, err := c.informer.GetIndexer().GetByKey(event.key)
	if err != nil {
		c.logger.Errorf("Fetching object with key %s from store failed with %v", event.key, err)
		return err
	}
	if !exists {
		c.logger.Info("Ingress %s does not exist anymore\n", event.key)
	}else {

		ingObj, ok := obj.(*v1beta1.Ingress)
		if ! ok{
			return fmt.Errorf("ingress obj conversion failure")
		}
		switch event.eventType {
		case "create":
			err := c.handleObject(ingObj)
			if err != nil{
				c.logger.Errorf(err.Error())
			}
		case "delete":
			fmt.Println("detail obj: ", ingObj.Name)
		case "update":
			err := c.handleObject(ingObj)
			if err != nil{
				c.logger.Errorf(err.Error())
			}
		}
	}
	return  nil
}

func (c *Controller) runWorker()  {
		for c.processNextItem() {
	}
}

func (c *Controller) filterExistingOrClusterDomain(obj interface{}) (string,string,bool)  {
	//parser empty interface obj to Ingress format
	ingObj, ok := obj.(*v1beta1.Ingress)
	if !ok{
		panic(ok)
	}
	host := ingObj.Spec.Rules[0].Host

	if len(ingObj.Status.LoadBalancer.Ingress) == 0{
		c.logger.Infof("host is %s, this is add action, but ingress no ip.", host)
		return host,"", false
	}
	// filter exist domain
	if ip, ok := c.m.Get(host); ok {
		// ing ip is equal store ip
		newip := ingObj.Status.LoadBalancer.Ingress[0].IP
		if newip == ip {
			c.logger.Infof("host is %s, ip is %s; this is first carried out, or change ingress but no change ingress ip", host,ip)
			return host,"", false
		}
		return host,newip,true
	}

	return host,ingObj.Status.LoadBalancer.Ingress[0].IP,true

}

func (c *Controller) handleObject(obj interface{}) error{
	hostDomain,ip,flag := c.filterExistingOrClusterDomain(obj)
	if ip == "" || !flag {
		return nil
	}
	c.logger.Infof("we will change dns record, host is %s, ip is %s", hostDomain,ip)
	hostList := strings.SplitN(hostDomain,".",2)
	ingDns := SelectClient(hostList[1])
	//  query dns server is not have rr record...
	_, flag = ingDns.QueryDns(hostList[0], hostList[1])
	// if dns sever has this RR record, just update record
	if flag {
		if ingDns.UpdateDns(hostList[0], hostList[1], ip) {
			c.m.Update(hostDomain,ip)
			c.logger.Infof("dns server has record, update dns record successful, host: %s, ip: %s", hostDomain, ip)
		}
		return nil
	} else {
		if ingDns.AddDns(hostList[0],ip, hostList[1]) {
			c.m.Add(hostDomain,ip)
			c.logger.Infof("add dns record successful, host: %s, ip: %s",hostDomain, ip)
		}
	}
	return nil
}






func WatchIngressMain(clients *kubernetes.Clientset, clusterDomain string)  {
	factory := informers.NewSharedInformerFactory(clients,0)

	ingressInformer := factory.Extensions().V1beta1().Ingresses().Informer()

	c := NewController(clients,ingressInformer, clusterDomain)
	// check cluster exist domain
	FirstQueryIngressesResource(c, clusterDomain)

	// Now let's start the controller
	stop := make(chan struct{})
	defer close(stop)
	go c.Run(stop)
	select {
	}
}