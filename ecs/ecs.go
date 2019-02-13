package ecs

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	aws_ecs "github.com/aws/aws-sdk-go/service/ecs"
	"github.com/whosonfirst/go-whosonfirst-aws/cloudwatch"
	"github.com/whosonfirst/go-whosonfirst-aws/session"
	"log"
	"strings"
)

type TaskResponse struct {
	Tasks []string
}

type TaskOptions struct {
	DSN            string
	Task           string
	Container      string
	Cluster        string
	LaunchType     string
	PublicIP       string
	Subnets        []string
	SecurityGroups []string
	Monitor        bool
	Logs           bool
	LogsDSN        string
}

func LaunchTask(task_opts *TaskOptions, cmd ...string) (*aws_ecs.RunTaskOutput, error) {

	ecs_sess, err := session.NewSessionWithDSN(task_opts.DSN)

	if err != nil {
		return nil, err
	}

	ecs_svc := aws_ecs.New(ecs_sess)

	cluster := aws.String(task_opts.Cluster)
	task := aws.String(task_opts.Task)

	launch_type := aws.String(task_opts.LaunchType)
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
		NetworkConfiguration: network,
		Overrides:            overrides,
	}

	task_rsp, err := ecs_svc.RunTask(input)

	if err != nil {
		return nil, err
	}

	if len(task_rsp.Tasks) == 0 {
		return nil, errors.New("run task returned no errors... but no tasks")
	}

	if task_opts.Monitor {

		task_arns := make([]string, len(task_rsp.Tasks))

		for i, t := range task_rsp.Tasks {
			task_arns[i] = *t.TaskArn
		}

		err := MonitorTasksWithECSService(ecs_svc, task_opts, task_arns...)

		if err != nil {
			return nil, err
		}
	}

	return task_rsp, nil
}

func MonitorTasks(task_opts *TaskOptions, task_arns ...string) error {

	ecs_sess, err := session.NewSessionWithDSN(task_opts.DSN)

	if err != nil {
		return err
	}

	ecs_svc := aws_ecs.New(ecs_sess)

	return MonitorTasksWithECSService(ecs_svc, task_opts, task_arns...)
}

func MonitorTasksWithECSService(ecs_svc *aws_ecs.ECS, task_opts *TaskOptions, task_arns ...string) error {

	count_tasks := len(task_arns)
	remaining := count_tasks

	ecs_tasks := make([]*string, count_tasks)

	for i, t := range task_arns {
		ecs_tasks[i] = aws.String(t)
	}

	task_errors := make([]error, 0)

	for remaining > 0 {

		monitor_req := &aws_ecs.DescribeTasksInput{
			Cluster: aws.String(task_opts.Cluster),
			Tasks:   ecs_tasks,
		}

		monitor_rsp, err := ecs_svc.DescribeTasks(monitor_req)

		if err != nil {
			return err
		}

		for _, t := range monitor_rsp.Tasks {

			for _, c := range t.Containers {

				if *c.Name != task_opts.Container {
					continue
				}

				if *c.LastStatus != "STOPPED" {
					continue
				}

				if task_opts.Logs {

					arn := strings.Split(*t.TaskArn, "/")

					cw_group := fmt.Sprintf("/ecs/%s", task_opts.Container)
					cw_stream := fmt.Sprintf("ecs/%s/%s", task_opts.Container, arn[1])

					events, err := cloudwatch.GetLogEvents(task_opts.LogsDSN, cw_group, cw_stream)

					if err == nil {

						for _, e := range events {
							log.Println(*e.Message)
						}
					}
				}

				if *c.ExitCode != 0 {
					msg := fmt.Sprintf("Task %s failed with exit code %d\n", *t.TaskArn, *c.ExitCode)
					err := errors.New(msg)
					task_errors = append(task_errors, err)
				}

				remaining -= 1
			}
		}
	}

	if len(task_errors) > 0 {

		for _, e := range task_errors {
			log.Println(e)
		}

	}

	return nil
}
