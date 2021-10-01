package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/blunghamer/devproxy"
	"github.com/blunghamer/systemd"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(installSudoCmd)
}

var installSudoCmd = &cobra.Command{
	Use:   "installsudo",
	Short: "move devproxy binary to bin folder, run as sudo please",
	RunE:  runInstallSudo,
}

func runInstallSudo(_ *cobra.Command, _ []string) error {

	targetFolder := "/usr/local/bin/"
	binaryName := "devproxy"

	unitTargetFolder := "/usr/lib/systemd/user/"
	serviceFile := "devproxy.service"

	exeFile := os.Args[0]
	outfile := path.Join(targetFolder, binaryName)
	serviceOut := path.Join(unitTargetFolder, serviceFile)

	if err := cp(exeFile, outfile); err != nil {
		log.Printf("Unable to copy %v from %v to %v: %v", binaryName, exeFile, outfile, err)
		return err
	}

	if err := os.Chmod(outfile, 0755); err != nil {
		log.Printf("Unable to chmod binary %v", err)
		return err
	}

	if err := binExists(targetFolder, binaryName); err != nil {
		return err
	}

	log.Printf("Successfully installed binary %v to %v", binaryName, targetFolder)

	src, err := devproxy.FS.Open(filepath.Join("static", serviceFile))
	if err != nil {
		return err
	}
	defer src.Close()

	if err := cpf(src, serviceOut); err != nil {
		log.Printf("Unable to copy %v from %v to %v: %v", serviceFile, serviceFile, serviceOut, err)
		return err
	}

	log.Printf("Successfully installed service file %v to %v", serviceFile, serviceOut)

	if err := systemd.DaemonReload(); err != nil {
		log.Printf("Unable to reload systemd daemon%v", err)
		return err
	}

	return nil
}

func binExists(folder, name string) error {
	findPath := path.Join(folder, name)
	if _, err := os.Stat(findPath); err != nil {
		return fmt.Errorf("unable to stat %s, install this binary before continuing", findPath)
	}
	return nil
}

func cpf(source io.Reader, dest string) error {

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, source)

	return err
}

func cp(source, dest string) error {
	file, err := os.Open(source)
	if err != nil {
		return err

	}
	defer file.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, file)

	return err
}
