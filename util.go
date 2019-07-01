package backedup

import (
	"bufio"
	"fmt"
	"io"
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
