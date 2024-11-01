package cmd

import (
	"app/config"
	"app/internal/node"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
)

var drainCmd = &cobra.Command{
	Use:   "drain",
	Short: "노드 드레인 실행",
	Run: func(cmd *cobra.Command, args []string) {
		kubeConfig := os.Getenv("KUBE_CONFIG")
		clientSet, err := config.GetKubeClientSet(kubeConfig)
		if err != nil {
			slog.Error("Failed to get kubernetes client", "error", err)
			return
		}

		handleNodeDrain(clientSet, percentage)
	},
}

func handleNodeDrain(clientSet *kubernetes.Clientset, percentage string) {
	slog.Info("노드 드레인 커맨드를 실행합니다.")

	_, err := node.NodeDrain(clientSet, percentage, node.MemoryUsage)
	if err != nil {
		slog.Error("Error during node drain: ", err)
		return
	}
	// slack 알람 받기
}

func init() {
	rootCmd.AddCommand(drainCmd)
}
