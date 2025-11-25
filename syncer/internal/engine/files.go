package engine

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"
)

func copyFileContents(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	tmpDst := dst + ".tmp"
	dstFile, err := os.Create(tmpDst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		dstFile.Close()
		os.Remove(tmpDst)
		return err
	}

	if err := dstFile.Sync(); err != nil {
		dstFile.Close()
		os.Remove(tmpDst)
		return err
	}

	if err := dstFile.Close(); err != nil {
		os.Remove(tmpDst)
		return err
	}

	if err := os.Chmod(tmpDst, srcInfo.Mode()); err != nil {
		os.Remove(tmpDst)
		return err
	}

	if err := os.Chtimes(tmpDst, time.Now(), srcInfo.ModTime()); err != nil {
		os.Remove(tmpDst)
		return err
	}

	if err := os.Rename(tmpDst, dst); err != nil {
		if removeErr := os.Remove(dst); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			os.Remove(tmpDst)
			return err
		}
		if err := os.Rename(tmpDst, dst); err != nil {
			os.Remove(tmpDst)
			return err
		}
	}

	return nil
}
