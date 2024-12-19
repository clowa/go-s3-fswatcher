package main

import (
	"flag"
	"os"
	"testing"
)

func TestLoadConfigFromEnvVars(t *testing.T) {
	os.Setenv("AWS_DEFAULT_REGION", "us-west-2")
	os.Setenv("WATCH_DIR", "/path/to/watch")
	os.Setenv("S3_BUCKET_NAME", "my-s3-bucket")
	os.Setenv("S3_BUCKET_PREFIX", "my-prefix")

	loadConfig()

	if config.watch_dir != "/path/to/watch" {
		t.Errorf("Expected watch_dir to be /path/to/watch, got %s", config.watch_dir)
	}
	if config.bucket_name != "my-s3-bucket" {
		t.Errorf("Expected bucket_name to be my-s3-bucket, got %s", config.bucket_name)
	}
	if config.bucket_prefix != "my-prefix" {
		t.Errorf("Expected bucket_prefix to be my-prefix, got %s", config.bucket_prefix)
	}
}

func TestLoadConfigFromFlags(t *testing.T) {
	os.Setenv("AWS_DEFAULT_REGION", "us-west-2")
	os.Setenv("WATCH_DIR", "./.vscode")
	os.Setenv("S3_BUCKET_NAME", "my-s3-bucket")
	os.Setenv("S3_BUCKET_PREFIX", "my-prefix")

	flag.Set("source", "./watch")
	flag.Set("bucket", "new-s3-bucket")
	flag.Set("prefix", "new-prefix")

	loadConfig()

	if config.watch_dir != "./watch" {
		t.Errorf("Expected watch_dir to be ./watch, got %s", config.watch_dir)
	}
	if config.bucket_name != "new-s3-bucket" {
		t.Errorf("Expected bucket_name to be new-s3-bucket, got %s", config.bucket_name)
	}
	if config.bucket_prefix != "new-prefix" {
		t.Errorf("Expected bucket_prefix to be new-prefix, got %s", config.bucket_prefix)
	}
}

func TestLoadConfigMissingEnvVarsAndFlags(t *testing.T) {
	os.Unsetenv("AWS_DEFAULT_REGION")
	os.Unsetenv("WATCH_DIR")
	os.Unsetenv("S3_BUCKET_NAME")
	os.Unsetenv("S3_BUCKET_PREFIX")

	valid := loadConfig()

	if valid {
		t.Errorf("Expected loadConfig to return false due to missing environment variables and flags")
	}
}

func TestValidateConfig(t *testing.T) {
	os.Setenv("AWS_DEFAULT_REGION", "us-west-2")
	os.Setenv("WATCH_DIR", "/path/to/watch")
	os.Setenv("S3_BUCKET_NAME", "my-s3-bucket")
	os.Setenv("S3_BUCKET_PREFIX", "my-prefix")

	valid := validateConfig()

	if !valid {
		t.Errorf("Expected validateConfig to return true for valid configuration")
	}
}

func TestValidateConfigMissingEnvVars(t *testing.T) {
	os.Unsetenv("AWS_DEFAULT_REGION")
	os.Unsetenv("WATCH_DIR")
	os.Unsetenv("S3_BUCKET_NAME")
	os.Unsetenv("S3_BUCKET_PREFIX")

	valid := validateConfig()

	if valid {
		t.Errorf("Expected validateConfig to return false due to missing environment variables")
	}
}
