package backedup

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// initConfig will either create or copy an existing config.
func initConfig(fs afero.Fs, stdin io.Reader, logger io.Writer, confPath string) error {
	exists, err := afero.Exists(fs, confPath)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	// first check if the default backup path has a config to use and copy that
	// over if there.
	backedupPath := filepath.Join(DefaultBackupTo, DefaultBackedupFilename)
	exists, err = afero.Exists(fs, backedupPath)
	if err != nil {
		return err
	}
	reader := bufio.NewReader(stdin)
	switch exists {
	case true:
		// the backup file exists so ask to copy that for use.
		fmt.Fprintf(logger, "%s exists, copy that to ~? <Yes|No>: ", backedupPath)
		text, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if strings.TrimSpace(strings.ToLower(text)) != "yes" {
			return ErrDefaultConfigNotCreated
		}
		fmt.Fprintf(logger, "copying %s config to %s\n", backedupPath, confPath)
		data, err := afero.ReadFile(fs, backedupPath)
		if err != nil {
			return err
		}
		if err := afero.WriteFile(fs, confPath, data, 0644); err != nil {
			return err
		}
	default:
		// backuped file not found so ask to create a new one.
		fmt.Fprintf(logger, "%s does not exist, create a default one? <Yes|No>: ", confPath)
		text, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if strings.TrimSpace(strings.ToLower(text)) != "yes" {
			return ErrDefaultConfigNotCreated
		}
		fmt.Fprintf(logger, "creating %s\n", confPath)
		if err := afero.WriteFile(fs, confPath, []byte(DefaultCfg), 0644); err != nil {
			return err
		}
	}
	return nil
}

// FileCopy copies a single file from src to dst
// From here https://blog.depado.eu/post/copy-files-and-directories-in-go
func FileCopy(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

// DirCopy copies a whole directory recursively
// From here https://blog.depado.eu/post/copy-files-and-directories-in-go
func DirCopy(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = DirCopy(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = FileCopy(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}
