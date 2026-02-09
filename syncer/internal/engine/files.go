// files.go - 파일 복사 유틸리티
// 파일을 안전하게 복사하며, 임시 파일을 사용하여 원자적 쓰기를 보장합니다.
package engine

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"
)

// copyFileContents: 파일을 src에서 dst로 안전하게 복사
// 1. 임시 파일에 먼저 복사
// 2. 권한과 수정 시간 복사
// 3. 임시 파일을 최종 위치로 이동
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
