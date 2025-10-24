package inputs

import (
	"fmt"
	"github.com/spf13/cobra"
)

var commandsCmd = &cobra.Command{
	Use:     "commands [session]",
	Aliases: []string{"cmds"},
	Short:   "등록된 명령 목록 보기",
	Long:    "지정한 세션에 등록된 명령어 목록을 표시합니다.",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sessionID := args[0]
		// TODO: 실제 명령 목록 조회 로직 구현
		fmt.Printf("세션 %s에 등록된 명령 목록:\n", sessionID)
		fmt.Println("---------------------------------------")
		// 예시 출력 (profile.yaml에서 읽어올 예정)
		fmt.Println("1. ssh limjihoon@127.0.0.1")
		fmt.Println("2. ls")
		fmt.Println("3. whoami")
		fmt.Println("4. neofetch")
	},
}
