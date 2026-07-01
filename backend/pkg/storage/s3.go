package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

type Uploader struct {
	endpoint  string
	bucket    string
	accessKey string
	secretKey string
	region    string
	enabled   bool
	client    *http.Client
}

func NewFromEnv() *Uploader {
	endpoint := strings.TrimRight(os.Getenv("S3_ENDPOINT"), "/")
	bucket := os.Getenv("S3_BUCKET")
	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")
	region := os.Getenv("S3_REGION")
	if region == "" {
		region = "us-east-1"
	}
	enabled := strings.EqualFold(os.Getenv("S3_UPLOAD_ENABLED"), "true") &&
		endpoint != "" && bucket != "" && accessKey != "" && secretKey != ""
	return &Uploader{
		endpoint:  endpoint,
		bucket:    bucket,
		accessKey: accessKey,
		secretKey: secretKey,
		region:    region,
		enabled:   enabled,
		client:    &http.Client{Timeout: 60 * time.Second},
	}
}

func (u *Uploader) Enabled() bool {
	return u != nil && u.enabled
}

func (u *Uploader) Upload(ctx context.Context, objectKey string, contentType string, data []byte) (string, error) {
	if !u.Enabled() {
		return "", fmt.Errorf("S3 upload is not configured")
	}
	key := strings.TrimPrefix(objectKey, "/")
	url := fmt.Sprintf("%s/%s/%s", u.endpoint, u.bucket, key)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", contentType)
	req.SetBasicAuth(u.accessKey, u.secretKey)

	resp, err := u.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("S3 upload failed: %s %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return url, nil
}

func ReportObjectKey(reportName, filename string) string {
	date := time.Now().Format("2006-01-02")
	return path.Join("reports", date, reportName, filename)
}

func BackupObjectKey(filename string) string {
	date := time.Now().Format("2006-01-02")
	return path.Join("backups", date, filename)
}
