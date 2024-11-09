package notification

import (
	"app/types"
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

type SlackMessage struct {
	Text string `json:"text"`
}

func SendNodeDrainComplete(results []types.NodeDrainResult) error {
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		return fmt.Errorf("SLACK_WEBHOOK_URL is not set")
	}

	message := formatNodeDrainMessage(results)
	return sendSlackMessage(webhookURL, message)
}

func SendNodeDrainError(err error) error {
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		return fmt.Errorf("SLACK_WEBHOOK_URL is not set")
	}
	return sendSlackMessage(webhookURL, err.Error())
}

func formatNodeDrainMessage(results []types.NodeDrainResult) string {
	if len(results) == 0 {
		return fmt.Sprintf("ℹ️ 드레인 대상 노드가 없습니다 (클러스터: %s)\n\n메모리 사용률이 기준치보다 낮은 노드가 없거나 잘못된 정보를 입력하였습니다.",
			os.Getenv("CLUSTER_NAME"))
	}

	var message string
	message = fmt.Sprintf("🔄 노드 드레인 작업이 완료되었습니다 (클러스터: %s)\n\n", os.Getenv("CLUSTER_NAME"))

	for _, result := range results {
		message += fmt.Sprintf("• 노드: %s\n  인스턴스 타입: %s\n  메모리 사용률: %.2f%%\n\n",
			result.NodeName,
			result.InstanceType,
			result.Percentage)
	}

	return message
}

func sendSlackMessage(webhookURL string, message string) error {
	payload := SlackMessage{Text: message}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack notification failed with status code: %d", resp.StatusCode)
	}

	slog.Info("Slack 알림 전송 완료")
	return nil
}
