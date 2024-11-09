package notification

import (
	"app/types"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendNodeDrainComplete(t *testing.T) {
	// 테스트 서버 설정
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// 테스트 데이터 설정
	os.Setenv("SLACK_WEBHOOK_URL", server.URL)
	os.Setenv("CLUSTER_NAME", "test-cluster")
	results := []types.NodeDrainResult{
		{
			NodeName:        "test-node-1",
			InstanceType:    "t3.medium",
			ProvisionerName: "test-pool",
			Percentage:      25.5,
		},
	}

	// 테스트 실행
	err := SendNodeDrainComplete(results)
	assert.NoError(t, err)
}

func TestSendNodeDrainError(t *testing.T) {
	// 테스트 서버 설정
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// 테스트 데이터 설정
	os.Setenv("SLACK_WEBHOOK_URL", server.URL)
	testError := fmt.Errorf("test error message")

	// 테스트 실행
	err := SendNodeDrainError(testError)
	assert.NoError(t, err)
}

func TestSendSlackMessageError(t *testing.T) {
	// 실패 케이스를 위한 테스트 서버
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// 테스트 실행
	err := sendSlackMessage(server.URL, "test message")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "slack notification failed with status code: 500")
}

func TestFormatNodeDrainMessage(t *testing.T) {
	t.Run("드레인된 노드가 있는 경우", func(t *testing.T) {
		results := []types.NodeDrainResult{
			{
				NodeName:        "test-node-1",
				InstanceType:    "t3.medium",
				ProvisionerName: "test-pool",
				Percentage:      25.5,
			},
			{
				NodeName:        "test-node-2",
				InstanceType:    "t3.large",
				ProvisionerName: "test-pool",
				Percentage:      15.3,
			},
		}

		message := formatNodeDrainMessage(results)

		assert.Contains(t, message, "🔄 노드 드레인 작업이 완료되었습니다")
		assert.Contains(t, message, "test-node-1")
		assert.Contains(t, message, "t3.medium")
		assert.Contains(t, message, "25.50%")
		assert.Contains(t, message, "test-node-2")
		assert.Contains(t, message, "t3.large")
		assert.Contains(t, message, "15.30%")
	})

	t.Run("드레인된 노드가 없는 경우", func(t *testing.T) {
		os.Setenv("CLUSTER_NAME", "test-cluster")
		results := []types.NodeDrainResult{}

		message := formatNodeDrainMessage(results)

		assert.Contains(t, message, "ℹ️ 드레인 대상 노드가 없습니다")
		assert.Contains(t, message, "test-cluster")
		assert.Contains(t, message, "메모리 사용률이 기준치보다 낮은 노드가 없거나 잘못된 정보를 입력하였습니다.")
	})
}
