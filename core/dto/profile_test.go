package dto

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestSomething(t *testing.T) {
	pwd, _ := os.Getwd()
	file := pwd + "/../../profile.yaml"
	f, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	// ProfileConfig로 언마샬 (최상위가 profiles 이므로)
	var config ProfileConfig

	// Unmarshal our input YAML file into ProfileConfig
	if err = yaml.Unmarshal(f, &config); err != nil {
		log.Fatal(err)
	}

	// Print out the new struct
	fmt.Printf("%+v\n", config)

	// 검증
	assert.NotEmpty(t, config.Profiles, "프로파일 목록이 비어있으면 안됨")

	if len(config.Profiles) > 0 {
		profile := config.Profiles[0]
		fmt.Printf("First Profile: %+v\n", profile)
		assert.NotEmpty(t, profile.Name, "프로파일 이름이 비어있으면 안됨")
		assert.Equal(t, "project-a-dev", profile.Name)
		assert.Equal(t, "A", profile.Project)
		assert.Equal(t, "dev", profile.Stage)
	}
}
