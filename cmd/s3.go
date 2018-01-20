package main

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-aws/s3"
	_ "io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	valid_flags := s3.ValidS3CredentialsString()

	var bucket = flag.String("bucket", "", "A valid S3 bucket name")
	var prefix = flag.String("prefix", "", "An optional path/prefix inside the S3 bucket")
	var region = flag.String("region", "us-east-1", "A valid AWS S3 region")
	var credentials = flag.String("credentials", "env:", "A valid S3 credentials flag. "+valid_flags)

	flag.Parse()

	config := s3.S3Config{
		Bucket:      *bucket,
		Prefix:      *prefix,
		Region:      *region,
		Credentials: *credentials,
	}

	conn, err := s3.NewS3Connection(config)

	if err != nil {
		log.Fatal(err)
	}

	args := flag.Args()
	count := len(args)

	if count == 0 {
		log.Fatal("Missing S3 command")
	}

	if count == 1 {
		log.Fatal("Missing anything to do with your S3 command")
	}

	cmd := strings.ToUpper(args[0])

	for _, path := range args[1:] {

		var rsp interface{}
		var err error

		switch cmd {

		case "HEAD":
			rsp, err = conn.Head(path)

		case "GET":
			rsp, err = conn.Get(path)

		case "PUT":

			abs_path, err := filepath.Abs(path)

			if err != nil {
				log.Fatal(err)
			}

			key := filepath.Base(abs_path)

			fh, err := os.Open(abs_path)

			if err != nil {
				log.Fatal(err)
			}

			err = conn.Put(key, fh)

		case "DELETE":

			err = conn.Delete(path)
		default:
			log.Fatal("Invalid command")
		}

		if err != nil {
			log.Fatal(err)
		}

		log.Println(rsp)
	}

	os.Exit(0)
}
