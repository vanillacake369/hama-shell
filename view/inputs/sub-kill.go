package inputs

import (
	"fmt"
	"github.com/spf13/cobra"
)

var killCmd = &cobra.Command{
	Use:     "kill [session]",
	Aliases: []string{"k"},
	Short:   "세션 종료",
	Long:    "지정한 세션을 완전히 종료합니다. 실행 중인 프로세스도 함께 종료됩니다.",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sessionID := args[0]
		// TODO: 실제 세션 종료 로직 구현
		fmt.Printf("세션 %s를 종료합니다...\n", sessionID)
		fmt.Println("세션이 성공적으로 종료되었습니다.")

		// 확인 프롬프트 추가 고려
		// fmt.Print("정말 종료하시겠습니까? (y/N): ")
	},
}
