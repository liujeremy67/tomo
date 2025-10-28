package utils

import (
	"bytes"
	"context"
	"fmt"
	"net/url" // for URL parsing
	"os"
	"strings"
	"time" // for timeout duration

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// NewS3Client initializes and returns a new AWS S3 client.
//
// It loads configuration, credentials, and region from environment variables.
// The context allows cancellation or timeout control (e.g., if AWS config loading hangs).
func NewS3Client(ctx context.Context) (*s3.Client, error) {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		return nil, fmt.Errorf("AWS_REGION environment variable not set")
	}

	// Load the AWS configuration with context support.
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	// Create an S3 client from the configuration.
	return s3.NewFromConfig(cfg), nil
}

// ExtractS3Key parses the S3 object key from the full public URL.
//
// Example: "https://bucket-name.s3.region.amazonaws.com/uploads/file.jpg"
// becomes: "uploads/file.jpg"
func ExtractS3Key(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// The path component starts with a leading slash, e.g., "/uploads/file.jpg".
	// We strip the leading slash to get the required object key.
	key := strings.TrimPrefix(u.Path, "/")

	if key == "" {
		return "", fmt.Errorf("S3 URL path component is empty")
	}

	return key, nil
}

// UploadToS3 uploads fileBytes to the configured S3 bucket and returns the public URL.
//
// This version uses a context with a 20-second timeout.
// If the upload takes longer or the HTTP request is canceled, the operation stops early.
func UploadToS3(parentCtx context.Context, fileBytes []byte, filename, contentType string) (string, error) {
	region := os.Getenv("AWS_REGION")
	bucket := os.Getenv("S3_BUCKET_NAME")

	if bucket == "" {
		return "", fmt.Errorf("S3_BUCKET_NAME environment variable not set")
	}

	// Create a context with a 20-second timeout derived from the parent context.
	ctx, cancel := context.WithTimeout(parentCtx, 20*time.Second)
	defer cancel()

	// 1. Initialize the S3 client using the same context (for timeout/cancellation).
	svc, err := NewS3Client(ctx)
	if err != nil {
		return "", err // NewS3Client already wraps the error
	}

	// Define the object key (path in the bucket).
	// CAN MAKE THIS CONFIGURABLE LATER TODO
	key := fmt.Sprintf("uploads/%s", filename)

	// 2. Upload the file. If the context times out, this call is canceled automatically.
	_, err = svc.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(fileBytes),
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPublicRead,
	})
	if err != nil {
		return "", fmt.Errorf("failed to put S3 object: %w", err)
	}

	// 3. Construct and return the public URL.
	publicURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, key)
	return publicURL, nil
}

// DeleteFromS3 deletes a file from the configured S3 bucket using its public URL.
//
// It first extracts the object key (e.g. "uploads/file.jpg") from the full URL,
// then calls AWS S3's DeleteObject API.
// The call respects a 20-second timeout and can be canceled if the HTTP request ends.
func DeleteFromS3(parentCtx context.Context, fileURL string) error {
	bucket := os.Getenv("S3_BUCKET_NAME")
	if bucket == "" {
		return fmt.Errorf("S3_BUCKET_NAME environment variable not set")
	}

	ctx, cancel := context.WithTimeout(parentCtx, 20*time.Second)
	defer cancel()

	svc, err := NewS3Client(ctx)
	if err != nil {
		return err
	}

	key, err := ExtractS3Key(fileURL)
	if err != nil {
		return err
	}

	_, err = svc.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete S3 object: %w", err)
	}

	return nil
}
