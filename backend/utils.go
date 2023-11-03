package backend

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/shirou/gopsutil/v3/process"
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

func CopyFile(src string, dst string, fileModel ...os.FileMode) (err error) {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	fm := os.ModePerm
	if fileModel != nil {
		fm = fileModel[0]
	}
	err = os.WriteFile(dst, data, fm)
	if err != nil {
		return err
	}
	return nil
}

func CheckProcessExist(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func KillProcessByName(name string) error {
	ps, err := process.Processes()
	if err != nil {
		return err
	}
	for _, p := range ps {
		pName, _ := p.Name()
		if strings.Contains(pName, name) {
			err = p.Kill()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func UnZip(src string, dst string) error {
	zr, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer zr.Close()

	if dst != "" {
		if err := os.MkdirAll(dst, 0755); err != nil {
			return err
		}
	}

	for _, file := range zr.File {
		path := path.Join(dst, file.Name)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(path, file.Mode()); err != nil {
				return err
			}
			continue
		}

		fr, err := file.Open()
		if err != nil {
			return err
		}

		fw, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, file.Mode())
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

func RemoveByWildcard(wildcard string) error {
	files, err := filepath.Glob(wildcard)
	if err != nil {
		return err
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return err
		}
	}
	return nil
}
