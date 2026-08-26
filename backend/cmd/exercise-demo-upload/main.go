package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	appstorage "github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const maxDemoBytes int64 = 5 << 20

var standardKeyPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type input struct {
	StandardKey string
	FilePath    string
	Duration    int
	DryRun      bool
	Confirm     bool
}

type fileInfo struct {
	File     *os.File
	Name     string
	MIMEType string
	Type     string
	Size     int64
}

func parse(args []string) (input, error) {
	fs := flag.NewFlagSet("exercise-demo-upload", flag.ContinueOnError)
	var in input
	fs.StringVar(&in.StandardKey, "standard-key", "", "standard exercise key")
	fs.StringVar(&in.FilePath, "file", "", "local MP4, WebM, GIF, JPEG, PNG or WebP")
	fs.IntVar(&in.Duration, "duration-seconds", 0, "video duration from trusted metadata")
	fs.BoolVar(&in.DryRun, "dry-run", false, "validate file and database plan")
	fs.BoolVar(&in.Confirm, "confirm", false, "upload, register and attach")
	if err := fs.Parse(args); err != nil {
		return in, err
	}
	if !standardKeyPattern.MatchString(in.StandardKey) || in.FilePath == "" || in.DryRun == in.Confirm {
		return in, errors.New("--standard-key, --file and exactly one of --dry-run/--confirm are required")
	}
	return in, nil
}

func inspectFile(path string, duration int) (fileInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return fileInfo{}, err
	}
	closeWithError := func(err error) (fileInfo, error) { _ = file.Close(); return fileInfo{}, err }
	stat, err := file.Stat()
	if err != nil {
		return closeWithError(err)
	}
	if !stat.Mode().IsRegular() || stat.Size() <= 0 || stat.Size() > maxDemoBytes {
		return closeWithError(errors.New("demo must be a regular non-empty file no larger than 5 MB"))
	}
	header := make([]byte, 512)
	n, err := io.ReadFull(file, header)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return closeWithError(err)
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return closeWithError(err)
	}
	detected := http.DetectContentType(header[:n])
	extension := strings.ToLower(filepath.Ext(path))
	byExtension := mime.TypeByExtension(extension)
	if index := strings.IndexByte(byExtension, ';'); index >= 0 {
		byExtension = byExtension[:index]
	}
	allowed := map[string]string{
		"video/mp4": "video", "video/webm": "video", "image/gif": "image",
		"image/jpeg": "image", "image/png": "image", "image/webp": "image",
	}
	mimeType := detected
	if allowed[mimeType] == "" && allowed[byExtension] != "" && detected == "application/octet-stream" {
		mimeType = byExtension
	}
	kind := allowed[mimeType]
	if kind == "" || (byExtension != "" && allowed[byExtension] != "" && byExtension != mimeType) {
		return closeWithError(fmt.Errorf("unsupported or mismatched media type: detected %s", detected))
	}
	if kind == "video" && (duration < 1 || duration > 6) {
		return closeWithError(errors.New("video requires trusted --duration-seconds between 1 and 6"))
	}
	if kind == "image" && duration != 0 {
		return closeWithError(errors.New("--duration-seconds is only valid for video"))
	}
	return fileInfo{File: file, Name: filepath.Base(path), MIMEType: mimeType, Type: kind, Size: stat.Size()}, nil
}

func storageKey(standardKey, mimeType string) string {
	extensions := map[string]string{"video/mp4": ".mp4", "video/webm": ".webm", "image/gif": ".gif", "image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}
	return "exercises/standard/" + standardKey + "/demo" + extensions[mimeType]
}

func run(args []string) error {
	in, err := parse(args)
	if err != nil {
		return err
	}
	media, err := inspectFile(in.FilePath, in.Duration)
	if err != nil {
		return err
	}
	defer media.File.Close()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return errors.New("DATABASE_URL is required")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return err
	}
	defer pool.Close()
	key := storageKey(in.StandardKey, media.MIMEType)
	var exerciseID string
	var currentMediaID *string
	err = pool.QueryRow(ctx, `SELECT id::text,demo_media_id::text FROM exercises WHERE standard_key=$1 AND owner_user_id IS NULL`, in.StandardKey).Scan(&exerciseID, &currentMediaID)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("conflict: standard exercise %s does not exist", in.StandardKey)
	}
	if err != nil {
		return err
	}
	var existingID string
	err = pool.QueryRow(ctx, `SELECT id::text FROM media_assets WHERE storage_key=$1 AND owner_user_id IS NULL`, key).Scan(&existingID)
	existed := err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	action := "attach"
	if currentMediaID != nil && *currentMediaID == existingID && existed {
		action = "unchanged"
	} else if currentMediaID != nil {
		action = "replace"
	}
	fmt.Printf("VALID: key=%s type=%s size=%d action=%s\n", in.StandardKey, media.MIMEType, media.Size, action)
	if in.DryRun || action == "unchanged" {
		return nil
	}
	if os.Getenv("OBJECT_STORAGE_PROVIDER") != "s3" {
		return fmt.Errorf("%w: OBJECT_STORAGE_PROVIDER must be s3", appstorage.ErrUnavailable)
	}
	provider, err := appstorage.NewS3(appstorage.S3Config{
		Endpoint: os.Getenv("OBJECT_STORAGE_ENDPOINT"), Region: os.Getenv("OBJECT_STORAGE_REGION"),
		Bucket: os.Getenv("OBJECT_STORAGE_BUCKET"), AccessKey: os.Getenv("OBJECT_STORAGE_ACCESS_KEY"),
		SecretKey: os.Getenv("OBJECT_STORAGE_SECRET_KEY"), PublicURL: os.Getenv("OBJECT_STORAGE_PUBLIC_URL"),
	})
	if err != nil {
		return err
	}
	object, err := provider.Upload(ctx, appstorage.UploadInput{Key: key, ContentType: media.MIMEType, Size: media.Size, Body: media.File})
	if err != nil {
		return err
	}
	cleanup := !existed
	defer func() {
		if cleanup {
			_ = provider.Delete(context.Background(), object.Key)
		}
	}()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var mediaID string
	err = tx.QueryRow(ctx, `INSERT INTO media_assets(owner_user_id,type,status,storage_provider,storage_key,url,original_filename,mime_type,size_bytes,duration_seconds) VALUES(NULL,$1,'ready','s3',$2,$3,$4,$5,$6,$7) ON CONFLICT(storage_key) DO UPDATE SET url=EXCLUDED.url,original_filename=EXCLUDED.original_filename,mime_type=EXCLUDED.mime_type,size_bytes=EXCLUDED.size_bytes,duration_seconds=EXCLUDED.duration_seconds,status='ready' WHERE media_assets.owner_user_id IS NULL RETURNING id::text`, media.Type, object.Key, object.URL, media.Name, media.MIMEType, media.Size, nullableDuration(media.Type, in.Duration)).Scan(&mediaID)
	if err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `UPDATE exercises SET demo_media_id=$2::uuid WHERE id=$1::uuid AND owner_user_id IS NULL`, exerciseID, mediaID)
	if err != nil || tag.RowsAffected() != 1 {
		return errors.New("database attach failed")
	}
	if err = tx.Commit(ctx); err != nil {
		return err
	}
	cleanup = false
	fmt.Printf("ATTACHED: key=%s media_id=%s\n", in.StandardKey, mediaID)
	return nil
}

func nullableDuration(kind string, duration int) any {
	if kind == "video" {
		return duration
	}
	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
