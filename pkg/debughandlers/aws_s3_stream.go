package debughandlers

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	BucketName = "diagnostics"
	AWSRegion  = "us-east-1"
	AWSKey     = ""
)

func DebugInfoS3Getter(podMetrics v1beta1.PodMetrics, clientset *kubernetes.Clientset, metricsclient *metrics.Clientset) error {
	return StreamToS3(getCentralEndpoint(podMetrics), BucketName, AWSKey)
}

// TODO: test it, only experimental currently
func StreamToS3(fileURL, bucketName, s3Key string) error {
	resp, err := http.Get(fileURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create a new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(AWSRegion)}, // TODO: configurable
	)
	if err != nil {
		return err
	}

	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess)

	// Upload the file to S3, streaming it directly from the HTTP response
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(s3Key),
		Body:   resp.Body,
	})
	return err
}
