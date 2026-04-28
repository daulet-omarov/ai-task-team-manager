package uploader

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscreds "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const (
	MaxFileSize       = 5 << 20   // 5 MB  (photos)
	MaxAttachmentSize = 200 << 20 // 200 MB
	UploadsDir        = "./uploads"
)

var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

var allowedVideoTypes = map[string]string{
	"video/mp4":       ".mp4",
	"video/webm":      ".webm",
	"video/ogg":       ".ogv",
	"video/quicktime": ".mov",
	"video/x-msvideo": ".avi",
}

var allowedExtensions = map[string]string{
	".jpg": ".jpg", ".jpeg": ".jpg",
	".png": ".png", ".gif": ".gif", ".webp": ".webp",
	".mp4": ".mp4", ".webm": ".webm",
	".ogv": ".ogv", ".mov": ".mov", ".avi": ".avi",
}

// Config holds S3/DO Spaces credentials. Leave Bucket empty to use local storage.
type Config struct {
	Endpoint  string // https://blr1.digitaloceanspaces.com
	Region    string // blr1
	AccessKey string
	SecretKey string
	Bucket    string
	PublicURL string // https://BUCKET.blr1.digitaloceanspaces.com
}

type Uploader struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

var instance *Uploader

// Init initialises S3 storage. Call once at startup. If cfg.Bucket is empty, falls back to local disk.
func Init(cfg Config) {
	if cfg.Bucket == "" {
		return
	}
	client := s3.NewFromConfig(aws.Config{
		Region:      cfg.Region,
		Credentials: awscreds.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
	}, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
	})
	instance = &Uploader{
		client:    client,
		bucket:    cfg.Bucket,
		publicURL: strings.TrimRight(cfg.PublicURL, "/"),
	}
}

func (u *Uploader) put(key, contentType string, body io.Reader, size int64) (string, error) {
	_, err := u.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:        aws.String(u.bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
		ACL:           s3types.ObjectCannedACLPublicRead,
	})
	if err != nil {
		return "", err
	}
	return u.publicURL + "/" + key, nil
}

func randomFilename(ext string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", errors.New("failed to generate filename")
	}
	return hex.EncodeToString(b) + ext, nil
}

// SavePhoto saves a profile image. Returns ("", nil) when fh is nil.
func SavePhoto(fh *multipart.FileHeader) (string, error) {
	if fh == nil {
		return "", nil
	}
	if fh.Size > MaxFileSize {
		return "", errors.New("file too large (max 5 MB)")
	}

	file, err := fh.Open()
	if err != nil {
		return "", errors.New("cannot open uploaded file")
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", errors.New("cannot read file")
	}
	mimeType := http.DetectContentType(buf[:n])

	ext, ok := allowedImageTypes[mimeType]
	if !ok {
		origExt := strings.ToLower(filepath.Ext(fh.Filename))
		mapped, found := map[string]string{".jpg": ".jpg", ".jpeg": ".jpg", ".png": ".png", ".gif": ".gif", ".webp": ".webp"}[origExt]
		if !found {
			return "", errors.New("unsupported file type (allowed: jpeg, png, gif, webp)")
		}
		ext = mapped
	}

	filename, err := randomFilename(ext)
	if err != nil {
		return "", err
	}

	body := io.MultiReader(bytes.NewReader(buf[:n]), file)

	if instance != nil {
		return instance.put(filename, mimeType, body, fh.Size)
	}
	return saveLocal(filename, buf[:n], file)
}

