# backedup

alpha: at this point use this in a non important environment.

Backs up config files. Because a bash script would be boring,
mackup is in python, so why not use Go.
The default path is for a Dropbox folder but could be any.

Create a file ~/.backedup.yaml with the following format. Use
$HOME for any home directory paths.

```
backup_to: $HOME/Dropbox/backedup
paths:
	- $HOME/.backedup.yaml
	- $HOME/.ackrc
	- $HOME/.aws
```

A default config will be optionally generated if not in $HOME the first
time around.

Backup your defined configurations. The files defined in paths will be copied
to the field backup_to path and symlinks will be created if the files exist.

```
cd cmd/backedup && go install

# first time use
backedup -backup

# Restore a previously configured setup. This will look in ~/Dropbox/backedup
# for a .backedup.yaml, and if one isn't there it will create a default one
# ~/.backedup.yaml. use only after -backup
backedup -restore

# give up
backedup -uninstall
```
