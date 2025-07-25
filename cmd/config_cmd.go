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
		//		Step 1: Configuration File Setup
		//      "Let's start by setting up your configuration. What would you like to name your config.yaml file?"
		//	 - 어떤 프로젝트?
		//		Step 2: Project Selection
		//      "Which project are you configuring? This helps organize your connections by project scope."
		//	 - 어떤 서비스? (db, api-server, gitlab runner ,,,)
		//		Step 3: Service Definition
		//      "What type of service are you connecting to? (e.g., database, API server, GitLab runner, etc.)"
		//	 - 어떤 스테이지? (dev, prod ,,)
		//		Step 4: Environment Stage
		//      "Which environment stage is this for? (e.g., development, production, staging, etc.)"
		//	 - 어떤 명령어? (한 줄 한 줄 입력받되, 빈 줄 입력 시 명령어 입력 단계 종료)
		//		Step 5: Command Input
		//      "Now let's define the commands for this connection. Enter each command on a separate line. When you're finished, press Enter on an empty line to continue."
		//   - 입력한 명령어 최종 확인 (y -> yes 로 입력받아 다음 단계로 넘어감, n -> no 로 입력받아 명령어 다시 입력받게끔 처리)
		//		Step 6: Command Confirmation
		//      "Please review your commands below. Type 'y' to confirm and proceed, or 'n' to edit them again."
		// 	 - 글로벌 세팅 (재)설정
		//		Step 7: Global Settings
		//      "Finally, let's configure your global settings. These will apply across all your connections."

	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