// SaveFile saves an image or video attachment. Returns ("", nil) when fh is nil.
func SaveFile(fh *multipart.FileHeader) (string, error) {
	if fh == nil {
		return "", nil
	}
	if fh.Size > MaxAttachmentSize {
		return "", errors.New("file too large (max 200 MB)")
	}

	file, err := fh.Open()
	if err != nil {
		return "", errors.New("cannot open uploaded file")
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", errors.New("cannot read file")
	}
	sniffedMime := http.DetectContentType(buf[:n])

	ext, ok := allowedImageTypes[sniffedMime]
	if !ok {
		origExt := strings.ToLower(filepath.Ext(fh.Filename))
		if mapped, found := allowedExtensions[origExt]; found {
			ext, ok = mapped, true
		}
	}
	if !ok {
		if mapped, found := allowedVideoTypes[fh.Header.Get("Content-Type")]; found {
			ext, ok = mapped, true
		}
	}
	if !ok {
		if mapped, found := allowedVideoTypes[sniffedMime]; found {
			ext, ok = mapped, true
		}
	}
	if !ok {
		return "", errors.New("unsupported file type (allowed: jpeg, png, gif, webp, mp4, webm, mov, avi, ogv)")
	}

	filename, err := randomFilename(ext)
	if err != nil {
		return "", err
	}

	body := io.MultiReader(bytes.NewReader(buf[:n]), file)

	if instance != nil {
		return instance.put(filename, sniffedMime, body, fh.Size)
	}
	return saveLocal(filename, buf[:n], file)
}

// SaveAny saves any file type (chat attachments). Returns ("", "", nil) when fh is nil.
func SaveAny(fh *multipart.FileHeader) (path string, mimeType string, err error) {
	if fh == nil {
		return "", "", nil
	}
	if fh.Size > MaxAttachmentSize {
		return "", "", errors.New("file too large (max 200 MB)")
	}

	file, err := fh.Open()
	if err != nil {
		return "", "", errors.New("cannot open uploaded file")
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", "", errors.New("cannot read file")
	}
	mimeType = http.DetectContentType(buf[:n])
	if ct := fh.Header.Get("Content-Type"); ct != "" && mimeType == "application/octet-stream" {
		mimeType = ct
	}

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if ext == "" {
		ext = ".bin"
	}

	filename, err := randomFilename(ext)
	if err != nil {
		return "", "", err
	}

	body := io.MultiReader(bytes.NewReader(buf[:n]), file)

	if instance != nil {
		p, err := instance.put(filename, mimeType, body, fh.Size)
		return p, mimeType, err
	}
	p, err := saveLocal(filename, buf[:n], file)
	return p, mimeType, err
}

// SaveFromURL downloads a remote file and stores it. Returns path, size, error.
func SaveFromURL(remoteURL, originalName string) (string, int64, error) {
	resp, err := http.Get(remoteURL) //nolint:gosec
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", 0, fmt.Errorf("remote file responded %d", resp.StatusCode)
	}

	ext := strings.ToLower(filepath.Ext(originalName))
	if ext == "" {
		ext = ".bin"
	}

	filename, err := randomFilename(ext)
	if err != nil {
		return "", 0, err
	}

	if instance != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", 0, errors.New("failed to read remote file")
		}
		contentType := http.DetectContentType(body)
		path, err := instance.put(filename, contentType, bytes.NewReader(body), int64(len(body)))
		return path, int64(len(body)), err
	}

	if err := os.MkdirAll(UploadsDir, 0755); err != nil {
		return "", 0, errors.New("cannot create uploads directory")
	}
	dst, err := os.Create(filepath.Join(UploadsDir, filename))
	if err != nil {
		return "", 0, errors.New("failed to create file")
	}
	defer dst.Close()
	written, err := io.Copy(dst, resp.Body)
	if err != nil {
		return "", 0, errors.New("failed to write file")
	}
	return "/uploads/" + filename, written, nil
}

// saveLocal writes buf + remaining reader to UploadsDir and returns "/uploads/filename".
func saveLocal(filename string, buf []byte, rest io.Reader) (string, error) {
	if err := os.MkdirAll(UploadsDir, 0755); err != nil {
		return "", errors.New("cannot create uploads directory")
	}
	dst, err := os.Create(filepath.Join(UploadsDir, filename))
	if err != nil {
		return "", errors.New("failed to save file")
	}
	defer dst.Close()
	if _, err := dst.Write(buf); err != nil {
		return "", errors.New("failed to write file")
	}
	if _, err := io.Copy(dst, rest); err != nil {
		return "", errors.New("failed to write file")
	}
	return "/uploads/" + filename, nil
}

// FullURL builds an absolute URL from a path like "/uploads/foo.png" or returns S3 URLs unchanged.
func FullURL(r *http.Request, path string) string {
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}
	return scheme + "://" + r.Host + path
}
