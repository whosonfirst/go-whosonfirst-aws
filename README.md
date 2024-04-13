# go-whosonfirst-aws

Go package for working with Who's On First records in an AWS setting.

## Tools

```
$> make cli
go build -mod vendor -ldflags="-s -w" -o bin/wof-launch-task cmd/wof-launch-task/main.go
```

### wof-launch-task

Launch an ECS task for one of more Who's On First (style) data GitHub repositories than have been updated with an ISO-8601-encoded period of time.

```
$> ./bin/wof-launch-task -h
  -aws-session-uri string
    	A valid aaronland/go-aws-session URI string.
  -dryrun
    	Go through the motions but do not launch any indexing tasks.
  -ecs-cluster string
    	The name of your ECS cluster.
  -ecs-container string
    	The name of your ECS container.
  -ecs-launch-type string
    	A valid ECS launch type. (default "FARGATE")
  -ecs-platform-version string
    	A valid ECS platform version. (default "1.4.0")
  -ecs-public-ip string
    	A valid ECS public IP string. (default "ENABLED")
  -ecs-security-group value
    	A valid AWS security group to run your task under.
  -ecs-subnet value
    	One or more subnets to run your ECS task in. Subnets may also passed in a single comma-separated string.
  -ecs-task string
    	The name (and version) of your ECS task.
  -ecs-task-command string
    	A option command string to pass to the ECS task
  -github-access-token-uri string
    	A valid gocloud.dev/runtimevar URI that dereferences to a GitHub API access token.
  -github-exclude value
    	Zero or more prefixes to exclude repositories by (must NOT match). Prefixes may also passed in a single comma-separated string.
  -github-organization string
    	The GitHub organization to poll for recently updated repositories. (default "whosonfirst-data")
  -github-prefix value
    	Zero or more prefixes to filter repositories by (must match). Prefixes may also passed in a single comma-separated string.
  -github-updated-since string
    	A valid ISO-8601 duration string. (default "PT24H")
  -mode string
    	Valid options are: cli, lambda (default "cli")
  -task-per-repo
    	A boolean flag indicating whether individual tasks should be launched for each repo updated.
```

#### Lambda

```
$> make lambda-launch-task
if test -f bootstrap; then rm -f bootstrap; fi
if test -f launch-task.zip; then rm -f launch-task.zip; fi
GOARCH=arm64 GOOS=linux go build -mod vendor -ldflags="-s -w" -tags lambda.norpc -o bootstrap cmd/wof-launch-task/main.go
zip launch-task.zip bootstrap
  adding: bootstrap (deflated 75%)
rm -f bootstrap
```

## See also

* https://github.com/whosonfirst/go-whosonfirst-github
* https://github.com/aaronland/go-aws-ecs