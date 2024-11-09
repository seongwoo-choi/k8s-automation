package cmd

import (
	"app/config"
	"app/pkg/node"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
)

var (
	usageCmd = &cobra.Command{
		Use:   "usage",
		Short: "노드 사용량 조회",
	}

	memoryCmd = &cobra.Command{
		Use:   "memory",
		Short: "메모리 사용량 조회",
		Run: func(cmd *cobra.Command, args []string) {
			kubeConfig := os.Getenv("KUBE_CONFIG")
			clientSet, err := config.GetKubeClientSet(kubeConfig)
			if err != nil {
				slog.Error("쿠버네티스 클라이언트를 가져오는 중 오류가 발생했습니다.", err)
				return
			}

			handleNodeUsage(clientSet, percentage, node.MemoryUsage, "메모리")
		},
	}

	diskCmd = &cobra.Command{
		Use:   "disk",
		Short: "디스크 사용량 조회",
		Run: func(cmd *cobra.Command, args []string) {
			kubeConfig := "local"
			clientSet, err := config.GetKubeClientSet(kubeConfig)
			if err != nil {
				slog.Error("쿠버네티스 클라이언트를 가져오는 중 오류가 발생했습니다.", err)
				return
			}

			handleNodeUsage(clientSet, percentage, node.DiskUsage, "디스크")
		},
	}
)

func handleNodeUsage(clientSet kubernetes.Interface, percentage string, usageType node.UsageType, resourceType string) {
	fmt.Printf("노드 %s 사용량 조회 커맨드를 실행합니다.\n", resourceType)
	usage, err := node.GetNodeUsage(clientSet, percentage, usageType)
	if err != nil {
		slog.Error("노드 사용량 조회 중 오류가 발생했습니다.: ", err)
		return
	}
	fmt.Println("--------------------------------------------------------------------------------")
	for _, u := range usage {
		fmt.Printf("노드ID: %-60s 사용량: %6.2f%%\n", u.NodeName, u.NodeUsage)
	}
}

func init() {
	usageCmd.AddCommand(memoryCmd)
	usageCmd.AddCommand(diskCmd)
	rootCmd.AddCommand(usageCmd)
}
