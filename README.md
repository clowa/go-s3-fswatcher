# Overview

This is a small binary to upload a local directory to s3. It's mainly designed to simple upload files written to a certain directory to s3 without handling any lifecycle or versioning.

## Usage

The binary expects AWS default credentials to be available in the environment. This can ether be configured via [environment variables](https://docs.aws.amazon.com/cli/v1/userguide/cli-configure-envvars.html) or via the [shared credentials file](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html) at `~/.aws/credentials`.

The binary expects the following configuration:

| CLI Flag   | Description | Example |
|------------|-------------|---------|
| `--source` | The directory to upload to s3 | `/path/to/source` |
| `--bucket` | The name of the bucket to upload the files to | `my-s3-bucket` |
| `--prefix` | The directory to upload to s3 | `my-prefix/` |

## Run Tests

### Run Unit Tests

To run the unit tests of the whole project, use the following command:

```sh
go test ./...
```

### Running Integration Tests

Make sure to build the binary before running the tests:

```sh
go build -o go-s3-fswatcher
```

To run the integration tests, use the following command:

```sh
go test ./... -tags=integration
```
