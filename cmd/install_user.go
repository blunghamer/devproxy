package cmd

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/blunghamer/devproxy/pkg/systemd"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(installUserCmd)
}

var installUserCmd = &cobra.Command{
	Use:   "installuser",
	Short: "install user service of devproxy run as normal user",
	RunE:  runInstallUser,
}

const workingDirectoryPermission = 0755
const userdevproxydir = ".config/devproxy"

func runInstallUser(_ *cobra.Command, _ []string) error {

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	if err := ensureWorkingDir(path.Join(home, userdevproxydir)); err != nil {
		return err
	}

	absConfig := path.Join(home, userdevproxydir, "devproxy.yaml")
	if err := cp("devproxy.yaml", absConfig); err != nil {
		log.Fatalf("Unable to copy devproxy.yaml to %v: %v", absConfig, err)
		return err
	}

	err = systemd.Enable("devproxy", true)
	if err != nil {
		return err
	}

	err = systemd.Start("devproxy", true)
	if err != nil {
		return err
	}

	fmt.Println(`Check status with:
  systemctl --user status devproxy
  journalctl --user --user-unit devproxy --lines 100 -f`)

	return nil
}

func ensureWorkingDir(folder string) error {
	if _, err := os.Stat(folder); err != nil {
		err = os.MkdirAll(folder, workingDirectoryPermission)
		if err != nil {
			return err
		}
	}

	return nil
}
