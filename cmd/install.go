package cmd

import (
	"fmt"
	"io"
	"os"
	"path"

	systemd "github.com/blunghamer/devproxy/pkg/systemd"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install devproxy",
	RunE:  runInstall,
}

const workingDirectoryPermission = 0755
const userdevproxydir = ".devproxy"
const userunitdir = ".config/systemd/user"

func runInstall(_ *cobra.Command, _ []string) error {

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	if err := ensureWorkingDir(path.Join(home, userdevproxydir)); err != nil {
		return err
	}

	if err := ensureWorkingDir(path.Join(home, userunitdir)); err != nil {
		return err
	}

	if err := cp("devproxy.yaml", path.Join(home, userdevproxydir)); err != nil {
		return err
	}

	if err := cp("devproxy", path.Join("/usr/local/bin/", "devproxy")); err != nil {
		return err
	}

	if err := binExists("/usr/local/bin/", "devproxy"); err != nil {
		return err
	}

	err = systemd.InstallUnit("devproxy", map[string]string{"Cwd": path.Join(home, userunitdir)})
	if err != nil {
		return err
	}

	err = systemd.DaemonReload()
	if err != nil {
		return err
	}

	err = systemd.Enable("devproxy")
	if err != nil {
		return err
	}

	err = systemd.Start("devproxy")
	if err != nil {
		return err
	}

	fmt.Println(`Check status with:
  sudo systemctl --user status devproxy
  sudo journalctl --user --user-unit -u devproxy --lines 100 -f`)

	return nil
}

func binExists(folder, name string) error {
	findPath := path.Join(folder, name)
	if _, err := os.Stat(findPath); err != nil {
		return fmt.Errorf("unable to stat %s, install this binary before continuing", findPath)
	}
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

func cp(source, destFolder string) error {
	file, err := os.Open(source)
	if err != nil {
		return err

	}
	defer file.Close()

	out, err := os.Create(path.Join(destFolder, source))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, file)

	return err
}
