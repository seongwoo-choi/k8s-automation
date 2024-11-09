package node

import (
	"app/types"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CordonNodes(clientSet kubernetes.Interface, nodes *coreV1.NodeList, overNodes []types.NodeInfo, percentage string) error {
	drainNodeLabels := strings.Split(os.Getenv("DRAIN_NODE_LABELS"), ",")
	slog.Info("Memory 사용률이 기준 이하인 노드 개수", "percentage", percentage, "count", len(overNodes))
	for _, node := range nodes.Items {
		if err := CheckOverNode(clientSet, node, overNodes, drainNodeLabels); err != nil {
			return err
		}
	}
	return nil
}

func CheckOverNode(clientSet kubernetes.Interface, node coreV1.Node, overNodes []types.NodeInfo, drainNodeLabels []string) error {
	for _, overNode := range overNodes {
		provisionerName := node.Labels[os.Getenv("DRAIN_NODE_LABEL_KEY")]
		if strings.Contains(node.Annotations["alpha.kubernetes.io/provided-node-ip"], overNode.NodeName) {
			for _, label := range drainNodeLabels {
				if strings.TrimSpace(provisionerName) == strings.TrimSpace(label) {
					if err := CordonNode(clientSet, node.Name); err != nil {
						return fmt.Errorf("노드 %s를 cordon하는 데 실패했습니다: %w", node.Name, err)
					}
				}
			}
		}
	}
	return nil
}

func CordonNode(clientSet kubernetes.Interface, nodeName string) error {
	node, err := clientSet.CoreV1().Nodes().Get(context.Background(), nodeName, metaV1.GetOptions{})
	if err != nil {
		return err
	}

	// 이미 스케줄링 불가능 상태라면 스킵
	if node.Spec.Unschedulable {
		return nil
	}

	node.Spec.Unschedulable = true
	if _, err = clientSet.CoreV1().Nodes().Update(context.Background(), node, metaV1.UpdateOptions{}); err != nil {
		return err
	}
	slog.Info("노드 Cordon 완료", "nodeName", nodeName)

	return nil
}
