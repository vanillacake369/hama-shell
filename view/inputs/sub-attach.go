package inputs

import (
	"fmt"
	"hama-shell/controller"
	"hama-shell/core"

	"github.com/spf13/cobra"
)

var attachCmd = &cobra.Command{
	Use:     "attach",
	Aliases: []string{"a"},
	Short:   "실행 중인 프로세스의 TTY에 접속",
	Long:    "지정한 세션의 터미널에 접속합니다.",
	Run: func(cmd *cobra.Command, args []string) {
		// flag 로 넘겨받은 프로젝트, 스테이지명
		project, _ := cmd.Flags().GetString("project")
		stage, _ := cmd.Flags().GetString("stage")

		fmt.Printf("프로젝트 %s 의 스테이지 %s 에 접속합니다...\n", project, stage)

		// TODO : "" 으로 선언된 magic value 를 어떻게 하면 피할 수 있지 ??
		config, err := core.LoadProfileConfig("profile.yaml")
		if err != nil {
			fmt.Printf("Failed to load profile: %v\n", err)
			return
		}

		// 프로젝트와 스테이지로 Stage 찾기
		stageData, err := config.GetStage(project, stage)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Stage의 모든 액션 가져오기
		commands := stageData.GetAllActions()

		// TODO : 어떻게 하면 PTY 안에서 실행된 것으로 인지될 수 있도록 TUI 안 별도의 창 안에서 처리될 수 있게끔 하지 ??
		controller.Attach(commands)
	},
}

func init() {
	// attach 명령어에 필수 flag 추가
	attachCmd.Flags().StringP("project", "p", "", "프로젝트 이름 (필수)")
	attachCmd.Flags().StringP("stage", "s", "dev", "스테이지 이름 (기본값: dev)")
	attachCmd.MarkFlagRequired("project")
}
