package cmd

import (
	"app/config"
	"app/internal/node"
	"log/slog"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
)

var dryRunCmd = &cobra.Command{
	Use:   "dry-run",
	Short: "노드 드레인 드라이런 실행",
	Run: func(cmd *cobra.Command, args []string) {
		kubeConfig := "local"
		clientSet, err := config.GetKubeClientSet(kubeConfig)
		if err != nil {
			slog.Error("Failed to get kubernetes client", "error", err)
			return
		}

		handleDryRun(clientSet, percentage)
	},
}

func handleDryRun(clientSet *kubernetes.Clientset, percentage string) {
	slog.Info("노드 드레인 커맨드를 실행합니다.")
	_, err := node.NodeDrainDryRun(clientSet, percentage, node.MemoryUsage)
	if err != nil {
		slog.Error("노드 드레인 드라이 중 오류가 발생했습니다.: ", err)
	}
}

func init() {
	rootCmd.AddCommand(dryRunCmd)
}
