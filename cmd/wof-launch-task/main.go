// Launch an ECS task for one of more Who's On First (style) data GitHub repositories than
// have been updated with an ISO-8601-encoded period of time.
package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aaronland/go-aws-ecs"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/multi"
	"github.com/sfomuseum/iso8601duration"
	"github.com/sfomuseum/runtimevar"
	"github.com/whosonfirst/go-whosonfirst-github/organizations"
)

func main() {

	var mode string

	var github_org string
	var github_prefix multi.MultiCSVString
	var github_exclude multi.MultiCSVString
	var github_access_token_uri string
	var github_updated_since string

	var aws_session_uri string

	var ecs_task string
	var ecs_container string
	var ecs_cluster string
	var ecs_launch_type string
	var ecs_platform string
	var ecs_public_ip string
	var ecs_task_command string
	var ecs_subnets multi.MultiCSVString
	var ecs_security_groups multi.MultiCSVString

	var task_per_repo bool
	var dryrun bool

	fs := flagset.NewFlagSet("update")

	fs.StringVar(&mode, "mode", "cli", "Valid options are: cli, lambda")

	fs.StringVar(&github_org, "github-organization", "whosonfirst-data", "The GitHub organization to poll for recently updated repositories.")
	fs.Var(&github_prefix, "github-prefix", "Zero or more prefixes to filter repositories by (must match). Prefixes may also passed in a single comma-separated string.")
	fs.Var(&github_prefix, "github-exclude", "Zero or more prefixes to exclude repositories by (must NOT match). Prefixes may also passed in a single comma-separated string.")
	fs.StringVar(&github_access_token_uri, "github-access-token-uri", "", "A valid gocloud.dev/runtimevar URI that dereferences to a GitHub API access token.")
	fs.StringVar(&github_updated_since, "github-updated-since", "PT24H", "A valid ISO-8601 duration string.")

	fs.StringVar(&aws_session_uri, "aws-session-uri", "", "A valid aaronland/go-aws-session URI string.")

	fs.StringVar(&ecs_task, "ecs-task", "", "The name (and version) of your ECS task.")
	fs.StringVar(&ecs_container, "ecs-container", "", "The name of your ECS container.")
	fs.StringVar(&ecs_cluster, "ecs-cluster", "", "The name of your ECS cluster.")
	fs.StringVar(&ecs_launch_type, "ecs-launch-type", "FARGATE", "A valid ECS launch type.")
	fs.StringVar(&ecs_platform, "ecs-platform-version", "1.4.0", "A valid ECS platform version.")
	fs.StringVar(&ecs_public_ip, "ecs-public-ip", "ENABLED", "A valid ECS public IP string.")
	fs.Var(&ecs_subnets, "ecs-subnet", "One or more subnets to run your ECS task in. Subnets may also passed in a single comma-separated string.")
	fs.Var(&ecs_security_groups, "ecs-security-group", "A valid AWS security group to run your task under.")
	fs.StringVar(&ecs_task_command, "ecs-task-command", "", "A option command string to pass to the ECS task")
	fs.BoolVar(&task_per_repo, "task-per-repo", false, "A boolean flag indicating whether individual tasks should be launched for each repo updated.")

	fs.BoolVar(&dryrun, "dryrun", false, "Go through the motions but do not launch any indexing tasks.")
	flagset.Parse(fs)

	err := flagset.SetFlagsFromEnvVars(fs, "WHOSONFIRST")

	if err != nil {
		log.Fatalf("Failed to set flags from environment variables, %w", err)
	}

	ctx := context.Background()

	svc, err := ecs.NewService(aws_session_uri)

	if err != nil {
		log.Fatalf("Failed to create new service, %v", err)
	}

	task_opts := &ecs.TaskOptions{
		Task:            ecs_task,
		Container:       ecs_container,
		Cluster:         ecs_cluster,
		LaunchType:      ecs_launch_type,
		PlatformVersion: ecs_platform,
		PublicIP:        ecs_public_ip,
		Subnets:         ecs_subnets,
		SecurityGroups:  ecs_security_groups,
	}

	list_opts := organizations.NewDefaultListOptions()

	d, err := duration.FromString(github_updated_since)

	if err != nil {
		log.Fatalf("Failed to parse '%s', %w", github_updated_since, err)
	}

	now := time.Now()
	since := now.Add(-d.ToDuration())

	list_opts.PushedSince = &since

	if len(github_prefix) > 0 {
		list_opts.Prefix = github_prefix
	}

	if len(github_exclude) > 0 {
		list_opts.Exclude = github_exclude
	}

	if github_access_token_uri != "" {

		access_token, err := runtimevar.StringVar(ctx, github_access_token_uri)

		if err != nil {
			log.Fatalf("Failed to deference github access token URI, %w", err)
		}

		list_opts.AccessToken = access_token
	}

	updateFunc := func(ctx context.Context) error {

		repos, err := organizations.ListRepos(github_org, list_opts)

		if err != nil {
			return fmt.Errorf("Failed to list repos for %s, %w", github_org, err)
		}

		if len(repos) == 0 {
			return nil
		}

		log.Printf("One or more (%d) repos has been updated, %s\n", len(repos), strings.Join(repos, ","))

		cmd := make([]string, 0)

		if !task_per_repo {

			if ecs_task_command != "" {
				cmd = strings.Split(ecs_task_command, " ")
			}

			if dryrun {
				log.Printf("[dryrun] Launch task '%s' (%s) with command '%s' (%v)\n", task_opts.Task, task_opts.Container, ecs_task_command, cmd)
			} else {

				task_rsp, err := ecs.LaunchTask(ctx, svc, task_opts, cmd...)

				if err != nil {
					return fmt.Errorf("Failed to launch ECS task for %s, %w", task_opts.Task, err)
				}

				log.Printf("Launched task %s with command '%s' with ARNs %s\n", task_opts.Task, ecs_task_command, strings.Join(task_rsp.Tasks, ","))
			}

		} else {

			for _, name := range repos {

				var task_cmd string

				if ecs_task_command != "" {
					task_cmd = strings.Replace(ecs_task_command, "{repo}", name, -1)
					cmd = strings.Split(task_cmd, " ")
				}

				if dryrun {
					log.Printf("[dryrun] %s (%s) %s\n", task_opts.Task, task_opts.Container, task_cmd)
					continue
				}

				task_rsp, err := ecs.LaunchTask(ctx, svc, task_opts, cmd...)

				if err != nil {
					return fmt.Errorf("Failed to launch ECS task for %s, %w", name, err)
				}

				log.Printf("Launched task %s for %s with ARNs %s\n", task_opts.Task, name, strings.Join(task_rsp.Tasks, ","))
			}
		}

		return nil
	}

	switch mode {
	case "cli":

		err := updateFunc(ctx)

		if err != nil {
			log.Fatalf("Failed to perform updates, %w", err)
		}

	case "lambda":

		lambda.Start(updateFunc)

	default:
		log.Fatalf("Invalid mode '%s'", mode)
	}
}
