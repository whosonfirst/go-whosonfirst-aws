package ecs

import (
	"github.com/aws/aws-sdk-go/aws"
	aws_cloudwatch "github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	aws_ecs "github.com/aws/aws-sdk-go/service/ecs"
	"github.com/whosonfirst/go-whosonfirst-aws/session"
)

type TaskOptions struct {
	DSN            string
	Task           string
	Container      string
	Cluster        string
	LaunchType     string
	PublicIP       string
	Subnets        []string
	SecurityGroups []string
}

type MonitorOptions struct {
	DSN string
}

func LaunchTaskAndMonitor(ecs_opts *TaskOptions, cw_opts *MonitorOptions, cmd ...string) error {

	cw_sess, err := session.NewSessionWithDSN(cw_opts.DSN)

	if err != nil {
		return err
	}

	cw_svc := aws_cloudwatchlogs.New(cw_sess)

	rsp, err := LaunchTask(ecs_opts*TaskOptions, cmd...)

	if err != nil {
		return err
	}

	count_tasks := len(rsp.Tasks)
	remaining := count_tasks

	ecs_tasks := make([]*string, count_tasks)

	for i, t := range rsp.Tasks {
		ecs_tasks[i] = t.TaskArn
	}

	task_errors := make([]error, 0)

	for remaining > 0 {

		monitor_req := &ecs.DescribeTasksInput{
			Cluster: aws.String(ecs_opts.Cluster),
			Tasks:   ecs_tasks,
		}

		monitor_rsp, err := ecs_svc.DescribeTasks(monitor_req)

		if err != nil {
			return err
		}

		for _, t := range monitor_rsp.Tasks {

			for _, c := range t.Containers {

				if *c.Name != ecs_opts.Container {
					continue
				}

				if *c.LastStatus != "STOPPED" {
					continue
				}

				// start of generic code to put in a function
				// TO DO: what if the logs haven't reached CW yet... ?

				arn := strings.Split(*t.TaskArn, "/")

				cw_group := fmt.Sprintf("/ecs/%s", opts.Container)
				cw_stream := fmt.Sprintf("ecs/%s/%s", opts.Container, arn[1])

				cw_req := &cloudwatchlogs.GetLogEventsInput{
					LogGroupName:  aws.String(cw_group),
					LogStreamName: aws.String(cw_stream),
					StartFromHead: aws.Bool(true),
				}

				cw_rsp, err := cw_svc.GetLogEvents(cw_req)

				if err == nil {

					for _, e := range cw_rsp.Events {
						log.Printf("[%s][%d] %s\n", *t.TaskArn, *e.Timestamp, *e.Message)
					}
				}

				// TODO: paginated logs...
				// end of generic code to put in a function

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

func LaunchTask(opts *TaskOptions, cmd ...string) (*aws_ecs.RunTaskOutput, error) {

	ecs_sess, err := session.NewSessionWithDSN(opts.DSN)

	if err != nil {
		return nil, err
	}

	ecs_svc := aws_ecs.New(ecs_sess)

	cluster := aws.String(opts.Cluster)
	task := aws.String(opts.Task)

	launch_type := aws.String(opts.LaunchType)
	public_ip := aws.String(opts.PublicIP)

	subnets := make([]*string, len(opts.Subnets))
	security_groups := make([]*string, len(opts.SecurityGroups))

	for i, sn := range opts.Subnets {
		subnets[i] = aws.String(sn)
	}

	for i, sg := range opts.SecurityGroups {
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
		Name:    aws.String(opts.Container),
		Command: aws_cmd,
	}

	overrides := &aws_ecs.TaskOverride{
		ContainerOverrides: []*ecs.ContainerOverride{
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

	rsp, err := ecs_svc.RunTask(input)

	if err != nil {
		return nil, err
	}

	if len(rsp.Tasks) == 0 {
		return nil, errors.New("run task returned no errors... but no tasks")
	}

	return rsp, nil
}
