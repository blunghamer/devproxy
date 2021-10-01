package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/blunghamer/devproxy"
	"github.com/blunghamer/systemd"

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

	fmt.Print("Please enter your proxy username: ")
	reader := bufio.NewReader(os.Stdin)
	username, _ := reader.ReadString('\n')
	username = strings.ReplaceAll(username, "\n", "")

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	if err := ensureWorkingDir(path.Join(home, userdevproxydir)); err != nil {
		return err
	}

	tmpl, err := template.ParseFS(devproxy.FS, "static/devproxy.yaml")
	if err != nil {
		return err
	}

	absConfig := path.Join(home, userdevproxydir, "devproxy.yaml")

	of, err := os.Create(absConfig)
	if err != nil {
		return err
	}
	defer of.Close()

	err = tmpl.Execute(of, map[string]interface{}{"ProxyUser": username})
	if err != nil {
		log.Printf("Unable to write config to %v: %v", absConfig, err)
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

	fmt.Printf(`Successfully installed, check status with:
  systemctl --user status devproxy
  journalctl --user --user-unit devproxy --lines 100 -f
  configuration in %v`, absConfig)

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
