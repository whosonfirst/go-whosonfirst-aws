# go-aws-ecs

Go package for basic AWS ECS related operations.

## Documentation

Documentation is incomplete.

## Tools

To build binary versions of these tools run the `cli` Makefile target. For example:

```
$> make cli
go build -mod vendor -ldflags="-s -w" -o bin/ecs-launch-task cmd/ecs-launch-task/main.go
```

### ecs-launch-task

Launch an ECS task from the command line.

```
$> ./bin/ecs-launch-task -h
Usage of ./bin/ecs-launch-task:
  -cluster string
    	The name of your ECS cluster.
  -container string
    	The name of your ECS container.
  -session-uri string
    	A valid aaronland/go-aws-session URI.
  -launch-type string
    	A valid ECS launch type.
  -platform-version string
    	A valid ECS platform version.
  -public-ip string
    	A valid ECS public IP string.
  -security-group value
    	A valid AWS security group to run your task under.
  -subnet value
    	One or more subnets to run your ECS task in.
  -task string
    	The name (and version) of your ECS task.
```

#### Session URI strings

The following parameters are required in session URI string:

##### Credentials

Credentials for AWS sessions are defined as string labels. They are:

| Label | Description |
| --- | --- |
| `env:` | Read credentials from AWS defined environment variables. |
| `iam:` | Assume AWS IAM credentials are in effect. |
| `{AWS_PROFILE_NAME}` | This this profile from the default AWS credentials location. |
| `{AWS_CREDENTIALS_PATH}:{AWS_PROFILE_NAME}` | This this profile from a user-defined AWS credentials location. |

##### Region

Any valid AWS region.

##### For example:

```
aws://?region=us-east-1&credentials=session
```

## See also

* https://github.com/aws/aws-sdk-go
* https://github.com/aaronland/go-aws-session