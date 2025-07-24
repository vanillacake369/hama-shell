package cmd

import (
	"github.com/spf13/cobra"
)

// 죽이기
var kill = &cobra.Command{
	Use:   "kill",
	Short: "Read config file",
	Long:  "Read config file",
	Run: func(cmd *cobra.Command, args []string) {
		// ToDo : Shell Job 삭제
	},
}

// 대시보드 보기
var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Read config file",
	Long:  "Read config file",
	Run: func(cmd *cobra.Command, args []string) {
		// ToDo : Shell Job 대시보드 보기
	},
}

// 프로세스의 실행 명령어 보기
var explainCmd = &cobra.Command{
	Use:   "explain",
	Short: "Read config file",
	Long:  "Read config file",
	Run: func(cmd *cobra.Command, args []string) {
		// ToDo : Shell Job 실행 중인 명령어 보기
	},
}

// 프로세스 실행
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Read config file",
	Long:  "Read config file",
	Run: func(cmd *cobra.Command, args []string) {
		// ToDo : Shell Job 실행
	},
}

func init() {
	rootCmd.AddCommand(kill)
	rootCmd.AddCommand(dashboardCmd)
	rootCmd.AddCommand(explainCmd)
	rootCmd.AddCommand(runCmd)
}
