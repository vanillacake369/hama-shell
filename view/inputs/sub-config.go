package inputs

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "설정 파일 관리",
	Long:  "config 파일을 보기, 편집, 명령 추가합니다.",
}

var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "config 파일 내용 보기",
	Long:  "profile.yaml 파일의 내용을 표시합니다.",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: 실제 config 파일 경로 설정
		configPath := "profile.yaml"

		content, err := os.ReadFile(configPath)
		if err != nil {
			fmt.Printf("오류: config 파일을 읽을 수 없습니다: %v\n", err)
			return
		}

		fmt.Println("=== Config 파일 내용 ===")
		fmt.Println(string(content))
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "config 파일 편집/생성",
	Long:  "기본 편집기로 config 파일을 엽니다. 파일이 없으면 생성됩니다.",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: 실제 config 파일 경로 설정
		configPath := "profile.yaml"

		// 파일이 없으면 생성
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			file, err := os.Create(configPath)
			if err != nil {
				fmt.Printf("오류: config 파일을 생성할 수 없습니다: %v\n", err)
				return
			}
			file.Close()
			fmt.Println("새 config 파일이 생성되었습니다.")
		}

		// 환경변수에서 편집기 가져오기, 없으면 vi 사용
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}

		editCmd := exec.Command(editor, configPath)
		editCmd.Stdin = os.Stdin
		editCmd.Stdout = os.Stdout
		editCmd.Stderr = os.Stderr

		if err := editCmd.Run(); err != nil {
			fmt.Printf("오류: 편집기를 실행할 수 없습니다: %v\n", err)
		}
	},
}

var configAddCmd = &cobra.Command{
	Use:   "add",
	Short: "실행할 명령 등록",
	Long: `대화형 모드로 실행할 명령을 등록합니다.

키 바인딩:
  Alt+c: 명령어 입력 프로세스 취소
  Alt+f: 명령어 입력 완료 및 저장`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: 실제 대화형 입력 구현 (bubbletea 등 사용 고려)
		fmt.Println("명령 추가 모드 (대화형)")
		fmt.Println("---------------------------------------")
		fmt.Println("Alt+c: 취소")
		fmt.Println("Alt+f: 완료 및 저장")
		fmt.Println()
		fmt.Print("명령어를 입력하세요: ")

		// 간단한 예시 (실제로는 bubbletea로 구현 예정)
		var command string
		fmt.Scanln(&command)

		if command != "" {
			fmt.Printf("명령 '%s'가 추가되었습니다.\n", command)
			// TODO: profile.yaml에 저장하는 로직 구현
		} else {
			fmt.Println("명령 추가가 취소되었습니다.")
		}
	},
}

func init() {
	configCmd.AddCommand(configViewCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configAddCmd)
}
