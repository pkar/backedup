package backedup

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

func Test_initConfig(t *testing.T) {
	var stdin bytes.Buffer
	var writer bytes.Buffer
	fs := afero.NewMemMapFs()
	confPath := "/tmp/.backedup.yaml"
	exists, err := afero.Exists(fs, confPath)
	Ok(t, err)
	Equals(t, false, exists)

	stdin.Write([]byte("Yes\n"))
	initConfig(fs, &stdin, &writer, confPath)
	msg := fmt.Sprintf("%s does not exist, create a default one? <Yes|No>: creating %s\n", confPath, confPath)
	Equals(t, msg, writer.String())
	exists, err = afero.Exists(fs, confPath)
	Ok(t, err)
	Equals(t, true, exists)
}

func Test_initConfigExists(t *testing.T) {
	var stdin bytes.Buffer
	var writer bytes.Buffer
	fs := afero.NewMemMapFs()
	confPath := "/tmp/.backedup.yaml"
	err := afero.WriteFile(fs, confPath, []byte(DefaultCfg), 0644)
	Ok(t, err)
	err = initConfig(fs, &stdin, &writer, confPath)
	Equals(t, nil, err)
}

func Test_initConfigBackupExists(t *testing.T) {
	// the backup file exists so ask to copy that for use.
	var stdin bytes.Buffer
	var writer bytes.Buffer
	fs := afero.NewMemMapFs()
	confPath := "/tmp/.backedup.yaml"
	backedupPath := filepath.Join(DefaultBackupTo, DefaultBackedupFilename)
	err := afero.WriteFile(fs, backedupPath, []byte(DefaultCfg), 0644)
	Ok(t, err)

	stdin.Write([]byte("Yes\n"))
	initConfig(fs, &stdin, &writer, confPath)
	msg := fmt.Sprintf("%s exists, copy that to ~? <Yes|No>: copying %s config to %s\n", backedupPath, backedupPath, confPath)
	Equals(t, msg, writer.String())
	exists, err := afero.Exists(fs, confPath)
	Ok(t, err)
	Equals(t, true, exists)
}

func Test_initConfigNoCreate(t *testing.T) {
	var stdin bytes.Buffer
	var writer bytes.Buffer
	fs := afero.NewMemMapFs()
	confPath := "/tmp/.backedup.yaml"
	stdin.Write([]byte("No\n"))
	err := initConfig(fs, &stdin, &writer, confPath)
	Equals(t, err, ErrDefaultConfigNotCreated)
}
