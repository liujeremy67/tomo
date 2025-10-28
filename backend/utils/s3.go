package utils

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func UploadToS3(fileBytes []byte, filename, contentType string) (string, error) {
	region := os.Getenv("AWS_REGION")
	bucket := os.Getenv("S3_BUCKET_NAME")

	// 1. Load the AWS configuration
	// This automatically reads credentials and region from environment variables
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		return "", fmt.Errorf("failed to load aws config: %w", err)
	}

	// 2. Create an S3 client from the configuration
	svc := s3.NewFromConfig(cfg)

	// Define the object key (path in the bucket)
	key := fmt.Sprintf("uploads/%s", filename)

	// 3. Upload the file
	_, err = svc.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(fileBytes),
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPublicRead, // Use v2 enum
	})

	if err != nil {
		return "", err
	}

	// 4. Construct the public URL (using the region-specific format)
	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, key)
	return url, nil
}
