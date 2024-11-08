package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	percentage          string
	prometheusAddress   string
	prometheusOrgID     string
	drainNodeLabelKey   string
	drainNodeLabelValue string
	slackWebhookURL     string
	kubeConfig          string
)

var rootCmd = &cobra.Command{
	Use:   "node-manager",
	Short: "노드 관리 CLI 도구",
	Long:  `노드의 메모리/디스크 사용량을 모니터링하고 드레인을 수행하는 CLI 도구입니다.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&percentage, "percentage", "p", "30", "사용량 퍼센테이지")
	rootCmd.PersistentFlags().StringVar(&prometheusAddress, "prometheus-address", "http://localhost:8080/prometheus", "Prometheus 서버 주소")
	rootCmd.PersistentFlags().StringVar(&prometheusOrgID, "prometheus-org-id", "organization-dev", "Prometheus 조직 ID")
	rootCmd.PersistentFlags().StringVar(&drainNodeLabelKey, "drain-node-label-key", "karpenter.sh/nodepool", "드레인 대상 노드의 Label Key")
	rootCmd.PersistentFlags().StringVar(&drainNodeLabelValue, "drain-node-label-value", "my-workload-label-valu1, my-workload-label-value2", "드레인 대상 노드 Label Value")
	rootCmd.PersistentFlags().StringVar(&slackWebhookURL, "slack-webhook-url", "", "Slack Webhook URL")
	rootCmd.PersistentFlags().StringVar(&kubeConfig, "kube-config", "local", "Kubernetes 설정 (local 또는 cluster)")
}

func initConfig() {
	// 환경 변수 설정
	os.Setenv("PERCENTAGE", percentage)
	os.Setenv("PROMETHEUS_ADDRESS", prometheusAddress)
	os.Setenv("PROMETHEUS_SCOPE_ORG_ID", prometheusOrgID)
	os.Setenv("DRAIN_NODE_LABEL_KEY", drainNodeLabelKey)
	os.Setenv("DRAIN_NODE_LABEL_VALUE", drainNodeLabelValue)
	os.Setenv("SLACK_WEBHOOK_URL", slackWebhookURL)
	os.Setenv("KUBE_CONFIG", kubeConfig)
}
