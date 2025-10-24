package inputs

import (
	"fmt"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hs",
	Short: "Hama-Shell: 터미널 세션 관리 도구",
	Long:  "선언한 스크립트를 실행 및 관리하는 터미널 에뮬레이터입니다.",
	Run: func(cmd *cobra.Command, args []string) {
		// TUI 접속 (인자가 없을 때)
		fmt.Println("TUI 모드로 진입합니다...")
		// TODO: TUI 구현 연동
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	// 세션 관리 명령들
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(attachCmd)
	rootCmd.AddCommand(detachCmd)
	rootCmd.AddCommand(killCmd)
	rootCmd.AddCommand(commandsCmd)

	// 설정 관리 명령
	rootCmd.AddCommand(configCmd)
}
