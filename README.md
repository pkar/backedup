# backedup

WIP: at this point use this in a non important container.

Backs up config files. Because a bash script would be boring,
mackup is in python, so why not use Go.
The default path is for Dropbox.

Create a file ~/.backedup.yaml with the following format.

```
backup_to: $HOME/Dropbox/backedup
paths:
	- $HOME/.ackrc
	- $HOME/.aws
	- $HOME/.backedup.yaml
```

Backup your defined configurations. The files defined in paths will be copied
to the field backup_to path and symlinks will be created if the files exist.

```
# first time use
backedup -backup

# use after -backup
backedup -restore

# give up
backedup -uninstall
```

Restore a previously configured setup. This will look in ~/Dropbox/backedup for
a .backedup.yaml, and if one isn't there it will create a default one ~/.backedup.yaml.

```
backedup restore
```
