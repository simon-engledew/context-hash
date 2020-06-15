package pkg

import (
	"archive/tar"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/docker/cli/cli/command/image/build"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/idtools"
	"github.com/docker/docker/pkg/pools"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

func closeOrPanic(format string, closer io.Closer) {
	if err := closer.Close(); err != nil {
		if errors.Is(err, io.ErrClosedPipe) {
			return
		}
		panic(fmt.Errorf(format, err))
	}
}

func isDir(path string) error {
	stat, err := os.Stat(path)
	if err != nil || !stat.IsDir() {
		return fmt.Errorf("%q is not a directory", path)
	}
	return nil
}

func isFile(path string) error {
	stat, err := os.Stat(path)
	if err != nil || stat.IsDir() {
		return fmt.Errorf("%q is not a file", path)
	}
	return nil
}

func HashContext(path, dockerfile string) (string, error) {
	if err := isDir(path); err != nil {
		return "", err
	}

	if err := isFile(filepath.Join(path, dockerfile)); err != nil {
		return "", err
	}

	excludes, err := build.ReadDockerignore(path)
	if err != nil {
		return "", fmt.Errorf("failed to read dockerignore: %w", err)
	}

	if err := build.ValidateContextDirectory(path, excludes); err != nil {
		return "", fmt.Errorf("failed to validate context directory: %w", err)
	}

	excludes = append(excludes, "!"+dockerfile)

	contextTar, err := archive.TarWithOptions(path, &archive.TarOptions{
		ExcludePatterns: excludes,
		ChownOpts:       &idtools.Identity{UID: 0, GID: 0},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create tar: %w", err)
	}

	stripTime := func(header *tar.Header) {
		log.Printf("+ %s", header.Name)
		header.ModTime = time.Time{}
	}

	tarHash := sha256.New()

	if _, err = pools.Copy(tarHash, rewriteTarHeaders(contextTar, stripTime)); err != nil {
		return "", fmt.Errorf("failed to generate hash from context tar: %w", err)
	}

	return fmt.Sprintf("%x", tarHash.Sum(nil)), nil
}

func rewriteTarHeaders(inputTarStream io.ReadCloser, modifyFn func(*tar.Header)) io.ReadCloser {
	pipeReader, pipeWriter := io.Pipe()

	go func() {
		tarReader := tar.NewReader(inputTarStream)
		tarWriter := tar.NewWriter(pipeWriter)
		defer closeOrPanic("failed to close inputTarStream: %w", inputTarStream)
		defer closeOrPanic("failed to close tarWriter: %w", tarWriter)

		var err error
		var originalHeader *tar.Header
		for {
			originalHeader, err = tarReader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				if err := pipeWriter.CloseWithError(err); err != nil {
					panic(fmt.Errorf("failed to read next tar entry: %w", err))
				}
				return
			}

			modifyFn(originalHeader)

			if err := tarWriter.WriteHeader(originalHeader); err != nil {
				if err := pipeWriter.CloseWithError(err); err != nil {
					panic(fmt.Errorf("failed to write header: %w", err))
				}
				return
			}
			if _, err := pools.Copy(tarWriter, tarReader); err != nil {
				if err := pipeWriter.CloseWithError(err); err != nil {
					panic(fmt.Errorf("failed to copy tarReader: %w", err))
				}
				return
			}
		}

		if err := pipeWriter.Close(); err != nil {
			panic(fmt.Errorf("failed to close pipeWriter: %w", err))
		}
	}()
	return pipeReader
}
