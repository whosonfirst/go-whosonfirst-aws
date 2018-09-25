package main

// https://docs.aws.amazon.com/sdk-for-go/api/service/secretsmanager/#example_SecretsManager_GetSecretValue_shared00

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"log"
	"os/user"
	"path/filepath"
	"strings"
)

type SecretsConfig struct {
	Region      string
	Credentials string // see notes below
}

func main() {

	secrets_cfg := SecretsConfig{
		Credentials: "session",
		Region:      "us-west-2",
	}

	cfg := aws.NewConfig()
	cfg.WithRegion(secrets_cfg.Region)

	if strings.HasPrefix(secrets_cfg.Credentials, "env:") {

		creds := credentials.NewEnvCredentials()
		cfg.WithCredentials(creds)

	} else if strings.HasPrefix(secrets_cfg.Credentials, "iam:") {

		// assume an IAM role suffient for doing whatever

	} else if secrets_cfg.Credentials != "" {

		details := strings.Split(secrets_cfg.Credentials, ":")

		var creds_file string
		var profile string

		if len(details) == 1 {

			whoami, err := user.Current()

			if err != nil {
				log.Fatal(err)
			}

			dotaws := filepath.Join(whoami.HomeDir, ".aws")
			creds_file = filepath.Join(dotaws, "credentials")

			profile = details[0]

		} else {

			path, err := filepath.Abs(details[0])

			if err != nil {
				log.Fatal(err)
			}

			creds_file = path
			profile = details[1]
		}

		creds := credentials.NewSharedCredentials(creds_file, profile)
		cfg.WithCredentials(creds)

	} else {

		// for backwards compatibility as of 05a6042dc5956c13513bdc5ab4969877013f795c
		// (20161203/thisisaaronland)

		creds := credentials.NewEnvCredentials()
		cfg.WithCredentials(creds)
	}

	sess := session.New(cfg)

	if secrets_cfg.Credentials != "" {

		_, err := sess.Config.Credentials.Get()

		if err != nil {
			log.Fatal(err)
		}
	}

	secret_name := "sfomuseum_collection_ro"

	svc := secretsmanager.New(sess)

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secret_name),
		// VersionStage: aws.String("AWSPREVIOUS"),
	}

	result, err := svc.GetSecretValue(input)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)
}
