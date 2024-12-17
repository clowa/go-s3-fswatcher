package main

import (
	"os"
	"os/exec"
	"testing"
)

func TestIntegrationValidSourceFlag(t *testing.T) {
	cmd := exec.Command("./go-s3-fswatcher", "--source", "/valid/path", "--bucket", "my-s3-bucket", "--prefix", "my-prefix")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Expected no error, got %v. Output: %s", err, output)
	}
}

func TestIntegrationInvalidSourceFlag(t *testing.T) {
	cmd := exec.Command("./go-s3-fswatcher", "--source", "/invalid/path", "--bucket", "my-s3-bucket", "--prefix", "my-prefix")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("Expected error, got none. Output: %s", output)
	}
	expectedError := "Invalid source directory. Please provide a valid directory path. Example: /path/to/source"
	if !contains(output, expectedError) {
		t.Fatalf("Expected error message: %s, got: %s", expectedError, output)
	}
}

func TestIntegrationValidBucketFlag(t *testing.T) {
	cmd := exec.Command("./go-s3-fswatcher", "--source", "/valid/path", "--bucket", "my-s3-bucket", "--prefix", "my-prefix")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Expected no error, got %v. Output: %s", err, output)
	}
}

func TestIntegrationInvalidBucketFlag(t *testing.T) {
	cmd := exec.Command("./go-s3-fswatcher", "--source", "/valid/path", "--bucket", "", "--prefix", "my-prefix")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("Expected error, got none. Output: %s", output)
	}
	expectedError := "Invalid S3 bucket name. Please provide a valid bucket name. Example: my-s3-bucket"
	if !contains(output, expectedError) {
		t.Fatalf("Expected error message: %s, got: %s", expectedError, output)
	}
}

func TestIntegrationValidPrefixFlag(t *testing.T) {
	cmd := exec.Command("./go-s3-fswatcher", "--source", "/valid/path", "--bucket", "my-s3-bucket", "--prefix", "my-prefix")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Expected no error, got %v. Output: %s", err, output)
	}
}

func TestIntegrationInvalidPrefixFlag(t *testing.T) {
	cmd := exec.Command("./go-s3-fswatcher", "--source", "/valid/path", "--bucket", "my-s3-bucket", "--prefix", "")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("Expected error, got none. Output: %s", output)
	}
	expectedError := "Invalid S3 prefix. Please provide a valid prefix. Example: my-prefix/"
	if !contains(output, expectedError) {
		t.Fatalf("Expected error message: %s, got: %s", expectedError, output)
	}
}

func contains(output []byte, expectedError string) bool {
	return string(output) == expectedError
}