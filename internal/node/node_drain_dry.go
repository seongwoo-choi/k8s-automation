package node

import (
	"app/model"
	"context"
	"log/slog"
	"os"
	"strings"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func NodeDrainDryRun(clientSet kubernetes.Interface, percentage string, usageType UsageType) ([]model.NodeDrainResult, error) {
	overNodes, err := GetNodeUsage(clientSet, percentage, usageType)
	if err != nil {
		slog.Error("노드 사용량을 가져오는 중 오류 발생", err)
		return nil, err
	}

	nodes, err := clientSet.CoreV1().Nodes().List(context.Background(), metaV1.ListOptions{})
	if err != nil {
		slog.Error("노드 목록을 가져오는 중 오류 발생", err)
		return nil, err
	}

	dryRunResults, err := handleDryRun(nodes, overNodes, percentage)
	if err != nil {
		slog.Error("Node Drain Dry Run 처리 중 오류 발생", err)
		return nil, err
	}

	return dryRunResults, nil
}

func handleDryRun(nodes *coreV1.NodeList, overNodes []model.NodeInfo, percentage string) ([]model.NodeDrainResult, error) {
	drainNodeLabels := strings.Split(os.Getenv("DRAIN_NODE_LABELS"), ",")
	for i, label := range drainNodeLabels {
		drainNodeLabels[i] = strings.TrimSpace(label)
	}
	slog.Info("node drain 에 사용할 노드 labels 은" + strings.Join(drainNodeLabels, ",") + " 입니다.")

	var dryRunResults []model.NodeDrainResult
	slog.Info("Dry run mode 실행")
	slog.Info("Memory 사용률이 기준 이하인 노드 개수", "percentage", percentage, "count", len(overNodes))

	for _, overNode := range overNodes {
		matchingNodes, err := findMatchingNodes(nodes, overNode, drainNodeLabels)
		if err != nil {
			return nil, err
		}
		dryRunResults = append(dryRunResults, matchingNodes...)
	}

	slog.Info("Memory 사용률이 기준 이하인 실제 Dry run 대상 노드 개수", "percentage", percentage, "count", len(dryRunResults))
	if len(dryRunResults) == 0 {
		slog.Warn("드레인 대상 노드가 없습니다")
	} else {
		for _, result := range dryRunResults {
			slog.Info("Dry Run 대상 노드",
				"nodeName", result.NodeName,
				"memoryUsage", result.Percentage)
		}
	}

	return dryRunResults, nil
}

func findMatchingNodes(nodes *coreV1.NodeList, overNode model.NodeInfo, drainNodeLabels []string) ([]model.NodeDrainResult, error) {
	var results []model.NodeDrainResult

	for _, node := range nodes.Items {
		provisionerName := node.Labels["karpenter.sh/nodepool"]
		if strings.Contains(node.Annotations["alpha.kubernetes.io/provided-node-ip"], overNode.NodeName) {
			for _, label := range drainNodeLabels {
				if strings.TrimSpace(provisionerName) == strings.TrimSpace(label) {
					results = append(results, model.NodeDrainResult{
						NodeName:        node.Name,
						InstanceType:    node.Labels["node.kubernetes.io/instance-type"],
						ProvisionerName: provisionerName,
						Percentage:      overNode.NodeUsage,
					})
				}
			}
		}
	}
	return results, nil
}
