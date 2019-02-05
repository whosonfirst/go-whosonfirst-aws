package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"	
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/whosonfirst/go-whosonfirst-aws/session"
	"github.com/whosonfirst/go-whosonfirst-cli/flags"
	"log"
)

func main() {

	var ecs_dsn = flag.String("ecs-dsn", "", "A valid (go-whosonfirst-aws) ECS DSN.")

	var container = flag.String("container", "", "The name of your AWS ECS container.")
	var cluster = flag.String("cluster", "", "The name of your AWS ECS cluster.")
	var task = flag.String("task", "", "The name of your AWS ECS task (inclusive of its version number),")

	var launch_type = flag.String("launch-type", "FARGATE", "...")
	var public_ip = flag.String("public-ip", "ENABLED", "...")
	
	var subnets flags.MultiString
	flag.Var(&subnets, "subnet", "One or more AWS subnets in which your task will run.")

	var security_groups flags.MultiString
	flag.Var(&security_groups, "security-group", "One of more AWS security groups your task will assume.")

	flag.Parse()
	
	sess, err := session.NewSessionWithDSN(*ecs_dsn)

	if err != nil {
		log.Fatal(err)
	}
	
	svc := ecs.New(sess)

	ecs_cluster := aws.String(*cluster)
	ecs_task := aws.String(*task)

	ecs_launch_type := aws.String(*launch_type)
	ecs_public_ip := aws.String(*public_ip)

	ecs_cmd := make([]*string, len(flag.Args()))

	for i, fl := range flag.Args() {
		ecs_cmd[i] = aws.String(fl)
	}
	
	ecs_subnets := make([]*string, len(subnets))
	ecs_security_groups := make([]*string, len(security_groups))

	for i, sn := range subnets {
		ecs_subnets[i] = aws.String(sn)
	}

	for i, sg := range security_groups {
		ecs_security_groups[i] = aws.String(sg)
	}

	ecs_network := &ecs.NetworkConfiguration{
		AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
			AssignPublicIp: ecs_public_ip,
			SecurityGroups: ecs_security_groups,
			Subnets:        ecs_subnets,
		},
	}

	ecs_process_override := &ecs.ContainerOverride{
		Name:    aws.String(*container),
		Command: ecs_cmd,
	}

	ecs_overrides := &ecs.TaskOverride{
		ContainerOverrides: []*ecs.ContainerOverride{
			ecs_process_override,
		},
	}

	req := &ecs.RunTaskInput{
		Cluster:              ecs_cluster,
		TaskDefinition:       ecs_task,
		LaunchType:           ecs_launch_type,
		NetworkConfiguration: ecs_network,
		Overrides:            ecs_overrides,
	}

	rsp, err := svc.RunTask(req)

	if err != nil {
		log.Fatal(err)
	}

	task_id := rsp.Tasks[0].TaskArn
	fmt.Println(task_id)
}
