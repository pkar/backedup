package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pkar/backedup"
	"github.com/spf13/afero"
)

func main() {
	backedupCfgPath := flag.String("config", backedup.DefaultCfgPath, "The the path to the backedup config file")
	backup := flag.Bool("backup", false, "create symlinks for files configured to the backup path.")
	restore := flag.Bool("restore", false, "create symlinks for previously backed up files.")
	uninstall := flag.Bool("uninstall", false, "restore the configured files found in the config file to the originals without symlinks.")
	flag.Parse()

	b, err := backedup.New(afero.NewOsFs(), os.Stdin, os.Stderr, *backedupCfgPath)
	if err != nil {
		fmt.Println("ERRO:", err)
		os.Exit(1)
	}
	switch {
	case *uninstall:
		if err := b.Uninstall(); err != nil {
			fmt.Println("ERRO:", err)
			os.Exit(1)
		}
	case *backup:
		if err := b.Backup(); err != nil {
			fmt.Println("ERRO:", err)
			os.Exit(1)
		}
	case *restore:
		if err := b.Restore(); err != nil {
			fmt.Println("ERRO:", err)
			os.Exit(1)
		}
	default:
		flag.Usage()
		os.Exit(1)
	}
	fmt.Println("done")
}
