package node

import (
	"app/model"
	"context"
	"os"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// CreateMockNode creates a test node with given name, labels and annotations
func CreateMockNode(name string, labels, annotations map[string]string) *corev1.Node {
	return &corev1.Node{
		ObjectMeta: metaV1.ObjectMeta{
			Name:        name,
			Labels:      labels,
			Annotations: annotations,
		},
	}
}

func TestCordonNode(t *testing.T) {
	// 테스트용 클라이언트 생성
	client := fake.NewSimpleClientset()

	// 테스트용 노드 생성
	node := CreateMockNode("test-node", nil, nil)
	_, err := client.CoreV1().Nodes().Create(context.Background(), node, metaV1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create mock node: %v", err)
	}

	// 테스트 실행
	err = CordonNode(client, "test-node")
	if err != nil {
		t.Errorf("CordonNode failed: %v", err)
	}

	// 결과 확인
	updatedNode, err := client.CoreV1().Nodes().Get(context.Background(), "test-node", metaV1.GetOptions{})
	if err != nil {
		t.Errorf("Failed to get node: %v", err)
	}
	if !updatedNode.Spec.Unschedulable {
		t.Error("Node should be unschedulable after cordon")
	}
}

func TestCheckOverNode(t *testing.T) {
	// 테스트 설정
	clientset := fake.NewSimpleClientset()
	labels := map[string]string{os.Getenv("DRAIN_NODE_LABEL_KEY"): "test-pool"}
	annotations := map[string]string{"alpha.kubernetes.io/provided-node-ip": "192.168.1.1"}
	node := CreateMockNode("test-node", labels, annotations)
	clientset.CoreV1().Nodes().Create(context.Background(), node, metaV1.CreateOptions{})

	overNodes := []model.NodeInfo{{NodeName: "192.168.1.1", NodeUsage: 50.0}}
	drainNodeLabels := []string{"test-pool"}

	// 테스트 실행
	err := CheckOverNode(clientset, *node, overNodes, drainNodeLabels)
	if err != nil {
		t.Errorf("CheckOverNode failed: %v", err)
	}

	// 결과 확인
	updatedNode, err := clientset.CoreV1().Nodes().Get(context.Background(), "test-node", metaV1.GetOptions{})
	if err != nil {
		t.Errorf("Failed to get node: %v", err)
	}
	if !updatedNode.Spec.Unschedulable {
		t.Error("Node should be unschedulable after check")
	}
}
