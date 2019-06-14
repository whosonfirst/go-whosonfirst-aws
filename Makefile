fmt:
	go fmt cloudwatch/*.go
	go fmt cmd/*.go
	go fmt config/*.go
	go fmt ecs/*.go
	go fmt lambda/*.go
	go fmt s3/*.go
	go fmt sqs/*.go
	go fmt session/*.go
	go fmt util/*.go

tools:
	go build -mod vendor -o bin/s3 cmd/s3/main.go
	go build -mod vendor -o bin/secret cmd/secret/main.go
	go build -mod vendor -o bin/ecs-run-task cmd/ecs-run-task/main.go
	go build -mod vendor -o bin/lambda-run-task cmd/lambda-run-task/main.go
