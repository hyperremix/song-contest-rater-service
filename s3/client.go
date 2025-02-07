package s3

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	presignedURLExpiration = 15 * time.Minute
)

type Client struct {
	region     string
	bucketName string
	s3Client   *s3.S3
}

func New() *Client {
	region := os.Getenv("SONGCONTESTRATERSERVICE_AWS_REGION")
	bucketName := os.Getenv("SONGCONTESTRATERSERVICE_AWS_BUCKET_NAME")

	session := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("SONGCONTESTRATERSERVICE_AWS_ACCESS_KEY_ID"),
			os.Getenv("SONGCONTESTRATERSERVICE_AWS_SECRET_ACCESS_KEY"),
			""),
	}))

	return &Client{
		region:     region,
		bucketName: bucketName,
		s3Client:   s3.New(session),
	}
}

type GetPresignedUrlResponse struct {
	PresignedURL string `json:"presigned_url"`
	ImageURL     string `json:"image_url"`
}

func (c *Client) GetPresignedURL(ctx context.Context, fileName, contentType string) (GetPresignedUrlResponse, error) {
	completeFileName := fmt.Sprintf("profile-pictures/%s", fileName)

	req, _ := c.s3Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket:      aws.String(c.bucketName),
		Key:         aws.String(completeFileName),
		ContentType: aws.String(contentType),
	})

	url, _, err := req.PresignRequest(presignedURLExpiration)
	if err != nil {
		return GetPresignedUrlResponse{}, err
	}

	return GetPresignedUrlResponse{
		PresignedURL: url,
		ImageURL:     fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", c.bucketName, c.region, completeFileName),
	}, nil
}
