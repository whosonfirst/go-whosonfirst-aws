package ecs

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aaronland/go-aws-session"
	"github.com/aws/aws-sdk-go/aws"
	aws_ecs "github.com/aws/aws-sdk-go/service/ecs"
)

type TaskResponse struct {
	Tasks      []string
	TaskOutput *aws_ecs.RunTaskOutput
}

type TaskOptions struct {
	Task            string
	Container       string
	Cluster         string
	LaunchType      string
	PlatformVersion string
	PublicIP        string
	Subnets         []string
	SecurityGroups  []string
}

type WaitTasksOptions struct {
	Cluster  string
	TaskArns []string
	Timeout  time.Duration
	Interval time.Duration
	Logger   *log.Logger
}

func NewService(session_uri string) (*aws_ecs.ECS, error) {

	sess, err := session.NewSession(session_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create session, %v", err)
	}

	return aws_ecs.New(sess), nil
}

func LaunchTask(ctx context.Context, ecs_svc *aws_ecs.ECS, task_opts *TaskOptions, cmd ...string) (*TaskResponse, error) {

	cluster := aws.String(task_opts.Cluster)
	task := aws.String(task_opts.Task)

	launch_type := aws.String(task_opts.LaunchType)
	platform_version := aws.String(task_opts.PlatformVersion)
	public_ip := aws.String(task_opts.PublicIP)

	subnets := make([]*string, len(task_opts.Subnets))
	security_groups := make([]*string, len(task_opts.SecurityGroups))

	for i, sn := range task_opts.Subnets {
		subnets[i] = aws.String(sn)
	}

	for i, sg := range task_opts.SecurityGroups {
		security_groups[i] = aws.String(sg)
	}

	aws_cmd := make([]*string, len(cmd))

	for i, str := range cmd {
		aws_cmd[i] = aws.String(str)
	}

	network := &aws_ecs.NetworkConfiguration{
		AwsvpcConfiguration: &aws_ecs.AwsVpcConfiguration{
			AssignPublicIp: public_ip,
			SecurityGroups: security_groups,
			Subnets:        subnets,
		},
	}

	process_override := &aws_ecs.ContainerOverride{
		Name:    aws.String(task_opts.Container),
		Command: aws_cmd,
	}

	overrides := &aws_ecs.TaskOverride{
		ContainerOverrides: []*aws_ecs.ContainerOverride{
			process_override,
		},
	}

	input := &aws_ecs.RunTaskInput{
		Cluster:              cluster,
		TaskDefinition:       task,
		LaunchType:           launch_type,
		PlatformVersion:      platform_version,
		NetworkConfiguration: network,
		Overrides:            overrides,
	}

	task_output, err := ecs_svc.RunTask(input)

	if err != nil {
		return nil, err
	}

	if len(task_output.Tasks) == 0 {
		return nil, fmt.Errorf("run task returned no errors... but no tasks")
	}

	task_arns := make([]string, len(task_output.Tasks))

	for i, t := range task_output.Tasks {
		task_arns[i] = *t.TaskArn
	}

	task_rsp := &TaskResponse{
		Tasks:      task_arns,
		TaskOutput: task_output,
	}

	return task_rsp, nil
}

func WaitForTasksToComplete(ctx context.Context, ecs_svc *aws_ecs.ECS, opts *WaitTasksOptions) error {

	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	remaining := len(opts.TaskArns)

	for remaining > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case now := <-ticker.C:

			list_input := &aws_ecs.ListTasksInput{
				Cluster:       aws.String(opts.Cluster),
				DesiredStatus: aws.String("STOPPED"),
			}

			list_rsp, err := ecs_svc.ListTasks(list_input)

			if err != nil {
				return fmt.Errorf("Failed to list tasks, %w", err)
			}

			for _, stopped_t := range list_rsp.TaskArns {

				for _, t := range opts.TaskArns {

					if *stopped_t == t {
						remaining -= 1
						break
					}
				}
			}

			if opts.Logger != nil {
				opts.Logger.Printf("%v %d tasks remaining", now, remaining)
			}
		}
	}

	return nil
}
