package backedup

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

const (
	backupHomeDirName = "_HOME"
)

var (
	// ErrNoDefaultConfig if the default config is not found
	ErrNoDefaultConfig = errors.New("no default config")
	// ErrDefaultConfigNotCreated for errors creating the default config
	ErrDefaultConfigNotCreated = errors.New("config not created")
)

// Backedup will handle the backing up of files.
type Backedup struct {
	ConfPath string
	Config   *Config
	homeDir  string
	fs       afero.Fs
	logger   io.Writer
	stdin    io.Reader
}

// New will initialize a new Backedup configuration. If the input configuration file
// path is not found, it will prompt for creating a new default one.
func New(fs afero.Fs, stdin io.Reader, logger io.Writer, confPath string) (*Backedup, error) {
	// get the users home directory for expanding $HOME in config.
	if os.Getenv("HOME") == "" {
		homeDir, err := homedir.Dir()
		if err != nil {
			return nil, err
		}
		os.Setenv("HOME", filepath.Clean(homeDir))
	}
	// expand any $HOME environment variables.
	confPath = os.ExpandEnv(confPath)

	// check the conf path existance and prompt to create if not there.
	if err := initConfig(fs, stdin, logger, confPath); err != nil {
		return nil, err
	}
	confData, err := afero.ReadFile(fs, confPath)
	if err != nil {
		return nil, err
	}
	conf := &Config{}
	if err = yaml.Unmarshal(confData, &conf); err != nil {
		return nil, err
	}
	conf.BackupTo = os.ExpandEnv(conf.BackupTo)
	for i, path := range conf.Paths {
		conf.Paths[i] = os.ExpandEnv(path)
	}

	b := &Backedup{
		Config:   conf,
		ConfPath: confPath,
		homeDir:  os.Getenv("HOME"),
		fs:       fs,
		logger:   logger,
		stdin:    stdin,
	}
	return b, nil
}

// Backup moves files configured and creates symlinks to the backed up directory.
// This should be done once on an initial run. NOTE Backup only works for the
// case that fs is *afero.OsFs because due to lack of symlink support in afero.
func (b *Backedup) Backup() error {
	// create backup directory if it doesn't exist
	if err := b.fs.MkdirAll(b.Config.BackupTo, 0755); err != nil {
		return err
	}
	backupHomeDir := filepath.Join(b.Config.BackupTo, backupHomeDirName)
	if err := b.fs.MkdirAll(backupHomeDir, 0755); err != nil {
		return err
	}
	// iterate config list of paths, checking for symlinks first, if not there
	// move the current file and create a new symlink.
PATH_LOOP:
	for _, path := range b.Config.Paths {
		var fi os.FileInfo
		var err error
		if fsOs, ok := b.fs.(*afero.OsFs); ok {
			fi, _, err = fsOs.LstatIfPossible(path)
		} else {
			fi, err = os.Lstat(path)
		}
		if err != nil {
			b.logger.Write([]byte(fmt.Sprintf("ERRO: %s %s\n", path, err)))
			continue PATH_LOOP
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			// skip symlink files
			b.logger.Write([]byte(fmt.Sprintf("ERRO: %s symlink exists\n", path)))
			continue PATH_LOOP
		}
		// change home paths to $HOME/path
		isHomeDir := false
		if strings.HasPrefix(path, b.homeDir) {
			isHomeDir = true
		}
		if fi.IsDir() {
			// handle directories
			dirBackupName := filepath.Base(path)
			dirBackupDir := filepath.Join(b.Config.BackupTo, filepath.Dir(path))
			if isHomeDir {
				dirBackupDir = filepath.Join(backupHomeDir, strings.Replace(filepath.Dir(path), b.homeDir, "", -1))
			}
			if err := b.fs.MkdirAll(dirBackupDir, 0755); err != nil {
				b.logger.Write([]byte(fmt.Sprintf("ERRO: %s %s\n", path, err)))
				continue PATH_LOOP
			}
			backupPath := filepath.Join(dirBackupDir, dirBackupName)
			if err := b.fs.Rename(path, backupPath); err != nil {
				b.logger.Write([]byte(fmt.Sprintf("ERRO: %s %s\n", path, err)))
				continue PATH_LOOP
			}
			if err := os.Symlink(backupPath, path); err != nil {
				b.logger.Write([]byte(fmt.Sprintf("ERRO: %s %s\n", path, err)))
			}
			continue PATH_LOOP
		}
		// handle files
		fileBackupName := filepath.Base(path)
		fileBackupDir := backupHomeDir
		if !isHomeDir {
			fileBackupDir = filepath.Join(b.Config.BackupTo, filepath.Dir(path))
		}
		if err := b.fs.MkdirAll(fileBackupDir, 0755); err != nil {
			b.logger.Write([]byte(fmt.Sprintf("ERRO: %s %s\n", path, err)))
			continue PATH_LOOP
		}
		backupPath := filepath.Join(fileBackupDir, fileBackupName)
		if err := b.fs.Rename(path, backupPath); err != nil {
			b.logger.Write([]byte(fmt.Sprintf("ERRO: %s %s\n", path, err)))
			continue PATH_LOOP
		}
		if err := os.Symlink(backupPath, path); err != nil {
			b.logger.Write([]byte(fmt.Sprintf("ERRO: %s %s\n", path, err)))
		}
	}

	return nil
}

