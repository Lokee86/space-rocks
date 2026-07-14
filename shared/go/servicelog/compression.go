package servicelog

import (
	"compress/gzip"
	"io"
	"os"
)

func compressArchivedSegment(sourcePath, destinationPath string, dependencies runtimeDependencies) (err error) {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := source.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	tempPath := destinationPath + ".tmp"
	tempFile, err := dependencies.openFile(tempPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}

	cleanupTemp := true
	defer func() {
		if tempFile != nil {
			_ = tempFile.Close()
		}
		if cleanupTemp {
			_ = dependencies.remove(tempPath)
		}
	}()

	gzipWriter := gzip.NewWriter(tempFile)
	if _, err = io.Copy(gzipWriter, source); err != nil {
		_ = gzipWriter.Close()
		return err
	}
	if err = gzipWriter.Close(); err != nil {
		return err
	}
	if err = tempFile.Sync(); err != nil {
		return err
	}
	if err = tempFile.Close(); err != nil {
		return err
	}
	tempFile = nil
	if err = dependencies.rename(tempPath, destinationPath); err != nil {
		return err
	}
	cleanupTemp = false
	return nil
}
