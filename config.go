package backedup

import "path/filepath"

var (
	// DefaultBackupTo is the path to the directory for backing up files.
	DefaultBackupTo = "$HOME/Dropbox/backedup"
	// DefaultBackedupFilename is the default name of the backedup config file.
	DefaultBackedupFilename = ".backedup.yaml"
	// DefaultCfgPath is the default path to where the backedup config lives.
	DefaultCfgPath = filepath.Join("$HOME/", DefaultBackedupFilename)
)

const (
	// DefaultCfg is the default settings for an initial backup.
	DefaultCfg = `
backup_to: $HOME/Dropbox/backedup
paths:
  - $HOME/.ackrc
  - $HOME/.aws
  - $HOME/.backedup.yaml
  - $HOME/.bash_history
  - $HOME/.bash_profile
  - $HOME/.bashrc
  - $HOME/.curlrc
  - $HOME/.dlv
  - $HOME/.dockercfg
  - $HOME/.docker
  - $HOME/.gitignore
  - $HOME/.gitconfig
  - $HOME/.oh-my-zsh
  - $HOME/.m2/settings.xml
  - $HOME/.m2/toolchains.xml
  - $HOME/.netrc
  - $HOME/.profile
  - $HOME/.ron
  - $HOME/.ssh
  - $HOME/.tmux.conf
  - $HOME/.vim
  - $HOME/.vim-go
  - $HOME/.vimrc
  - $HOME/.wget-hsts
  - $HOME/.z
  - $HOME/.zshrc
  - $HOME/.zprofile
  - $HOME/.zlogin
  - $HOME/.zlogout
  - /Library/Application Support/Sequel Pro/Data
  - /Library/Preferences/com.sequelpro.SequelPro.plist
  - /Library/Preferences/com.googlecode.iterm2.plist
  - /Library/Preferences/com.apple.dt.Xcode.plist
  - /Library/Application Support/Developer/Shared/Xcode/Plug-ins
  - /Library/Developer/Xcode/UserData/CodeSnippets
  - /Library/Developer/Xcode/UserData/FontAndColorThemes
  - /Library/Developer/Xcode/UserData/KeyBindings
  - /Library/Developer/Xcode/UserData/Debugger
  - /Library/Developer/Xcode/UserData/xcdebugger
  - /Library/Developer/Xcode/UserData/SearchScopes.xcsclist
  - /Library/Developer/Xcode/Templates
`
)

// Config holds the main config file for backedup
type Config struct {
	// BackupTo is the folder to move files to that are then symlinked.
	BackupTo string `json:"backup_to" yaml:"backup_to"`
	// Paths are the file or directory paths to symlink
	Paths []string `json:"paths" yaml:"paths"`
}