// Restore creates symlinks for previously backed up files.
func (b *Backedup) Restore() error {
	exists, err := afero.Exists(b.fs, b.Config.BackupTo)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%s does not exist", b.Config.BackupTo)
	}
	backupHomeDir := filepath.Join(b.Config.BackupTo, backupHomeDirName)
	if err := b.fs.MkdirAll(backupHomeDir, 0755); err != nil {
		return err
	}
PATH_LOOP:
	for _, path := range b.Config.Paths {
		exists, err := afero.Exists(b.fs, b.Config.BackupTo)
		if err != nil {
			b.logger.Write([]byte(fmt.Sprintf("ERRO: %s %s\n", path, err)))
			continue PATH_LOOP
		}
		if !exists {
			b.logger.Write([]byte(fmt.Sprintf("ERRO: backup path doesn't exist %s\n", path)))
			continue PATH_LOOP
		}

		isHomeDir := false
		if strings.HasPrefix(path, b.homeDir) {
			isHomeDir = true
		}

		backupPath := ""
		if isHomeDir {
			backupPath = filepath.Join(backupHomeDir, strings.Replace(filepath.Dir(path), b.homeDir, "", -1))
		} else {
			backupPath = filepath.Join(b.Config.BackupTo, path)
		}

		if err := b.fs.MkdirAll(filepath.Dir(path), 0755); err != nil {
			b.logger.Write([]byte(fmt.Sprintf("ERRO: %s %s\n", path, err)))
			continue PATH_LOOP
		}

		fmt.Println("creating symlink", backupPath, path)
		// create the symlink from the backup path to path
		if err := os.Symlink(backupPath, path); err != nil {
			b.logger.Write([]byte(fmt.Sprintf("ERRO: %s %s\n", path, err)))
			continue PATH_LOOP
		}
	}

	return nil
}

// Uninstall removes symlinks to backed up files and restores
// the original files.
func (b *Backedup) Uninstall() error {
	exists, err := afero.Exists(b.fs, b.Config.BackupTo)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%s does not exist", b.Config.BackupTo)
	}
	backupHomeDir := filepath.Join(b.Config.BackupTo, backupHomeDirName)
	if err := b.fs.MkdirAll(backupHomeDir, 0755); err != nil {
		return err
	}

PATH_LOOP:
	for _, path := range b.Config.Paths {
		// check that the path is a symlink, remove it and move the backed up files
		// to path
		var fi os.FileInfo
		var err error
		if fsOs, ok := b.fs.(*afero.OsFs); ok {
			fi, _, err = fsOs.LstatIfPossible(path)
		} else {
			fi, err = os.Lstat(path)
		}
		if err != nil {
			b.logger.Write([]byte(fmt.Sprintf("ERRO: %s %s\n", path, err)))
			continue PATH_LOOP
		}
		if fi.Mode()&os.ModeSymlink == 0 {
			// skip non symlink files
			b.logger.Write([]byte(fmt.Sprintf("ERRO: %s file exists, not symlink\n", path)))
			continue PATH_LOOP
		}
		// remove the current path symlink
		if err := os.Remove(path); err != nil {
			b.logger.Write([]byte(fmt.Sprintf("ERRO: %s %s\n", path, err)))
			continue PATH_LOOP
		}

		// move backed up data to path
		isHomeDir := false
		if strings.HasPrefix(path, b.homeDir) {
			isHomeDir = true
		}
		backupPath := ""
		if isHomeDir {
			backupPath = filepath.Join(backupHomeDir, strings.Replace(path, b.homeDir, "", -1))
		} else {
			backupPath = filepath.Join(b.Config.BackupTo, path)
		}

		if fsOs, ok := b.fs.(*afero.OsFs); ok {
			fi, _, err = fsOs.LstatIfPossible(backupPath)
		} else {
			fi, err = os.Lstat(backupPath)
		}
		if err != nil {
			b.logger.Write([]byte(fmt.Sprintf("ERRO: %s %s\n", path, err)))
			continue PATH_LOOP
		}

		if err := b.fs.MkdirAll(filepath.Dir(path), 0755); err != nil {
			b.logger.Write([]byte(fmt.Sprintf("ERRO: %s %s\n", path, err)))
			continue PATH_LOOP
		}

		if fi.IsDir() {
			if err := DirCopy(backupPath, path); err != nil {
				b.logger.Write([]byte(fmt.Sprintf("ERRO: %s %s\n", path, err)))
			}
		} else {
			if err := FileCopy(backupPath, path); err != nil {
				b.logger.Write([]byte(fmt.Sprintf("ERRO: %s %s\n", path, err)))
			}
		}
	}

	return nil
}
