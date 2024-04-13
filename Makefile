GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")
LDFLAGS=-s -w

cli:
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/wof-launch-task cmd/wof-launch-task/main.go

lambda:
	@make lambda-launch-task

lambda-launch-task:
	if test -f bootstrap; then rm -f bootstrap; fi
	if test -f launch-task.zip; then rm -f launch-task.zip; fi
	GOARCH=arm64 GOOS=linux go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -tags lambda.norpc -o bootstrap cmd/wof-launch-task/main.go
	zip launch-task.zip bootstrap
	rm -f bootstrap
