// Package archive provides atomic extraction of .zip and .tar.gz archives.
// Extraction is atomic: files are first written to a temporary directory,
// then renamed into the final destination on success.
package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mahtdy/phpvm/internal/cli"
	"github.com/mahtdy/phpvm/internal/logger"
)

const (
	// maxFileSize guards against zip bomb / decompression attacks (2 GB).
	maxFileSize = 2 << 30
)

// Extract detects the archive type from the filename and extracts it to destDir.
// Extraction is atomic: a temp dir is used and renamed on success.
func Extract(archivePath, destDir string) error {
	switch {
	case strings.HasSuffix(archivePath, ".zip"):
		return extractZip(archivePath, destDir)
	case strings.HasSuffix(archivePath, ".tar.gz") || strings.HasSuffix(archivePath, ".tgz"):
		return extractTarGz(archivePath, destDir)
	default:
		return cli.NewArchiveError(
			fmt.Sprintf("unsupported archive format: %s", filepath.Base(archivePath)), nil,
		)
	}
}

// extractZip extracts a .zip archive atomically.
func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return cli.NewArchiveError("failed to open zip archive", err)
	}
	defer r.Close()

	tmpDir, err := atomicTempDir(destDir)
	if err != nil {
		return err
	}

	logger.Debug("Extracting ZIP", "src", zipPath, "dest", destDir)

	for _, f := range r.File {
		if err := extractZipEntry(f, tmpDir); err != nil {
			_ = os.RemoveAll(tmpDir)
			return err
		}
	}

	return atomicFinish(tmpDir, destDir)
}

func extractZipEntry(f *zip.File, destDir string) error {
	// Sanitise path to prevent zip-slip.
	target, err := sanitisePath(destDir, f.Name)
	if err != nil {
		return err
	}

	if f.FileInfo().IsDir() {
		return os.MkdirAll(target, 0o755)
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return cli.NewArchiveError("failed to create directory", err)
	}

	rc, err := f.Open()
	if err != nil {
		return cli.NewArchiveError(fmt.Sprintf("failed to open zip entry %s", f.Name), err)
	}
	defer rc.Close()

	return writeFile(target, rc, f.Mode())
}

// extractTarGz extracts a .tar.gz archive atomically.
func extractTarGz(archivePath, destDir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return cli.NewArchiveError("failed to open archive", err)
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return cli.NewArchiveError("failed to create gzip reader", err)
	}
	defer gzr.Close()

	tmpDir, err := atomicTempDir(destDir)
	if err != nil {
		return err
	}

	logger.Debug("Extracting TAR.GZ", "src", archivePath, "dest", destDir)

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			_ = os.RemoveAll(tmpDir)
			return cli.NewArchiveError("failed to read tar entry", err)
		}

		if err := extractTarEntry(header, tr, tmpDir); err != nil {
			_ = os.RemoveAll(tmpDir)
			return err
		}
	}

	return atomicFinish(tmpDir, destDir)
}

func extractTarEntry(header *tar.Header, r io.Reader, destDir string) error {
	target, err := sanitisePath(destDir, header.Name)
	if err != nil {
		return err
	}

	switch header.Typeflag {
	case tar.TypeDir:
		return os.MkdirAll(target, 0o755)
	case tar.TypeReg:
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return cli.NewArchiveError("failed to create directory", err)
		}
		return writeFile(target, r, header.FileInfo().Mode())
	case tar.TypeSymlink:
		// Sanitise symlink target as well.
		linkTarget, err := sanitisePath(destDir, header.Linkname)
		if err != nil {
			return err
		}
		return os.Symlink(linkTarget, target)
	default:
		// Skip unsupported entry types (devices, etc.).
		return nil
	}
}

// writeFile copies src into a new file at path with given permissions.
func writeFile(path string, src io.Reader, mode os.FileMode) error {
	dst, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return cli.NewArchiveError(fmt.Sprintf("cannot create file %s", path), err)
	}
	defer dst.Close()

	limited := io.LimitReader(src, maxFileSize)
	if _, err := io.Copy(dst, limited); err != nil {
		return cli.NewArchiveError(fmt.Sprintf("failed to write %s", path), err)
	}
	return nil
}

// sanitisePath prevents path traversal (zip-slip) by ensuring the joined
// path is inside destDir.
func sanitisePath(destDir, entry string) (string, error) {
	// Clean slashes.
	entry = filepath.FromSlash(entry)
	// Strip any leading path separators.
	entry = strings.TrimLeft(entry, string(os.PathSeparator)+"/")

	target := filepath.Join(destDir, entry)

	// Ensure target is under destDir.
	if !strings.HasPrefix(target+string(os.PathSeparator), destDir+string(os.PathSeparator)) {
		return "", cli.NewArchiveError(
			fmt.Sprintf("zip-slip detected: entry %q escapes destination", entry), nil,
		)
	}
	return target, nil
}

// atomicTempDir creates a temporary directory adjacent to destDir.
func atomicTempDir(destDir string) (string, error) {
	parent := filepath.Dir(destDir)
	tmp, err := os.MkdirTemp(parent, ".phpvm-extract-*")
	if err != nil {
		return "", cli.NewArchiveError("failed to create temp extraction dir", err)
	}
	return tmp, nil
}

// atomicFinish renames tmpDir to destDir.
// If destDir already exists it is removed first.
func atomicFinish(tmpDir, destDir string) error {
	if err := os.RemoveAll(destDir); err != nil {
		_ = os.RemoveAll(tmpDir)
		return cli.NewArchiveError("failed to remove existing destination", err)
	}
	if err := os.Rename(tmpDir, destDir); err != nil {
		_ = os.RemoveAll(tmpDir)
		return cli.NewArchiveError("failed to move extracted files to destination", err)
	}
	return nil
}
