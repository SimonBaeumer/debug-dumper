package main

import (
	"context"
	"fmt"
	"github.com/simonbaeumer/debug-monitor/pkg/debughandlers"
	"github.com/simonbaeumer/debug-monitor/pkg/monitoring"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var (
	MemoryThreshold = resource.MustParse("1Mi")
)

func main() {
	var kubeconfig string

	// Check if the code is running inside a Kubernetes cluster
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	// Use the in-cluster config if running inside a cluster,
	// otherwise, use the default kubeconfig location
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println("Not running in a cluster. Using default kubeconfig.")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}
	}

	// Create the Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Connected to kubernetes cluster", config.Host)

	// Create metrics clientset
	metricsClientset, err := metrics.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	ctx, cancelFn := context.WithCancel(context.Background())
	dumper := monitoring.Dumper{
		DebugInfoGetter:  debughandlers.DebugInfoGetterFileSystem,
		MetricsClientset: metricsClientset,
		Clientset:        clientset,
		MemoryThreshold:  MemoryThreshold,
		Interval:         30 * time.Second,
	}
	go dumper.Watch(ctx)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigs:
		cancelFn()
	}
}
