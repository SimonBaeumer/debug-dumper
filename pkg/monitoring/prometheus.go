package monitoring

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type PrometheusQuery struct {
	Client v1.API
}

func NewPrometheusQuery(url string) (*PrometheusQuery, error) {
	client, err := api.NewClient(api.Config{
		Address: url,
	})
	if err != nil {
		return nil, err
	}

	return &PrometheusQuery{
		Client: v1.NewAPI(client),
	}, nil
}

func (pq *PrometheusQuery) QueryMetrics(query string) (model.Value, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, warnings, err := pq.Client.Query(ctx, query, time.Now())
	if err != nil {
		return nil, err
	}
	if len(warnings) > 0 {
		fmt.Println("Warnings:", warnings)
	}

	return result, nil
}

func (pq *PrometheusQuery) ProcessMetrics(query string) error {
	// Replace this with your logic
	metrics, err := pq.QueryMetrics(query)
	if err != nil {
		return err
	}

	// Process the metrics
	fmt.Println("Metrics:", metrics)

	return nil
}
