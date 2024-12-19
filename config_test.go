package main

import (
	"flag"
	"os"
	"testing"
)

func TestLoadFromEnvVars(t *testing.T) {
	os.Setenv("AWS_DEFAULT_REGION", "us-west-2")
	os.Setenv("WATCH_DIR", "./watch")
	os.Setenv("S3_BUCKET_NAME", "my-s3-bucket")
	os.Setenv("S3_BUCKET_PREFIX", "my-prefix")

	config := NewConfiguration()
	config.Load()

	if config.watch_dir != "./watch" {
		t.Errorf("Expected watch_dir to be ./watch, got %s", config.watch_dir)
	}
	if config.bucket_name != "my-s3-bucket" {
		t.Errorf("Expected bucket_name to be my-s3-bucket, got %s", config.bucket_name)
	}
	if config.bucket_prefix != "my-prefix" {
		t.Errorf("Expected bucket_prefix to be my-prefix, got %s", config.bucket_prefix)
	}
}

func TestLoadFromFlags(t *testing.T) {
	os.Setenv("AWS_DEFAULT_REGION", "us-west-2")
	os.Setenv("WATCH_DIR", "./.vscode")
	os.Setenv("S3_BUCKET_NAME", "my-s3-bucket")
	os.Setenv("S3_BUCKET_PREFIX", "my-prefix")

	flag.Set("source", "./watch")
	flag.Set("bucket", "new-s3-bucket")
	flag.Set("prefix", "new-prefix")
	flag.Set("region", "eu-central-1")

	config := NewConfiguration()
	config.Load()

	if config.watch_dir != "./watch" {
		t.Errorf("Expected watch_dir to be ./watch, got %s", config.watch_dir)
	}
	if config.bucket_name != "new-s3-bucket" {
		t.Errorf("Expected bucket_name to be new-s3-bucket, got %s", config.bucket_name)
	}
	if config.bucket_prefix != "new-prefix" {
		t.Errorf("Expected bucket_prefix to be new-prefix, got %s", config.bucket_prefix)
	}
	if config.aws_region != "eu-central-1" {
		t.Errorf("Expected aws_region to be eu-central-1, got %s", config.aws_region)
	}
}

func TestLoadMissingEnvVarsAndFlags(t *testing.T) {
	os.Unsetenv("AWS_DEFAULT_REGION")
	os.Unsetenv("WATCH_DIR")
	os.Unsetenv("S3_BUCKET_NAME")
	os.Unsetenv("S3_BUCKET_PREFIX")

	config := NewConfiguration()
	config.Load()

	if config.watch_dir != "" {
		t.Errorf("Expected watch_dir to be empty, got %s", config.watch_dir)
	}
	if config.bucket_name != "" {
		t.Errorf("Expected bucket_name to be empty, got %s", config.bucket_name)
	}
	if config.bucket_prefix != "" {
		t.Errorf("Expected bucket_prefix to be empty, got %s", config.bucket_prefix)
	}

}

func TestValidateEnvVars(t *testing.T) {
	os.Setenv("AWS_DEFAULT_REGION", "us-west-2")
	os.Setenv("WATCH_DIR", "./.vscode")
	os.Setenv("S3_BUCKET_NAME", "my-s3-bucket")
	os.Setenv("S3_BUCKET_PREFIX", "my-prefix")

	config := NewConfiguration()
	config.Load()

	if success := config.Validate(); success != true {
		t.Errorf("Expected Validate to return true for valid configuration, got %t", success)
	}
}

func TestValidateFlags(t *testing.T) {
	os.Setenv("AWS_DEFAULT_REGION", "us-west-2")
	os.Setenv("WATCH_DIR", "./.vscode")
	os.Setenv("S3_BUCKET_NAME", "my-s3-bucket")
	os.Setenv("S3_BUCKET_PREFIX", "my-prefix")

	flag.Set("source", "./watch")
	flag.Set("bucket", "new-s3-bucket")
	flag.Set("prefix", "new-prefix")
	flag.Set("region", "eu-central-1")

	config := NewConfiguration()
	config.Load()

	if success := config.Validate(); success != true {
		t.Errorf("Expected Validate to return true for valid configuration, got %t", success)
	}
}

func TestValidateMissingEnvVarsAndFlags(t *testing.T) {
	os.Unsetenv("AWS_DEFAULT_REGION")
	os.Unsetenv("WATCH_DIR")
	os.Unsetenv("S3_BUCKET_NAME")
	os.Unsetenv("S3_BUCKET_PREFIX")

	config := NewConfiguration()
	config.Load()

	if success := config.Validate(); success != false {
		t.Errorf("Expected Validate to return false due to missing environment variables, got %t", success)
	}
}
