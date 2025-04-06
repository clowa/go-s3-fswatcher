package main

import (
	"flag"
	"os"
	"testing"
)

// Only works if the flags don't have a default value.
// Otherwise you need to reset the flag to it's default value.
func restFlags() {
	flag.Set("source", "")
	flag.Set("bucket", "")
	flag.Set("prefix", "")
	flag.Set("region", "")
}

// unsetEnvVars unsets all environment variables respected by the configuration.
// This is useful in tests to ensure that the environment variables are not set.
func unsetEnvVars() {
	os.Unsetenv("AWS_REGION")
	os.Unsetenv("AWS_DEFAULT_REGION")
	os.Unsetenv("WATCH_DIR")
	os.Unsetenv("S3_BUCKET_NAME")
	os.Unsetenv("S3_BUCKET_PREFIX")
}

func TestLoadFromEnvVars(t *testing.T) {
	os.Setenv("AWS_DEFAULT_REGION", "us-west-1")
	os.Setenv("WATCH_DIR", "./watch")
	os.Setenv("S3_BUCKET_NAME", "my-s3-bucket")
	os.Setenv("S3_BUCKET_PREFIX", "my-prefix")
	defer unsetEnvVars()

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
	if config.aws_region != "us-west-1" {
		t.Errorf("Expected aws_region to be us-west-1, got %s", config.aws_region)
	}
}

func TestLoadFromEnvVarsWithAwsRegion(t *testing.T) {
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	os.Setenv("AWS_REGION", "us-east-2")
	os.Setenv("WATCH_DIR", "./watch")
	os.Setenv("S3_BUCKET_NAME", "my-s3-bucket")
	os.Setenv("S3_BUCKET_PREFIX", "my-prefix")
	defer unsetEnvVars()

	config := NewConfiguration()
	config.Load()

	if config.aws_region != "us-east-2" {
		t.Errorf("Expected aws_region to be us-east-2, got %s", config.aws_region)
	}

}

func TestLoadFromFlags(t *testing.T) {
	os.Setenv("AWS_DEFAULT_REGION", "us-west-1")
	os.Setenv("WATCH_DIR", "./.vscode")
	os.Setenv("S3_BUCKET_NAME", "my-s3-bucket")
	os.Setenv("S3_BUCKET_PREFIX", "my-prefix")
	defer unsetEnvVars()

	flag.Set("source", "./watch")
	flag.Set("bucket", "new-s3-bucket")
	flag.Set("prefix", "new-prefix")
	flag.Set("region", "eu-central-1")
	defer restFlags()

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
	unsetEnvVars()

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
	if config.aws_region != "" {
		t.Errorf("Expected aws_region to be empty, got %s", config.aws_region)
	}
}

func TestValidateEnvVars(t *testing.T) {
	os.Setenv("AWS_DEFAULT_REGION", "us-west-2")
	os.Setenv("WATCH_DIR", "./.vscode")
	os.Setenv("S3_BUCKET_NAME", "my-s3-bucket")
	os.Setenv("S3_BUCKET_PREFIX", "my-prefix")
	defer unsetEnvVars()

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
	defer unsetEnvVars()

	flag.Set("source", "./watch")
	flag.Set("bucket", "new-s3-bucket")
	flag.Set("prefix", "new-prefix")
	flag.Set("region", "eu-central-1")
	defer restFlags()

	config := NewConfiguration()
	config.Load()

	if success := config.Validate(); success != true {
		t.Errorf("Expected Validate to return true for valid configuration, got %t", success)
	}
}

func TestValidateMissingEnvVarsAndFlags(t *testing.T) {
	unsetEnvVars()

	config := NewConfiguration()
	config.Load()

	if success := config.Validate(); success != false {
		t.Errorf("Expected Validate to return false due to missing environment variables, got %t", success)
	}
}
