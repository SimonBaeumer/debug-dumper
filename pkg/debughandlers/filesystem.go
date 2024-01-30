package debughandlers

import (
	"crypto/tls"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	// TargetPath can be mounted to a PVC
	TargetPath = "/diagnostics"
)

func DebugInfoGetterFileSystem(podMetrics v1beta1.PodMetrics, clientset *kubernetes.Clientset, metricsclient *metrics.Clientset) error {
	container := podMetrics.Containers[0]
	centralEndpoint := getCentralEndpoint(podMetrics)
	targetPath := fmt.Sprintf("%s/ns-%s/%s-%d.zip", TargetPath, podMetrics.GetNamespace(), podMetrics.GetName(), time.Now().Unix())

	err := os.MkdirAll(filepath.Dir(targetPath), os.ModePerm)
	if err != nil {
		return errors.Errorf("Could not create directory %s: %s", filepath.Dir(targetPath), err.Error())
	}

	fmt.Printf("%s/%s: Start download, memory usage %s\n", podMetrics.GetNamespace(), podMetrics.GetName(), prettyPrintQuantityToMB(container.Usage.Memory()))
	err = downloadFile(centralEndpoint, targetPath)
	if err != nil {
		return errors.Errorf("failed to download diagnostic info for: %s %s Error: %s", centralEndpoint, podMetrics.GetName(), err.Error())
	}
	fmt.Println("Finished downloading debug information to ", targetPath)

	return nil
}

func getCentralEndpoint(podMetrics v1beta1.PodMetrics) string {
	return fmt.Sprintf("https://central.%s.svc:9095/diagnostics", podMetrics.GetNamespace())
}

// downloadFile downloads a file from a URL with insecure TLS and saves it to a specified path.
func downloadFile(url, targetPath string) error {
	// Create a custom HTTP client with insecure TLS configuration
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func prettyPrintQuantityToMB(q *resource.Quantity) string {
	bytes := q.Value()
	megabytes := float64(bytes) / math.Pow(1024, 2)
	return fmt.Sprintf("%.2f MB", megabytes)
}
