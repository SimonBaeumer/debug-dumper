package utils

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"
	"math"
)

func PrettyPrintQuantityToMB(q *resource.Quantity) string {
	bytes := q.Value()
	megabytes := float64(bytes) / math.Pow(1024, 2)
	return fmt.Sprintf("%.2f MB", megabytes)
}
