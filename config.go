package main

import (
	"flag"
	"log"
	"os"
)

type Configuration struct {
	watch_dir     string
	bucket_name   string
	bucket_prefix string
	aws_region    string
}

// NewConfiguration creates a new empty Configuration object.
func NewConfiguration() *Configuration {
	return &Configuration{}
}

// Load loads configuration values from environment variables or CLI flags.
func (c *Configuration) Load() {
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

	if *regionFlag != "" {
		c.aws_region = *regionFlag
	} else {
		if os.Getenv("AWS_REGION") != "" {
			c.aws_region = os.Getenv("AWS_REGION")
		} else {
			// When no region flag has been set try using the AWS_DEFAULT_REGION environment variable.
			// This will be an emptry string if the environment variable is not set.
			c.aws_region = os.Getenv("AWS_DEFAULT_REGION")
		}
	}
}

// Validate encapsulates the validation logic for the configuration values.
// It returns true if the configuration values are valid, false otherwise.
func (c *Configuration) Validate() bool {
	// Assume configuration is valid and check for invalid values
	valid := true

	// Validate source directory
	if _, err := os.Stat(c.watch_dir); os.IsNotExist(err) {
		log.Printf("Invalid source directory. Please provide a valid directory path. Example: /path/to/source")
		valid = false
	}

	// Validate bucket name
	if c.bucket_name == "" {
		log.Printf("Invalid S3 bucket name. Please provide a valid bucket name. Example: my-s3-bucket")
		valid = false
	}

	// Validate prefix
	// An empty prefix is valid, so no need to validate it

	// Validate AWS region
	if c.aws_region == "" {
		log.Printf("Invalid AWS region. Please provide a valid AWS region. Example: us-west-2")
		valid = false
	}

	return valid
}
