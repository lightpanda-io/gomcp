package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

func cleanup(_ context.Context) error {
	// get the dir
	dir, err := configdir()
	if err != nil {
		return fmt.Errorf("get config dir: %w", err)
	}
	bin := binfilename(dir)

	slog.Debug("remove lightpanda browser", slog.String("name", bin))

	if err := os.Remove(bin); err != nil {
		return fmt.Errorf("remove browser: %w", err)
	}

	return nil
}

func download(ctx context.Context) error {
	url, err := nightlyURL()
	if err != nil {
		return fmt.Errorf("get nightly url: %w", err)
	}

	// prepare the dir
	dir, err := configdir()
	if err != nil {
		return fmt.Errorf("get config dir: %w", err)
	}

	// create the dir if not exists.
	if err := os.Mkdir(dir, 0775); err != nil {
		// ignore existing dir.
		if !errors.Is(err, os.ErrExist) {
			return fmt.Errorf("create dir: %w", err)
		}
	}

	bin := binfilename(dir)

	// open the file
	f, err := os.OpenFile(bin, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0775)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	// download the file
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("create http req: %w", err)
	}

	cli := http.Client{}
	defer cli.CloseIdleConnections()

	slog.Info("start lightpanda browser download")

	resp, err := cli.Do(req)
	if err != nil {
		return fmt.Errorf("do req: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("copy file: %w", err)
	}

	slog.Debug("lightpanda browser downloaded", slog.String("name", bin))

	return nil
}

const releaseURL = "https://github.com/lightpanda-io/browser/releases/download/nightly"

// nightlyURL return the url to download the browser depending of the current
// os and arch.
func nightlyURL() (string, error) {
	var os string
	switch runtime.GOOS {
	case "linux":
		os = "linux"
	case "darwin":
		os = "macos"
	default:
		return "", fmt.Errorf("invalid os: %s", runtime.GOOS)
	}

	var arch string
	switch runtime.GOARCH {
	case "amd64":
		arch = "x86_64"
	case "arm64":
		arch = "aarch64"
	default:
		return "", fmt.Errorf("invalid arch: %s", runtime.GOARCH)
	}

	// retrieve the nightly build
	url := fmt.Sprintf("%s/lightpanda-%s-%s", releaseURL, arch, os)

	return url, nil
}

func binfilename(dir string) string {
	return filepath.Join(dir, "lightpanda")
}

func configdir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("usr config dir: %w", err)
	}

	return filepath.Join(dir, "lightpanda-gomcp"), nil
}
