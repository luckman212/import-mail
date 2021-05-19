package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/gonejack/import-mail/cmd"
)

var (
	host        string
	port        int
	username    string
	password    string
	remotedir   string
	remoteLimit string

	prog = &cobra.Command{
		Use:   "import-mail *.eml",
		Short: "Command line tool for importing .eml files to IMAP account.",
		Run: func(c *cobra.Command, args []string) {
			err := run(c, args)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func init() {
	log.SetOutput(os.Stdout)

	prog.Flags().SortFlags = false
	prog.PersistentFlags().SortFlags = false

	flags := prog.PersistentFlags()
	{
		flags.StringVarP(&host, "host", "", "", "host")
		flags.IntVarP(&port, "port", "", 993, "port")
		flags.StringVarP(&username, "username", "", "", "username")
		flags.StringVarP(&password, "password", "", "", "password")
		flags.StringVarP(&remotedir, "remote-dir", "", "INBOX", "remote directory")
		flags.StringVarP(&remoteLimit, "remote-limit", "", "20M", "will not import email exceed this size")
	}
}

func run(c *cobra.Command, args []string) error {
	switch "" {
	case host:
		return fmt.Errorf("argument --host is required")
	case username:
		return fmt.Errorf("argument --username is required")
	case password:
		return fmt.Errorf("argument --password is required")
	}

	if len(args) == 0 {
		args, _ = filepath.Glob("*.eml")
	}

	exec := cmd.ImportMail{
		Host:           host,
		Port:           port,
		Username:       username,
		Password:       password,
		RemoteDir:      remotedir,
		RemoteLimit:    remoteLimit,
		SaveImportedTo: "imported",
	}

	return exec.Execute(args)
}

func main() {
	_ = prog.Execute()
}
