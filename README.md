# Overview

This is a small binary to upload a local directory to s3. It's mainly designed to simple upload files written to a certain directory to s3 without handling any lifecycle or versioning.

## Usage

The binary expects AWS default credentials to be available in the environment. This can ether be configured via [environment variables](https://docs.aws.amazon.com/cli/v1/userguide/cli-configure-envvars.html) or via the [shared credentials file](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html) at `~/.aws/credentials`.

The binary expects the following arguments:

| Environment Variable | Description |
|----------------------|-------------|
| `WATCH_DIR` | The directory to upload to s3 |
| `S3_BUCKET_NAME` | The name of the bucket to upload the files to |
| `S3_BUCKET_PREFIX` | The directory to upload to s3 |
