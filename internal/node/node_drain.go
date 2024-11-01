package node

import (
	"app/dao"
	"app/internal/pod"
	"context"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"
	"time"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func NodeDrain(clientSet *kubernetes.Clientset, percentage string, usageType UsageType) ([]dao.NodeDrainResult, error) {
	nodeUsages, err := GetNodeUsage(clientSet, percentage, usageType)
	if err != nil {
		return nil, err
	}

	nodes, err := clientSet.CoreV1().Nodes().List(context.Background(), metaV1.ListOptions{})
	if err != nil {
		return nil, err
	}

	if err := CordonNodes(clientSet, nodes, nodeUsages, percentage); err != nil {
		return nil, err
	}

	return handleDrain(clientSet, nodes, nodeUsages, percentage)
}

func handleDrain(clientSet *kubernetes.Clientset, nodes *coreV1.NodeList, overNodes []dao.NodeInfo, percentage string) ([]dao.NodeDrainResult, error) {
	drainNodeLabels := strings.Split(os.Getenv("DRAIN_NODE_LABELS"), ",")
	slog.Info("Node drain에 사용할 노드 labels", "labels", strings.Join(drainNodeLabels, ","))

	sort.Slice(overNodes, func(i, j int) bool {
		return overNodes[i].NodeUsage < overNodes[j].NodeUsage
	})

	var results []dao.NodeDrainResult
	for _, overNode := range overNodes {
		drainedNodes, err := drainMatchingNodes(clientSet, nodes, overNode, drainNodeLabels)
		if err != nil {
			return nil, err
		}
		results = append(results, drainedNodes...)
	}

	slog.Info("Memory 사용률이 기준 이하이며 실제로 Drain 대상인 노드 개수", "percentage", percentage, "count", len(results))

	return results, nil
}

func drainMatchingNodes(clientSet *kubernetes.Clientset, nodes *coreV1.NodeList, overNode dao.NodeInfo, drainNodeLabels []string) ([]dao.NodeDrainResult, error) {
	var results []dao.NodeDrainResult

	for _, node := range nodes.Items {
		provisionerName := node.Labels["karpenter.sh/nodepool"]
		if strings.Contains(node.Annotations["alpha.kubernetes.io/provided-node-ip"], overNode.NodeName) {
			for _, label := range drainNodeLabels {
				if strings.TrimSpace(provisionerName) == strings.TrimSpace(label) {
					if err := drainSingleNode(clientSet, node.Name); err != nil {
						return nil, err
					}
					results = append(results, dao.NodeDrainResult{
						NodeName:        node.Name,
						InstanceType:    node.Labels["beta.kubernetes.io/instance-type"],
						ProvisionerName: provisionerName,
						Percentage:      overNode.NodeUsage,
					})
				}
			}
		}
	}
	return results, nil
}

func drainSingleNode(clientSet *kubernetes.Clientset, nodeName string) error {
	if err := CordonNode(clientSet, nodeName); err != nil {
		return fmt.Errorf("%s 노드를 cordon 하는 중 오류가 발생했습니다.: %w", nodeName, err)
	}

	if err := pod.EvictPods(clientSet, nodeName); err != nil {
		return fmt.Errorf("노드 %s 에서 파드를 제거하는 중 오류가 발생했습니다.: %w", nodeName, err)
	}

	if err := waitForPodsToTerminate(clientSet, nodeName); err != nil {
		return fmt.Errorf("노드 %s 에서 파드가 종료되는 중 오류가 발생했습니다.: %w", nodeName, err)
	}

	time.Sleep(5 * time.Minute)

	return nil
}

func waitForPodsToTerminate(clientSet kubernetes.Interface, nodeName string) error {
	slog.Info("노드에서 데몬셋을 제외한 모든 파드가 종료될 때까지 기다리는 중", "nodeName", nodeName)

	_, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	for {
		pods, err := pod.GetNonCriticalPods(clientSet, nodeName)
		if err != nil {
			return fmt.Errorf("노드 %s 에서 데몬셋을 제외한 파드를 가져오는 중 오류가 발생했습니다.: %v", nodeName, err)
		}

		if len(pods) == 0 {
			slog.Info("데몬셋을 제외한 모든 Pod가 종료됨", "nodeName", nodeName)
			return nil
		}

		slog.Info("Pod 종료 대기 중", "nodeName", nodeName, "remainingPods", len(pods))
		time.Sleep(5 * time.Second)
	}
}
