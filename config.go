package main

import (
	"flag"
	"log"
	"os"
)

type configuration struct {
	watch_dir     string
	bucket_name   string
	bucket_prefix string
	aws_region    string
}

// Load loads configuration values from environment variables and run validation checks.
// It returns true if the configuration values are valid, false otherwise.
func (c *configuration) Load() bool {
	// Parse CLI flags
	flag.Parse()

	c.aws_region = *regionFlag

	// Load configuration from CLI flags or environment variables
	if *sourceFlag != "" {
		c.watch_dir = *sourceFlag
	} else {
		c.watch_dir = os.Getenv("WATCH_DIR")
	}

	if *bucketFlag != "" {
		c.bucket_name = *bucketFlag
	} else {
		c.bucket_name = os.Getenv("S3_BUCKET_NAME")
	}

	if *prefixFlag != "" {
		c.bucket_prefix = *prefixFlag
	} else {
		c.bucket_prefix = os.Getenv("S3_BUCKET_PREFIX")
	}

	// Validate configuration values
	return c.Validate()
}

// Validate encapsulates the validation logic for the configuration values.
// It returns true if the configuration values are valid, false otherwise.
func (c *configuration) Validate() bool {
	// Validate source directory
	if _, err := os.Stat(config.watch_dir); os.IsNotExist(err) {
		log.Printf("Invalid source directory. Please provide a valid directory path. Example: /path/to/source")
		return false
	}

	// Validate bucket name
	if config.bucket_name == "" {
		log.Printf("Invalid S3 bucket name. Please provide a valid bucket name. Example: my-s3-bucket")
		return false
	}

	// Validate prefix
	if config.bucket_prefix == "" {
		log.Printf("Invalid S3 prefix. Please provide a valid prefix. Example: my-prefix/")
		return false
	}

	return true
}
