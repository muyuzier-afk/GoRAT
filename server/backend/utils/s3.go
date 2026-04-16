package utils

import (
	"bytes"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var S3Client *s3.S3

func InitS3Client() error {
	s3Endpoint := os.Getenv("S3_ENDPOINT")
	if s3Endpoint == "" {
		s3Endpoint = "https://s3.amazonaws.com"
	}

	s3AccessKey := os.Getenv("S3_ACCESS_KEY")
	s3SecretKey := os.Getenv("S3_SECRET_KEY")
	s3Region := os.Getenv("S3_REGION")
	if s3Region == "" {
		s3Region = "us-east-1"
	}

	sess, err := session.NewSession(&aws.Config{
		Endpoint:    aws.String(s3Endpoint),
		Region:      aws.String(s3Region),
		Credentials: credentials.NewStaticCredentials(s3AccessKey, s3SecretKey, ""),
	})
	if err != nil {
		return err
	}

	S3Client = s3.New(sess)
	return nil
}

func UploadToS3(bucket, key string, data []byte) error {
	_, err := S3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	return err
}

func GeneratePresignedPutURL(bucket, keyPrefix string, expiry time.Duration) (string, error) {
	key := strings.TrimSuffix(keyPrefix, "/")
	req, _ := S3Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return req.Presign(expiry)
}

func GeneratePresignedGetURL(bucket, key string, expiry time.Duration) (string, error) {
	req, _ := S3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return req.Presign(expiry)
}
