package aws

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
)

func IsAWSError(err error) bool {
	_, is_aws := err.(awserr.Error)
	return is_aws
}

func IsAWSErrorWithCode(err error, code int) bool {
	aws_err, is_aws := err.(awserr.Error)
	return is_aws && aws_err.Status() == code
}
