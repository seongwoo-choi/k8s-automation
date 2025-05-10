package node

import (
	"app/config"
	"app/types"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	prometheusModel "github.com/prometheus/common/model"
	"k8s.io/client-go/kubernetes"
)

type UsageType int

const (
	DiskUsage UsageType = iota
	MemoryUsage
)

func GetNodeUsage(clientSet kubernetes.Interface, percentage string, usageType UsageType) ([]types.NodeInfo, error) {
	query := buildQuery(percentage, usageType)

	prometheusClient, err := config.CreatePrometheusClient()
	if err != nil {
		slog.Error("Prometheus 클라이언트 생성 중 오류 발생", "error", err)
		return nil, err
	}

	result, err := config.QueryPrometheus(prometheusClient, query)
	if err != nil {
		slog.Error("Prometheus 쿼리 중 오류 발생", "error", err)
		return nil, err
	}

	return parseUsageResult(result), nil
}

func buildQuery(percentage string, usageType UsageType) string {
	switch usageType {
	case DiskUsage:
		return fmt.Sprintf("(1 - node_filesystem_avail_bytes / node_filesystem_size_bytes) * 100 > %s", percentage)
	case MemoryUsage:
		return fmt.Sprintf("100 * (1 - (node_memory_MemFree_bytes + node_memory_Cached_bytes + node_memory_Buffers_bytes) / node_memory_MemTotal_bytes) < %s", percentage)
	default:
		return ""
	}
}

func parseUsageResult(vector prometheusModel.Vector) []types.NodeInfo {
	var nodes []types.NodeInfo
	for _, sample := range vector {
		nodeName, usage := extractNodeUsage(sample)
		nodes = append(nodes, types.NodeInfo{
			NodeName:  nodeName,
			NodeUsage: usage,
		})
	}
	return nodes
}

func extractNodeUsage(sample *prometheusModel.Sample) (string, float64) {
	nodeName := string(sample.Metric["instance"])
	nodeName = nodeName[0:strings.Index(nodeName, ":")]
	usage, _ := strconv.ParseFloat(sample.Value.String(), 64)
	return nodeName, usage
}
