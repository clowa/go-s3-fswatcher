package main

import (
	"flag"
	"log"
	"os"
)

type Configuration struct {
	watch_dir string
	// bucket_name   string
	// bucket_prefix string
	// aws_config    aws.Config
}

// NewConfiguration creates a new empty Configuration object.
func NewConfiguration() *Configuration {
	return &Configuration{}
}

// Load loads configuration values from environment variables or CLI flags.
func (c *Configuration) Load() {
	// var err error

	// Parse CLI flags
	flag.Parse()

	// Load AWS configuration
	// c.aws_config, err = config.LoadDefaultConfig(context.TODO())
	// if err != nil {
	// 	log.Fatalf("Unable to load SDK config, %v", err)
	// }

	// Load configuration from CLI flags or environment variables
	if *sourceFlag == "" {
		log.Fatal("Please provide a source directory using the -source flag.")
	}
	c.watch_dir = *sourceFlag

	// if *bucketFlag == "" {
	// 	log.Fatal("Please provide a bucket name using the -bucket flag.")
	// }
	// c.bucket_name = *bucketFlag

	// if *prefixFlag == "" {
	// 	log.Fatal("Please provide a prefix using the -prefix flag.")
	// }
	// c.bucket_prefix = *prefixFlag
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

	return valid
}
