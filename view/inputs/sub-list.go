package inputs

import (
	"fmt"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "현재 실행 중인 세션 보기",
	Long:    "실행 중인 모든 세션의 ID, 상태, 시작시간을 표시합니다.",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: 실제 세션 목록 조회 로직 구현
		fmt.Println("현재 실행 중인 세션 목록:")
		fmt.Println("ID\t상태\t\t시작시간")
		fmt.Println("---------------------------------------")
		// 예시 출력
		fmt.Println("1\tRunning\t\t2024-10-24 10:30:00")
		fmt.Println("2\tRunning\t\t2024-10-24 11:15:00")
		fmt.Println("3\tStopped\t\t2024-10-24 09:45:00")
	},
}
