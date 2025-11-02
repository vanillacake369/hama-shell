package core

import (
	"fmt"
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/samber/lo"
	"github.com/spf13/viper"
)

// Step 스테이지 내의 단일 실행 단계를 나타냄
type Step struct {
	Type    string   `yaml:"type"`    // 단계 타입 (test-01, deploy, build 등)
	Actions []string `yaml:"actions"` // 실행할 셸 명령어 목록
}

// Stage 환경 스테이지를 나타냄 (dev, prod, staging 등)
type Stage struct {
	Steps []Step `yaml:"steps"` // 순서대로 실행할 단계 목록
}

// Project 여러 스테이지를 가진 프로젝트를 나타냄
// YAML 구조가 project-name -> stage-name -> Stage 이므로 wrapper 구조체 없이 map을 직접 사용
type Project map[string]Stage

// ProfileConfig 전체 profile.yaml 구조를 나타냄
type ProfileConfig struct {
	Profiles map[string]Project `yaml:"profiles"` // 프로젝트 이름을 Project로 매핑
}

// LoadProfileConfig YAML 파일에서 프로파일 설정을 로드. viper 를 사용
func LoadProfileConfig(filePath string) (*ProfileConfig, error) {
	// Viper 설정
	viper.SetConfigFile(filePath)
	viper.SetConfigType("yaml")

	// 설정 파일 읽기
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("파일 읽기 실패: %w", err)
	}

	// 구조체로 언마샬
	var config ProfileConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("YAML 파싱 실패: %w", err)
	}

	return &config, nil
}

// GetStage 프로젝트에서 특정 스테이지를 조회
func (c *ProfileConfig) GetStage(projectName, stageName string) (*Stage, error) {
	project, exists := c.Profiles[projectName]
	if !exists {
		return nil, fmt.Errorf("프로젝트 '%s'를 찾을 수 없습니다", projectName)
	}

	stage, exists := project[stageName]
	if !exists {
		return nil, fmt.Errorf("프로젝트 '%s'에서 스테이지 '%s'를 찾을 수 없습니다", projectName, stageName)
	}

	return &stage, nil
}

// ToString ProfileConfig를 포맷팅된 문자열로 반환
func (c *ProfileConfig) ToString() string {
	var result string
	result += "========================================\n"
	result += "Profile Configuration\n"
	result += "========================================\n\n"

	profileCount := 0
	for projectName, project := range c.Profiles {
		for stageName, stage := range project {
			profileCount++
			result += fmt.Sprintf("[Profile %d]\n", profileCount)
			result += fmt.Sprintf("  Project: %s\n", projectName)
			result += fmt.Sprintf("  Stage:   %s\n", stageName)
			result += "  Steps:\n"

			for i, step := range stage.Steps {
				result += fmt.Sprintf("    [Step %d] Type: %s\n", i+1, step.Type)
				result += "      Actions:\n"
				for j, action := range step.Actions {
					result += fmt.Sprintf("        %d. %s\n", j+1, action)
				}
			}
			result += "\n"
		}
	}

	result += "========================================\n"
	result += fmt.Sprintf("Total Profiles: %d\n", profileCount)
	result += "========================================\n"

	return result
}

// GetAllActions FlatMap을 사용하여 모든 단계의 액션을 순서대로 반환
func (s *Stage) GetAllActions() []string {
	return lo.FlatMap(s.Steps, func(step Step, _ int) []string {
		return step.Actions
	})
}

// SyncProfile profile.yaml 파일 변경을 감지하고 동기화 (viper 사용)
func SyncProfile() {
	fmt.Println("Starting profile.yaml watcher...")
	fmt.Println("Press Ctrl+C to exit")
	fmt.Println()

	// Viper 설정
	viper.SetConfigName("profile")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// 초기 설정 읽기
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	// 초기 프로파일 출력
	printProfile()

	// 설정 변경 핸들러 설정
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("\n[%s] Config file changed: %s\n", time.Now().Format("15:04:05"), e.Name)
		fmt.Printf("Event operation: %s\n\n", e.Op.String())

		// 업데이트된 프로파일 출력
		printProfile()
	})

	// 설정 파일 감시 시작
	viper.WatchConfig()

	// 프로그램 계속 실행
	select {}
}

func printProfile() {
	// LoadProfileConfig를 사용하여 로드
	config, err := LoadProfileConfig("profile.yaml")
	if err != nil {
		log.Printf("Error loading profile: %v", err)
		return
	}

	fmt.Println("========================================")
	fmt.Println("Profile Configuration:")
	fmt.Println("========================================")

	profileCount := 0
	for projectName, project := range config.Profiles {
		for stageName, stage := range project {
			profileCount++
			fmt.Printf("\n[Profile %d]\n", profileCount)
			fmt.Printf("  Project: %s\n", projectName)
			fmt.Printf("  Stage:   %s\n", stageName)
			fmt.Printf("  Steps:\n")

			for j, step := range stage.Steps {
				fmt.Printf("    [Step %d] Type: %s\n", j+1, step.Type)
				fmt.Printf("      Actions:\n")
				for k, action := range step.Actions {
					fmt.Printf("        %d. %s\n", k+1, action)
				}
			}
		}
	}

	fmt.Println("\n========================================")
	fmt.Printf("Total Profiles: %d\n", profileCount)
	fmt.Println("========================================")
}
