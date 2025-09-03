package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

func buildClient(ctx context.Context, endpoint, region, accessKey, secretKey string) (*awss3.Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("load_config_failed: %w", err)
	}

	ep := strings.TrimRight(endpoint, "/")

	return awss3.NewFromConfig(cfg, func(o *awss3.Options) {
		o.Region = region
		o.UsePathStyle = true
		o.BaseEndpoint = aws.String(ep)
	}), nil
}

func ValidateConfig(endpoint, region, bucket, accessKey, secretKey string) error {
	if endpoint == "" {
		return fmt.Errorf("endpoint_is_required")
	}
	if region == "" {
		return fmt.Errorf("region_is_required")
	}
	if bucket == "" {
		return fmt.Errorf("bucket_is_required")
	}
	if accessKey == "" {
		return fmt.Errorf("accessKey_is_required")
	}
	if secretKey == "" {
		return fmt.Errorf("secretKey_is_required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	s3, err := buildClient(ctx, endpoint, region, accessKey, secretKey)
	if err != nil {
		return err
	}

	var missing []string
	markMissing := func(perm string, e error) {
		var nfe *types.NotFound
		if errors.As(e, &nfe) {
			missing = append(missing, fmt.Sprintf("%s (bucket_not_found)", perm))
			return
		}
		missing = append(missing, fmt.Sprintf("%s (%v)", perm, e))
	}

	if _, err := s3.HeadBucket(ctx, &awss3.HeadBucketInput{Bucket: &bucket}); err != nil {
		markMissing("bucket:access", err)
	}

	if _, err := s3.ListObjectsV2(ctx, &awss3.ListObjectsV2Input{
		Bucket:  &bucket,
		MaxKeys: aws.Int32(1),
	}); err != nil {
		markMissing("s3:ListBucket", err)
	}

	key := "sylve-permcheck-" + uuid.NewString() + ".txt"
	body := strings.NewReader("permission check")

	if _, err := s3.PutObject(ctx, &awss3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   body,
	}); err != nil {
		markMissing("s3:PutObject", err)
	} else {
		got, err := s3.GetObject(ctx, &awss3.GetObjectInput{
			Bucket: &bucket,
			Key:    &key,
		})
		if err != nil {
			markMissing("s3:GetObject", err)
		} else {
			_, _ = io.Copy(io.Discard, got.Body)
			_ = got.Body.Close()
		}

		if _, err := s3.DeleteObject(ctx, &awss3.DeleteObjectInput{
			Bucket: &bucket,
			Key:    &key,
		}); err != nil {
			markMissing("s3:DeleteObject", err)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("s3_config_validation_failed: %s", strings.Join(missing, ", "))
	}
	return nil
}

type countingReader struct {
	r io.Reader
	n int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	return n, err
}

func Put(endpoint, region, bucket, accessKey, secretKey, key string, body io.Reader) (etag string, size int64, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	s3, err := buildClient(ctx, endpoint, region, accessKey, secretKey)
	if err != nil {
		return "", 0, err
	}

	cr := &countingReader{r: body}
	out, err := s3.PutObject(ctx, &awss3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   cr,
	})
	if err != nil {
		return "", 0, fmt.Errorf("put_failed: %w", err)
	}

	return aws.ToString(out.ETag), cr.n, nil
}

func Delete(endpoint, region, bucket, accessKey, secretKey, key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s3, err := buildClient(ctx, endpoint, region, accessKey, secretKey)
	if err != nil {
		return err
	}

	if _, err := s3.DeleteObject(ctx, &awss3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}); err != nil {
		return fmt.Errorf("delete_failed: %w", err)
	}
	return nil
}
