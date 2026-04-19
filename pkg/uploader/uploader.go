package uploader

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	MaxFileSize       = 5 << 20   // 5 MB  (photos)
	MaxAttachmentSize = 200 << 20 // 200 MB (photos + videos)
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
	// images
	".jpg": ".jpg", ".jpeg": ".jpg",
	".png": ".png", ".gif": ".gif", ".webp": ".webp",
	// videos
	".mp4": ".mp4", ".webm": ".webm",
	".ogv": ".ogv", ".mov": ".mov", ".avi": ".avi",
}

// SavePhoto saves an image file to UploadsDir and returns the public URL path.
// Returns ("", nil) if fh is nil (no file uploaded — field was omitted).
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

	// Detect MIME from first 512 bytes
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", errors.New("cannot read file")
	}
	mimeType := http.DetectContentType(buf[:n])

	ext, ok := allowedImageTypes[mimeType]
	if !ok {
		origExt := strings.ToLower(filepath.Ext(fh.Filename))
		allowed := map[string]string{
			".jpg": ".jpg", ".jpeg": ".jpg",
			".png": ".png", ".gif": ".gif", ".webp": ".webp",
		}
		if mapped, found := allowed[origExt]; found {
			ext = mapped
		} else {
			return "", errors.New("unsupported file type (allowed: jpeg, png, gif, webp)")
		}
	}

	if err := os.MkdirAll(UploadsDir, 0755); err != nil {
		return "", errors.New("cannot create uploads directory")
	}

	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		return "", errors.New("failed to generate filename")
	}
	filename := hex.EncodeToString(randBytes) + ext
	destPath := filepath.Join(UploadsDir, filename)

	dst, err := os.Create(destPath)
	if err != nil {
		return "", errors.New("failed to save file")
	}
	defer dst.Close()

	if _, err := dst.Write(buf[:n]); err != nil {
		return "", errors.New("failed to write file")
	}
	if _, err := io.Copy(dst, file); err != nil {
		return "", errors.New("failed to write file")
	}

	return "/uploads/" + filename, nil
}

// SaveFile saves any allowed file (image or video) to UploadsDir and returns the public URL path.
// Returns ("", nil) if fh is nil.
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

	// Detect MIME from first 512 bytes
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", errors.New("cannot read file")
	}
	sniffedMime := http.DetectContentType(buf[:n])

	// 1. Try sniffed MIME for images (most reliable for images)
	ext, ok := allowedImageTypes[sniffedMime]
	if !ok {
		// 2. Try file extension (most reliable for videos — Go's sniffer returns
		//    application/octet-stream for QuickTime/MOV and many other video formats)
		origExt := strings.ToLower(filepath.Ext(fh.Filename))
		if mapped, found := allowedExtensions[origExt]; found {
			ext = mapped
			ok = true
		}
	}
	if !ok {
		// 3. Try the multipart Content-Type header sent by the client
		ctHeader := fh.Header.Get("Content-Type")
		if mapped, found := allowedVideoTypes[ctHeader]; found {
			ext = mapped
			ok = true
		}
	}
	if !ok {
		// 4. Last resort: sniffed MIME for video
		if mapped, found := allowedVideoTypes[sniffedMime]; found {
			ext = mapped
			ok = true
		}
	}
	if !ok {
		return "", errors.New("unsupported file type (allowed: jpeg, png, gif, webp, mp4, webm, mov, avi, ogv)")
	}

	if err := os.MkdirAll(UploadsDir, 0755); err != nil {
		return "", errors.New("cannot create uploads directory")
	}

	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		return "", errors.New("failed to generate filename")
	}
	filename := hex.EncodeToString(randBytes) + ext
	destPath := filepath.Join(UploadsDir, filename)

	dst, err := os.Create(destPath)
	if err != nil {
		return "", errors.New("failed to save file")
	}
	defer dst.Close()

	if _, err := dst.Write(buf[:n]); err != nil {
		return "", errors.New("failed to write file")
	}
	if _, err := io.Copy(dst, file); err != nil {
		return "", errors.New("failed to write file")
	}

	return "/uploads/" + filename, nil
}

// FullURL constructs an absolute URL from the incoming request and a path like "/uploads/foo.png".
// Returns "" if path is empty. Returns path unchanged if it is already an absolute URL.
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
	// X-Forwarded-Proto is set by reverse proxies (nginx, etc.)
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}
	return scheme + "://" + r.Host + path
}
