package storage

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type S3Config struct{ Endpoint, Region, Bucket, AccessKey, SecretKey, PublicURL string }
type S3 struct {
	endpoint                                        *url.URL
	region, bucket, accessKey, secretKey, publicURL string
	client                                          *http.Client
	now                                             func() time.Time
}

func NewS3(cfg S3Config) (*S3, error) {
	if cfg.Endpoint == "" || cfg.Region == "" || cfg.Bucket == "" || cfg.AccessKey == "" || cfg.SecretKey == "" || cfg.PublicURL == "" {
		return nil, fmt.Errorf("%w: incomplete S3 configuration", ErrUnavailable)
	}
	endpoint, err := url.Parse(cfg.Endpoint)
	if err != nil || endpoint.Host == "" || (endpoint.Scheme != "https" && endpoint.Scheme != "http") || endpoint.Path != "" {
		return nil, errors.New("OBJECT_STORAGE_ENDPOINT must be an HTTP(S) origin without a path")
	}
	publicURL, err := url.Parse(cfg.PublicURL)
	if err != nil || publicURL.Scheme != "https" || publicURL.Host == "" {
		return nil, errors.New("OBJECT_STORAGE_PUBLIC_URL must be an HTTPS URL")
	}
	return &S3{endpoint: endpoint, region: cfg.Region, bucket: cfg.Bucket, accessKey: cfg.AccessKey, secretKey: cfg.SecretKey, publicURL: strings.TrimRight(cfg.PublicURL, "/"), client: &http.Client{Timeout: 60 * time.Second}, now: time.Now}, nil
}

func (s *S3) Upload(ctx context.Context, input UploadInput) (Object, error) {
	key, err := safeObjectKey(input.Key)
	if err != nil || input.Size < 0 || input.Size > 5<<20 {
		return Object{}, errors.New("invalid object key or size")
	}
	payload, err := io.ReadAll(io.LimitReader(input.Body, input.Size+1))
	if err != nil || int64(len(payload)) != input.Size {
		return Object{}, errors.New("object size does not match input")
	}
	headers := http.Header{"Content-Type": {input.ContentType}, "Cache-Control": {"public, max-age=31536000, immutable"}}
	if err = s.request(ctx, http.MethodPut, key, payload, headers); err != nil {
		return Object{}, err
	}
	return Object{Key: key, URL: s.PublicURL(key)}, nil
}
func (s *S3) Delete(ctx context.Context, key string) error {
	key, err := safeObjectKey(key)
	if err != nil {
		return err
	}
	return s.request(ctx, http.MethodDelete, key, nil, nil)
}
func (s *S3) PublicURL(key string) string { return s.publicURL + "/" + escapePath(key) }

func safeObjectKey(value string) (string, error) {
	key := strings.TrimLeft(path.Clean("/"+value), "/")
	if key == "." || key == "" || key != value {
		return "", errors.New("invalid object key")
	}
	return key, nil
}
func (s *S3) request(ctx context.Context, method, key string, payload []byte, headers http.Header) error {
	requestURL := *s.endpoint
	requestURL.Path = "/" + s.bucket + "/" + key
	request, err := http.NewRequestWithContext(ctx, method, requestURL.String(), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	for name, values := range headers {
		request.Header[name] = values
	}
	s.sign(request, payload, s.now().UTC())
	response, err := s.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 2048))
		return fmt.Errorf("S3 %s failed: status=%d body=%s", method, response.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}
func (s *S3) sign(request *http.Request, payload []byte, at time.Time) {
	date, timestamp, payloadHash := at.Format("20060102"), at.Format("20060102T150405Z"), sha256Hex(payload)
	request.Header.Set("x-amz-date", timestamp)
	request.Header.Set("x-amz-content-sha256", payloadHash)
	names := []string{"host", "x-amz-content-sha256", "x-amz-date"}
	values := map[string]string{"host": request.URL.Host, "x-amz-content-sha256": payloadHash, "x-amz-date": timestamp}
	if value := request.Header.Get("Content-Type"); value != "" {
		names = append([]string{"content-type"}, names...)
		values["content-type"] = strings.Join(strings.Fields(value), " ")
	}
	if value := request.Header.Get("Cache-Control"); value != "" {
		names = append([]string{"cache-control"}, names...)
		values["cache-control"] = strings.Join(strings.Fields(value), " ")
	}
	var canonicalHeaders strings.Builder
	for _, name := range names {
		canonicalHeaders.WriteString(name + ":" + values[name] + "\n")
	}
	signedHeaders := strings.Join(names, ";")
	canonicalRequest := request.Method + "\n" + request.URL.EscapedPath() + "\n\n" + canonicalHeaders.String() + "\n" + signedHeaders + "\n" + payloadHash
	scope := date + "/" + s.region + "/s3/aws4_request"
	stringToSign := "AWS4-HMAC-SHA256\n" + timestamp + "\n" + scope + "\n" + sha256Hex([]byte(canonicalRequest))
	dateKey := hmacSHA256([]byte("AWS4"+s.secretKey), date)
	regionKey := hmacSHA256(dateKey, s.region)
	serviceKey := hmacSHA256(regionKey, "s3")
	signingKey := hmacSHA256(serviceKey, "aws4_request")
	signature := hex.EncodeToString(hmacSHA256(signingKey, stringToSign))
	request.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential="+s.accessKey+"/"+scope+", SignedHeaders="+signedHeaders+", Signature="+signature)
}
func sha256Hex(value []byte) string { hash := sha256.Sum256(value); return hex.EncodeToString(hash[:]) }
func hmacSHA256(key []byte, value string) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(value))
	return mac.Sum(nil)
}
func escapePath(value string) string {
	parts := strings.Split(value, "/")
	for index := range parts {
		parts[index] = url.PathEscape(parts[index])
	}
	return strings.Join(parts, "/")
}
