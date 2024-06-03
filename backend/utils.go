package backend

import (
	"archive/zip"
	"fmt"
	"github.com/rangwea/swallows/assets"
	"golang.org/x/exp/slog"
	"io"
	"os"
	"os/exec"
	"path"
	"runtime"
)

var commands = map[string]string{
	"windows": "start",
	"darwin":  "open",
	"linux":   "xdg-open",
}

func OpenBrowser(uri string) error {
	run, ok := commands[runtime.GOOS]
	if !ok {
		return fmt.Errorf("don't know how to open things on %s platform", runtime.GOOS)
	}

	cmd := exec.Command(run, uri)
	return cmd.Start()
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CopyAsset(src string, dst string, fileModel ...os.FileMode) (err error) {
	hb, err := assets.Asserts.ReadFile(src)
	if err != nil {
		return
	}
	fm := os.ModePerm
	if fileModel != nil {
		fm = fileModel[0]
	}
	err = os.WriteFile(dst, hb, fm)
	if err != nil {
		return err
	}
	return nil
}

func CopyFile(src string, dst string, fileModel ...os.FileMode) (err error) {
	hb, err := os.ReadFile(src)
	if err != nil {
		return
	}
	fm := os.ModePerm
	if fileModel != nil {
		fm = fileModel[0]
	}
	err = os.WriteFile(dst, hb, fm)
	if err != nil {
		return err
	}
	return nil
}

func UnZip(src string, dst string) error {
	zr, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func(zr *zip.ReadCloser) {
		err := zr.Close()
		if err != nil {
			slog.Error("close zip fail", err)
		}
	}(zr)

	if dst != "" {
		if err := os.MkdirAll(dst, 0755); err != nil {
			return err
		}
	}

	for _, file := range zr.File {
		p := path.Join(dst, file.Name)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(p, file.Mode()); err != nil {
				return err
			}
			continue
		}

		fr, err := file.Open()
		if err != nil {
			return err
		}

		fw, err := os.OpenFile(p, os.O_CREATE|os.O_RDWR|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		_, err = io.Copy(fw, fr)
		if err != nil {
			return err
		}

		fw.Close()
		fr.Close()
	}
	return nil
}
