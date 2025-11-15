package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"

	"haodun_manage/backend/config"
)

var (
	cosClient     *cos.Client
	cosInitErr    error
	cosClientOnce sync.Once
)

func getBucketURL() (*url.URL, error) {
	if config.AppConfig == nil {
		return nil, errors.New("config not initialized")
	}

	baseURL := strings.TrimSpace(config.AppConfig.COSBaseURL)
	if baseURL == "" {
		if config.AppConfig.COSBucket == "" || config.AppConfig.COSRegion == "" {
			return nil, errors.New("cos bucket or region not configured")
		}
		baseURL = fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config.AppConfig.COSBucket, config.AppConfig.COSRegion)
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid COS base url: %w", err)
	}
	return u, nil
}

// GetCOSClient returns singleton cos client instance.
func GetCOSClient() (*cos.Client, error) {
	cosClientOnce.Do(func() {
		if config.AppConfig == nil {
			cosInitErr = errors.New("config not initialized")
			return
		}
		if config.AppConfig.COSSecretID == "" || config.AppConfig.COSSecretKey == "" {
			cosInitErr = errors.New("cos credentials missing")
			return
		}

		bucketURL, err := getBucketURL()
		if err != nil {
			cosInitErr = err
			return
		}

		cosClient = cos.NewClient(
			&cos.BaseURL{BucketURL: bucketURL},
			&http.Client{
				Transport: &cos.AuthorizationTransport{
					SecretID:  config.AppConfig.COSSecretID,
					SecretKey: config.AppConfig.COSSecretKey,
				},
			},
		)
	})

	return cosClient, cosInitErr
}

// UploadToCOS uploads reader content to target object key.
func UploadToCOS(ctx context.Context, objectKey string, reader io.Reader, contentLength int64, contentType string) (string, error) {
	client, err := GetCOSClient()
	if err != nil {
		return "", err
	}

	opts := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{},
	}
	if contentType != "" {
		opts.ObjectPutHeaderOptions.ContentType = contentType
	}
	if contentLength >= 0 {
		opts.ObjectPutHeaderOptions.ContentLength = contentLength
	}

	if _, err := client.Object.Put(ctx, objectKey, reader, opts); err != nil {
		return "", err
	}
	return BuildCOSObjectURL(objectKey)
}

// DeleteFromCOS removes object by key.
func DeleteFromCOS(ctx context.Context, objectKey string) error {
	client, err := GetCOSClient()
	if err != nil {
		return err
	}
	_, err = client.Object.Delete(ctx, objectKey)
	return err
}

// GenerateCOSPresignedURL returns temporary url for object access.
func GenerateCOSPresignedURL(ctx context.Context, objectKey string, method string, expire time.Duration) (string, error) {
	client, err := GetCOSClient()
	if err != nil {
		return "", err
	}
	if method == "" {
		method = http.MethodGet
	}
	if expire <= 0 {
		expire = time.Duration(config.AppConfig.COSURLExpires) * time.Second
	}

	u, err := client.Object.GetPresignedURL(ctx, method, objectKey, config.AppConfig.COSSecretID, config.AppConfig.COSSecretKey, expire, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// BuildCOSObjectURL returns absolute url for object location without presign.
func BuildCOSObjectURL(objectKey string) (string, error) {
	bucketURL, err := getBucketURL()
	if err != nil {
		return "", err
	}
	clone := *bucketURL
	clone.Path = path.Join(clone.Path, objectKey)
	return clone.String(), nil
}

// CreateReusableReader clones the initial bytes and combines with source reader.
func CreateReusableReader(src io.Reader, header []byte) io.Reader {
	if len(header) == 0 {
		return src
	}
	return io.MultiReader(bytes.NewReader(header), src)
}
