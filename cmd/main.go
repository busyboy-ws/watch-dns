package main

import (
	"os"
	"flag"
	"path/filepath"
	utils "watch-dns/pkg/tools"
	ingress "watch-dns/pkg/controller"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/sirupsen/logrus"
)

func main() {
	// load current cluster domain, This is the domain name provided by the cloud vendor.
	clusterDomain := utils.GetClusterDomain()
	logrus.Println("current cluster domain: ", clusterDomain)
	clientSet := getClientSet()
	ingress.WatchIngressMain(clientSet,clusterDomain)
}

func getClientSet() *kubernetes.Clientset {
	if os.Getenv("PHASE") == "prod"{
			// creates the in-cluster config
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		// creates the clientset
		clientSet, err := kubernetes.NewForConfig(config)
		logrus.Printf("run in cluster ...")
		return clientSet
	}
	var kubeConfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeConfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeConfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeConfig)
	if err != nil {
		panic(err.Error())
	}
	clients, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	logrus.Printf("run in develop...")
	return clients
}