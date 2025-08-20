package executor

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

func TestExecutor_Run_Success(t *testing.T) {
	// GIVEN
	// Create a new executor instance
	exec := New()
	if exec == nil {
		t.Fatal("Expected executor to be created, got nil")
	}
	testKey := "test.service"
	var testCommand string
	if runtime.GOOS == "windows" {
		testCommand = "echo 'this is a test'"
	} else {
		testCommand = "echo 'this is a test'"
	}

	// WHEN
	// Test the Run method
	err := exec.Run(testKey, testCommand)
	if err != nil {
		t.Fatalf("Expected Run to succeed, got error: %v", err)
	}

	// THEN
	// Verify the process is running
	status := exec.GetStatus()
	for k, v := range status {
		fmt.Println("Key:", k, "Value:", v)
	}
	assert.Len(t, status, 1)
}

func TestExecutor_MultipleRun_Success(t *testing.T) {
	// GIVEN
	// Create a new executor instance
	exec := New()
	if exec == nil {
		t.Fatal("Expected executor to be created, got nil")
	}
	testKey := "test.service"
	var testCommand []string
	testCommand = append(testCommand, "echo 'this is a test1'")
	testCommand = append(testCommand, "echo 'this is a test2'")
	testCommand = append(testCommand, "echo 'this is a test3'")
	testCommand = append(testCommand, "echo 'this is a test4'")
	testCommand = append(testCommand, "echo 'this is a test5'")

	// WHEN
	// Test the Run method
	for _, command := range testCommand {
		err := exec.Run(testKey, command)
		if err != nil {
			t.Fatalf("Expected Run to succeed, got error: %v", err)
		}
	}

	// THEN
	// Verify the process is running
	status := exec.GetStatus()
	for k, v := range status {
		fmt.Println("Key:", k, "Value:", v)
	}
	assert.Len(t, status, 1)
}
