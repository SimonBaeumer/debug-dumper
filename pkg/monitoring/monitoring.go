package monitoring

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
	"math"
	"time"
)

// todo: make it thread safe
var debugCache map[string]resource.Quantity

type Dumper struct {
	Clientset        *kubernetes.Clientset
	MetricsClientset *metrics.Clientset
	MemoryThreshold  resource.Quantity
	DebugInfoGetter  func(v1beta1.PodMetrics, *kubernetes.Clientset, *metrics.Clientset) error
	Interval         time.Duration
}

func init() {
	debugCache = make(map[string]resource.Quantity)
}

// Watch starts a watcher to poll pod metrics.
// Polling is necessary because the metrics server does not support a streaming API.
func (d *Dumper) Watch(ctx context.Context) {
	for {
		podMetrics, err := d.MetricsClientset.MetricsV1beta1().PodMetricses("").List(context.TODO(), metav1.ListOptions{
			// TODO: make it configurable to apply it to different components
			LabelSelector: "app.kubernetes.io/component=central",
		})
		if err != nil {
			fmt.Println("Error fetching pod metrics:", err)
			return
		}

		for _, podMetrics := range podMetrics.Items {
			memoryUsage := podMetrics.Containers[0].Usage.Memory()
			// memory usage exceeds threshold, take memory dump
			if memoryUsage.AsApproximateFloat64() > d.MemoryThreshold.AsApproximateFloat64() {
				usage := *podMetrics.Containers[0].Usage.Memory()
				cacheKey := getCacheKey(podMetrics)

				// Simple caching to only dump new debug info if memory had +/- 10% change.
				// If it has fallen below threshold does not take new dump.
				// TODO: Extend example with more capabilities to rate limit and configure debug dumps
				val, ok := debugCache[cacheKey]
				if ok {
					if math.Abs(val.AsApproximateFloat64()/usage.AsApproximateFloat64()*100) < 10 {
						fmt.Println("Skipping ", podMetrics.GetName(), " less than 10% memory difference")
						continue
					}
				}

				// Download debug info. Do it async.
				// TODO: Rate limit debug info getters
				// TODO: Make this call async and exclusive (allow only a single concurrent execution per Pod)
				err := d.DebugInfoGetter(podMetrics, d.Clientset, d.MetricsClientset)
				if err != nil {
					fmt.Println(err.Error())
				}

				debugCache[fmt.Sprintf("%s/%s", podMetrics.GetNamespace(), podMetrics.GetName())] = *podMetrics.Containers[0].Usage.Memory()
			}
		}

		select {
		case <-time.Tick(1 * time.Second):
			continue
		case <-ctx.Done(): // exit loop
			break
		}
	}
}

func getCacheKey(podMetrics v1beta1.PodMetrics) string {
	return fmt.Sprintf("%s/%s", podMetrics.GetNamespace(), podMetrics.GetName())
}
