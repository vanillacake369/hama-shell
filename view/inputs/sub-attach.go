package inputs

import (
	"fmt"
	"github.com/spf13/cobra"
	"hama-shell/controller"
)

var attachCmd = &cobra.Command{
	Use:     "attach [session]",
	Aliases: []string{"a"},
	Short:   "실행 중인 프로세스의 TTY에 접속",
	Long:    "지정한 세션의 터미널에 접속합니다.",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sessionID := args[0]
		// TODO: 실제 TTY 접속 로직 구현
		fmt.Printf("세션 %s에 접속합니다...\n", sessionID)
		fmt.Println("(Ctrl+B+D를 눌러 세션에서 분리할 수 있습니다)")
		fmt.Println("**== 현재는 세션에 대한 구현이 안 되어 있어 바로 예시 명령어를 수행하는 PTY 로 접속합니다. ==**")
		controller.Attach()
	},
}
