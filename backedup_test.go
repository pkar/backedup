package backedup

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"testing"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

// Ok fails the test if an err is not nil.
func Ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// Equals fails the test if exp is not equal to act.
func Equals(tb testing.TB, exp, act interface{}, msg ...string) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		if len(msg) > 0 {
			fmt.Printf(
				"\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n%v\n\n",
				filepath.Base(file), line, exp, act, msg)
		} else {
			fmt.Printf(
				"\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n",
				filepath.Base(file), line, exp, act)
		}
		tb.FailNow()
	}
}

func printFs(fs afero.Fs, dir string) {
	afero.Walk(fs, dir, func(path string, info os.FileInfo, err error) error {
		fmt.Println(path)
		return nil
	})
}

func getFs(fs afero.Fs, dir string) []string {
	dirtree := []string{}
	afero.Walk(fs, dir, func(path string, info os.FileInfo, err error) error {
		dirtree = append(dirtree, path)
		return nil
	})
	sort.Strings(dirtree)
	return dirtree
}

func TestNew(t *testing.T) {
	var stdin bytes.Buffer
	var logger bytes.Buffer
	fs := afero.NewMemMapFs()
	confPath := "/tmp/.backedup.yaml"
	err := afero.WriteFile(fs, confPath, []byte(DefaultCfg), 0644)
	Ok(t, err)

	b, err := New(fs, &stdin, &logger, confPath)
	Ok(t, err)
	want := &Config{}
	yaml.Unmarshal([]byte(DefaultCfg), want)
	want.BackupTo = os.ExpandEnv(want.BackupTo)
	for i, path := range want.Paths {
		want.Paths[i] = os.ExpandEnv(path)
	}
	Equals(t, want, b.Config)
}

func checkNewSymlink(t *testing.T, fs *afero.OsFs, path string) {
	exists, err := afero.Exists(fs, path)
	Ok(t, err)
	Equals(t, true, exists, "path not found "+path)
	// symlink should be there pointing to moved file
	fi, _, err := fs.LstatIfPossible(path)
	Ok(t, err)
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("%s is not a symlink", path)
	}
}

func checkNotSymlink(t *testing.T, fs *afero.OsFs, path string) {
	exists, err := afero.Exists(fs, path)
	Ok(t, err)
	Equals(t, true, exists, "path not found "+path)
	// symlink should be there pointing to moved file
	fi, _, err := fs.LstatIfPossible(path)
	Ok(t, err)
	if fi.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("%s is a symlink", path)
	}
}
func TestBackupRestoreUninstall(t *testing.T) {
	var stdin bytes.Buffer
	// var logger bytes.Buffer
	logger := os.Stderr
	fs := afero.NewOsFs()
	tmpDir, err := afero.TempDir(fs, "", "")
	Ok(t, err)
	defer fs.RemoveAll(tmpDir)
	confPath := filepath.Join(tmpDir, "home", ".backedup.yaml")
	os.Setenv("HOME", filepath.Join(tmpDir, "home"))

	// create a directory, file and symlink for backing up.
	dir := filepath.Join(tmpDir, ".dir", "nested")
	err = fs.MkdirAll(dir, 0755)
	Ok(t, err)

	file1 := filepath.Join(tmpDir, ".file1")
	err = afero.WriteFile(fs, file1, []byte("file1 contents"), 0644)
	Ok(t, err)

	// create home directory and files
	err = fs.MkdirAll(filepath.Join(tmpDir, "home"), 0755)
	Ok(t, err)
	filehome := filepath.Join(tmpDir, "home/.filehome")
	err = afero.WriteFile(fs, filehome, []byte("file1 contents"), 0644)
	Ok(t, err)
	dirhome := filepath.Join(tmpDir, "home/.dirhome/somedir")
	err = fs.MkdirAll(dirhome, 0755)
	Ok(t, err)

	// create file in a directory
	err = fs.MkdirAll(filepath.Join(tmpDir, "somedir"), 0755)
	Ok(t, err)
	file2 := filepath.Join(tmpDir, "somedir/.file2")
	err = afero.WriteFile(fs, file2, []byte("file2 contents"), 0644)
	Ok(t, err)

	// create a symlink file that shouldn't be backed up
	symlink := filepath.Join(tmpDir, ".symlink")
	err = os.Symlink(tmpDir, symlink)
	Ok(t, err)
	defer os.Remove(symlink)

	backupTo := filepath.Join(tmpDir, "backedup")

	cfg := fmt.Sprintf(`
backup_to: %s
paths:
  - %s
  - %s
  - %s
  - %s
  - %s
  - %s
  - %s`, backupTo, dir, dirhome, filehome, file1, file2, symlink, confPath)

	err = afero.WriteFile(fs, confPath, []byte(cfg), 0644)
	Ok(t, err)

	b, err := New(fs, &stdin, logger, confPath)
	Ok(t, err)
	err = b.Backup()
	Ok(t, err)

	exists, err := afero.Exists(fs, backupTo)
	Ok(t, err)
	Equals(t, true, exists)

	exists, err = afero.Exists(fs, filepath.Join(backupTo, dir))
	Ok(t, err)
	Equals(t, true, exists)
	checkNewSymlink(t, fs.(*afero.OsFs), dir)

	exists, err = afero.Exists(fs, filepath.Join(backupTo, filepath.Join(backupHomeDirName, ".dirhome")))
	Ok(t, err)
	Equals(t, true, exists)
	checkNewSymlink(t, fs.(*afero.OsFs), dirhome)

	exists, err = afero.Exists(fs, filepath.Join(backupTo, filepath.Join(backupHomeDirName, ".filehome")))
	Ok(t, err)
	Equals(t, true, exists)
	checkNewSymlink(t, fs.(*afero.OsFs), filehome)

	exists, err = afero.Exists(fs, filepath.Join(backupTo, file1))
	Ok(t, err)
	Equals(t, true, exists)
	checkNewSymlink(t, fs.(*afero.OsFs), file1)

	exists, err = afero.Exists(fs, filepath.Join(backupTo, file2))
	Ok(t, err)
	Equals(t, true, exists)
	checkNewSymlink(t, fs.(*afero.OsFs), file2)

	// test restore by removing symlinks and running Restore
	err = os.Remove(dir)
	Ok(t, err)
	err = os.Remove(dirhome)
	Ok(t, err)
	err = os.Remove(filehome)
	Ok(t, err)
	err = os.Remove(file1)
	Ok(t, err)
	err = os.Remove(file2)
	Ok(t, err)

	err = b.Restore()
	Ok(t, err)

	checkNewSymlink(t, fs.(*afero.OsFs), dir)
	checkNewSymlink(t, fs.(*afero.OsFs), dirhome)
	checkNewSymlink(t, fs.(*afero.OsFs), filehome)
	checkNewSymlink(t, fs.(*afero.OsFs), file1)
	checkNewSymlink(t, fs.(*afero.OsFs), file2)
	checkNewSymlink(t, fs.(*afero.OsFs), confPath)

	err = b.Uninstall()
	Ok(t, err)

	checkNotSymlink(t, fs.(*afero.OsFs), dir)
	checkNotSymlink(t, fs.(*afero.OsFs), dirhome)
	checkNotSymlink(t, fs.(*afero.OsFs), filehome)
	checkNotSymlink(t, fs.(*afero.OsFs), file1)
	checkNotSymlink(t, fs.(*afero.OsFs), file2)
	checkNotSymlink(t, fs.(*afero.OsFs), confPath)
}
