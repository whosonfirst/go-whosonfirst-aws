package s3

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	_ "log"
	"os/user"
	"path/filepath"
	"strings"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func ReadCloserFromBytes(b []byte) (io.ReadCloser, error) {
	body := bytes.NewReader(b)
	return nopCloser{body}, nil
}

type S3Connection struct {
	session *session.Session
	service *s3.S3
	bucket  string
	prefix  string
}

type S3Config struct {
	Bucket      string
	Prefix      string
	Region      string
	Credentials string // see notes below
}

func ValidS3Credentials() []string {

	valid := []string{
		"env:",
		"iam:",
		"shared:{PATH}:{PROFILE}",
		"{PROFILE}",
	}

	return valid
}

func ValidS3CredentialsString() string {

	valid := ValidS3Credentials()
	return fmt.Sprintf("Valid credential flags are: %s", strings.Join(valid, ", "))
}

func NewS3Connection(s3cfg S3Config) (*S3Connection, error) {

	if s3cfg.Bucket == "" {
		return nil, errors.New("Invalid S3 bucket name")
	}

	// https://docs.aws.amazon.com/sdk-for-go/v1/developerguide/configuring-sdk.html
	// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/

	cfg := aws.NewConfig()
	cfg.WithRegion(s3cfg.Region)

	if strings.HasPrefix(s3cfg.Credentials, "env:") {

		creds := credentials.NewEnvCredentials()
		cfg.WithCredentials(creds)

	} else if strings.HasPrefix(s3cfg.Credentials, "shared:") {

		details := strings.Split(s3cfg.Credentials, ":")

		if len(details) != 3 {
			return nil, errors.New("Shared credentials need to be defined as 'shared:CREDENTIALS_FILE:PROFILE_NAME'")
		}

		creds := credentials.NewSharedCredentials(details[1], details[2])
		cfg.WithCredentials(creds)

	} else if strings.HasPrefix(s3cfg.Credentials, "iam:") {

		// assume an IAM role suffient for doing whatever

	} else if s3cfg.Credentials != "" {

		// for backwards compatibility as of 05a6042dc5956c13513bdc5ab4969877013f795c
		// (20161203/thisisaaronland)

		whoami, err := user.Current()

		if err != nil {
			return nil, err
		}

		dotaws := filepath.Join(whoami.HomeDir, ".aws")
		creds_file := filepath.Join(dotaws, "credentials")

		profile := s3cfg.Credentials

		creds := credentials.NewSharedCredentials(creds_file, profile)
		cfg.WithCredentials(creds)

	} else {

		// for backwards compatibility as of 05a6042dc5956c13513bdc5ab4969877013f795c
		// (20161203/thisisaaronland)

		creds := credentials.NewEnvCredentials()
		cfg.WithCredentials(creds)
	}

	sess := session.New(cfg)

	if s3cfg.Credentials != "" {

		_, err := sess.Config.Credentials.Get()

		if err != nil {
			return nil, err
		}
	}

	service := s3.New(sess)

	c := S3Connection{
		session: sess,
		service: service,
		bucket:  s3cfg.Bucket,
		prefix:  s3cfg.Prefix,
	}

	return &c, nil
}

// https://tools.ietf.org/html/rfc7231#section-4.3.2
// https://docs.aws.amazon.com/AmazonS3/latest/API/RESTObjectHEAD.html

func (conn *S3Connection) Head(key string) (*s3.HeadObjectOutput, error) {

	key = conn.prepareKey(key)

	params := &s3.HeadObjectInput{
		Bucket: aws.String(conn.bucket),
		Key:    aws.String(key),
	}

	rsp, err := conn.service.HeadObject(params)

	if err != nil {
		return nil, err
	}

	return rsp, nil
}

func (conn *S3Connection) Get(key string) (io.ReadCloser, error) {

	key = conn.prepareKey(key)

	params := &s3.GetObjectInput{
		Bucket: aws.String(conn.bucket),
		Key:    aws.String(key),
	}

	rsp, err := conn.service.GetObject(params)

	if err != nil {
		return nil, err
	}

	return rsp.Body, nil
}

func (conn *S3Connection) GetBytes(key string) ([]byte, error) {

	fh, err := conn.Get(key)

	if err != nil {
		return nil, err
	}

	defer fh.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(fh)

	return buf.Bytes(), nil
}

func (conn *S3Connection) Put(key string, fh io.ReadCloser) error {

	defer fh.Close()

	key = conn.prepareKey(key)

	uploader := s3manager.NewUploader(conn.session)

	params := &s3manager.UploadInput{
		Bucket: aws.String(conn.bucket),
		Key:    aws.String(key),
		Body:   fh,
	}

	_, err := uploader.Upload(params)
	return err

	/*
		params := &s3.PutObjectInput{
			Bucket: aws.String(conn.bucket),
			Key:    aws.String(key),
			Body:   fh,
			ACL:    aws.String("public-read"),
		}

		_, err := conn.service.PutObject(params)

		if err != nil {
			return err
		}

		return nil
	*/
}

func (conn *S3Connection) PutBytes(key string, body []byte) error {

	fh, err := ReadCloserFromBytes(body)

	if err != nil {
		return err
	}

	return conn.Put(key, fh)
}

func (conn *S3Connection) Delete(key string) error {

	key = conn.prepareKey(key)

	params := &s3.DeleteObjectInput{
		Bucket: aws.String(conn.bucket),
		Key:    aws.String(key),
	}

	_, err := conn.service.DeleteObject(params)

	if err != nil {
		return err
	}

	return nil
}

func (conn *S3Connection) HasChanged(key string, local []byte) (bool, error) {

	// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/#HeadObjectInput
	// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/#HeadObjectOutput

	head, err := conn.Head(key)

	if err != nil {

		aws_err := err.(awserr.Error)

		if aws_err.Code() == "NotFound" {
			return true, nil
		}

		if aws_err.Code() == "SlowDown" {

		}

		return false, err
	}

	enc := md5.Sum(local)
	local_hash := hex.EncodeToString(enc[:])

	etag := *head.ETag
	remote_hash := strings.Replace(etag, "\"", "", -1)

	if local_hash == remote_hash {
		return false, nil
	}

	return true, nil
}

func (conn *S3Connection) prepareKey(key string) string {

	if conn.prefix == "" {
		return key
	}

	return filepath.Join(conn.prefix, key)
}
