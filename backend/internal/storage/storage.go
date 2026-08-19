package storage

import (
	"context"
	"errors"
	"io"
)

var ErrUnavailable = errors.New("object storage is not configured")

type UploadInput struct {
	Key, ContentType string
	Size             int64
	Body             io.Reader
}
type Object struct{ Key, URL string }
type Provider interface {
	Upload(context.Context, UploadInput) (Object, error)
	Delete(context.Context, string) error
	PublicURL(string) string
}
type Unavailable struct{}

func (Unavailable) Upload(context.Context, UploadInput) (Object, error) {
	return Object{}, ErrUnavailable
}
func (Unavailable) Delete(context.Context, string) error { return ErrUnavailable }
func (Unavailable) PublicURL(string) string              { return "" }
