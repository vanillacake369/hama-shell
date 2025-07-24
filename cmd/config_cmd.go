package cmd

import (
	"github.com/spf13/cobra"
)

// config.yaml 선언
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Read config file",
	Long:  "Read config file",
	Run: func(cmd *cobra.Command, args []string) {
		// ToDo : Config 선언
		//	 - 어떤 파일명으로 config.yaml 선언할 것인지?
		//	 - 어떤 프로젝트?
		//	 - 어떤 서비스? (db, api-server, gitlab runner ,,,)
		//	 - 어떤 스테이지? (dev, prod ,,)
		//	 - 어떤 명령어? (한 줄 한 줄 입력받되, 빈 줄 입력 시 명령어 입력 단계 종료)
		//   - 입력한 명령어 최종 확인 (y -> yes 로 입력받아 다음 단계로 넘어감, n -> no 로 입력받아 명령어 다시 입력받게끔 처리)
		// 	 - 글로벌 세팅 (재)설정
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
