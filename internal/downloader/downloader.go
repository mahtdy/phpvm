// Package downloader handles HTTP file downloads with progress display,
// resume support, retry logic, and SHA-256 checksum verification.
package downloader

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/mahtdy/phpvm/internal/cli"
	"github.com/mahtdy/phpvm/internal/logger"
	"github.com/schollz/progressbar/v3"
)

const (
	// maxRetries is the number of download attempts before giving up.
	maxRetries = 3
	// retryBaseDelay is the initial backoff delay between retries.
	retryBaseDelay = 2 * time.Second
	// defaultTimeout is the total download timeout.
	defaultTimeout = 30 * time.Minute
	// chunkSize is the read buffer size.
	chunkSize = 32 * 1024
)

// Options configures a download operation.
type Options struct {
	// URL is the download source.
	URL string
	// DestDir is the directory where the file will be saved.
	DestDir string
	// Filename overrides the derived filename (optional).
	Filename string
	// ExpectedSHA256 is the expected hex checksum (empty = skip verify).
	ExpectedSHA256 string
	// Timeout overrides the default download timeout.
	Timeout time.Duration
	// NoProgress disables the terminal progress bar.
	NoProgress bool
}

// Result holds information about a completed download.
type Result struct {
	// Path is the absolute path to the downloaded file.
	Path string
	// SHA256 is the hex-encoded checksum of the downloaded file.
	SHA256 string
	// Size is the number of bytes downloaded.
	Size int64
}

// Download downloads a file according to opts, retrying on transient errors.
// It verifies the SHA-256 checksum if Options.ExpectedSHA256 is set.
func Download(ctx context.Context, opts Options) (*Result, error) {
	if opts.URL == "" {
		return nil, cli.NewValidationError("download URL must not be empty", nil)
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	if err := os.MkdirAll(opts.DestDir, 0o750); err != nil {
		return nil, cli.NewPermissionError("cannot create download directory", err)
	}

	filename := opts.Filename
	if filename == "" {
		filename = filepath.Base(opts.URL)
	}
	destPath := filepath.Join(opts.DestDir, filename)

	var result *Result
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			delay := retryBaseDelay * time.Duration(attempt-1)
			logger.Warn(fmt.Sprintf("Download attempt %d/%d failed, retrying in %s...",
				attempt-1, maxRetries, delay))
			select {
			case <-ctx.Done():
				return nil, cli.NewNetworkError("download cancelled", ctx.Err())
			case <-time.After(delay):
			}
		}

		result, lastErr = downloadOnce(ctx, opts, destPath, timeout)
		if lastErr == nil {
			break
		}
		logger.Debug("download attempt failed", "attempt", attempt, "error", lastErr)
	}

	if lastErr != nil {
		_ = os.Remove(destPath)
		return nil, lastErr
	}

	// Checksum verification.
	if opts.ExpectedSHA256 != "" {
		if err := verifySHA256(destPath, opts.ExpectedSHA256); err != nil {
			_ = os.Remove(destPath)
			return nil, err
		}
		logger.Info("Checksum verified", "file", filename)
	}

	return result, nil
}

// downloadOnce performs a single download attempt, supporting resume via
// HTTP Range requests if a partial file already exists.
func downloadOnce(ctx context.Context, opts Options, destPath string, timeout time.Duration) (*Result, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Check for existing partial download.
	var startByte int64
	if info, err := os.Stat(destPath); err == nil {
		startByte = info.Size()
	}

	req, err := http.NewRequestWithContext(timeoutCtx, http.MethodGet, opts.URL, nil)
	if err != nil {
		return nil, cli.NewNetworkError("failed to build HTTP request", err)
	}

	if startByte > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", startByte))
		logger.Debug("Resuming download", "offset", startByte)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, cli.NewNetworkError(
			fmt.Sprintf("HTTP request failed for %s", opts.URL), err,
		)
	}
	defer resp.Body.Close()

	// 200 = full file, 206 = partial content (resume).
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return nil, cli.NewNetworkError(
			fmt.Sprintf("HTTP %d downloading %s", resp.StatusCode, opts.URL), nil,
		)
	}

	// If server doesn't support Range, restart from scratch.
	flags := os.O_CREATE | os.O_WRONLY
	if resp.StatusCode == http.StatusPartialContent {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
		startByte = 0
	}

	file, err := os.OpenFile(destPath, flags, 0o640)
	if err != nil {
		return nil, cli.NewPermissionError("cannot open destination file", err)
	}
	defer file.Close()

	// Build progress bar.
	totalSize := resp.ContentLength + startByte
	var reader io.Reader = resp.Body

	if !opts.NoProgress {
		bar := progressbar.NewOptions64(
			totalSize,
			progressbar.OptionSetDescription(fmt.Sprintf("  Downloading %s", filepath.Base(opts.URL))),
			progressbar.OptionSetWidth(40),
			progressbar.OptionShowBytes(true),
			progressbar.OptionShowCount(),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "=",
				SaucerHead:    ">",
				SaucerPadding: " ",
				BarStart:      "[",
				BarEnd:        "]",
			}),
		)
		bar.Set64(startByte) //nolint:errcheck
		reader = io.TeeReader(resp.Body, bar)
	}

	written, err := io.Copy(file, reader)
	if err != nil {
		return nil, cli.NewNetworkError("download interrupted", err)
	}

	totalWritten := startByte + written

	// Compute SHA-256 of the complete file.
	checksum, err := fileSHA256(destPath)
	if err != nil {
		return nil, cli.NewError("failed to compute checksum", err)
	}

	return &Result{
		Path:   destPath,
		SHA256: checksum,
		Size:   totalWritten,
	}, nil
}

// verifySHA256 checks that the file at path matches expectedHex.
func verifySHA256(path, expectedHex string) error {
	actual, err := fileSHA256(path)
	if err != nil {
		return err
	}
	if actual != expectedHex {
		return cli.NewError(
			fmt.Sprintf("checksum mismatch: expected %s, got %s", expectedHex, actual), nil,
		)
	}
	return nil
}

// fileSHA256 returns the lowercase hex SHA-256 of a file.
func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", cli.NewFileNotFoundError("cannot open file for checksum", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", cli.NewError("failed to hash file", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
