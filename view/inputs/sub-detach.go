package inputs

import (
	"fmt"
	"github.com/spf13/cobra"
)

var detachCmd = &cobra.Command{
	Use:   "detach [session]",
	Short: "붙어있던 세션에서 빠져나오기",
	Long: `세션에서 분리합니다.

키 바인딩:
  Ctrl+B+D: 세션 분리 (tmux 스타일)

세션은 백그라운드에서 계속 실행됩니다.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sessionID := args[0]
		// TODO: 실제 세션 분리 로직 구현
		fmt.Printf("세션 %s에서 분리합니다...\n", sessionID)
		fmt.Println("세션은 백그라운드에서 계속 실행됩니다.")
	},
}
