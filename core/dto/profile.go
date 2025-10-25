package dto

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Profile 프로파일 구조체 - profile.yaml 파싱용
type Profile struct {
	Name    string `yaml:"name"`    // 프로파일 이름
	Project string `yaml:"project"` // 프로젝트 이름
	Stage   string `yaml:"stage"`   // 스테이지 (dev, staging, prod 등)
	Steps   []Step `yaml:"steps"`   // 실행 단계 목록
}

// Step 실행 단계
type Step struct {
	Type    string   `yaml:"type"`    // 단계 타입 (test, deploy 등)
	Actions []string `yaml:"actions"` // 실행할 명령어 목록
}

// ProfileConfig 전체 프로파일 설정 (profile.yaml 최상위)
type ProfileConfig struct {
	Profiles []Profile `yaml:"profiles"` // 프로파일 목록
}

// LoadProfileConfig YAML 파일에서 프로파일 설정을 로드
func LoadProfileConfig(filePath string) (*ProfileConfig, error) {
	// 파일 읽기
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("파일 읽기 실패: %w", err)
	}

	// YAML 파싱
	var config ProfileConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("YAML 파싱 실패: %w", err)
	}

	return &config, nil
}

// FindProfile 이름으로 프로파일 찾기
func (c *ProfileConfig) FindProfile(name string) (*Profile, error) {
	for i := range c.Profiles {
		if c.Profiles[i].Name == name {
			return &c.Profiles[i], nil
		}
	}
	return nil, fmt.Errorf("프로파일 '%s'를 찾을 수 없습니다", name)
}

// GetAllActions 프로파일의 모든 액션을 순서대로 반환
func (p *Profile) GetAllActions() []string {
	var actions []string
	for _, step := range p.Steps {
		actions = append(actions, step.Actions...)
	}
	return actions
}

// GetActionsByStepType 특정 타입의 스텝에서 액션 추출
func (p *Profile) GetActionsByStepType(stepType string) []string {
	var actions []string
	for _, step := range p.Steps {
		if step.Type == stepType {
			actions = append(actions, step.Actions...)
		}
	}
	return actions
}
