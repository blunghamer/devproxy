package systemd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	execute "github.com/alexellis/go-execute/pkg/v1"
)

func Enable(unit string, user bool) error {
	us := ""
	if user {
		us = "--user"
	}
	task := execute.ExecTask{Command: "systemctl",
		Args:        []string{"enable", us, unit},
		StreamStdio: false,
	}

	res, err := task.Execute()
	if err != nil {
		return err
	}

	if res.ExitCode != 0 {
		return fmt.Errorf("error executing task %s %v, stderr: %s", task.Command, task.Args, res.Stderr)
	}

	return nil
}

func Start(unit string, user bool) error {
	us := ""
	if user {
		us = "--user"
	}
	task := execute.ExecTask{Command: "systemctl",
		Args:        []string{"start", us, unit},
		StreamStdio: false,
	}

	res, err := task.Execute()
	if err != nil {
		return err
	}

	if res.ExitCode != 0 {
		return fmt.Errorf("error executing task %s %v, stderr: %s", task.Command, task.Args, res.Stderr)
	}

	return nil
}

func DaemonReload() error {
	task := execute.ExecTask{Command: "systemctl",
		Args:        []string{"daemon-reload"},
		StreamStdio: false,
	}

	res, err := task.Execute()
	if err != nil {
		return err
	}

	if res.ExitCode != 0 {
		return fmt.Errorf("error executing task %s %v, stderr: %s", task.Command, task.Args, res.Stderr)
	}

	return nil
}

func InstallUnit(name string, tokens map[string]string, user bool) error {
	if len(tokens["Cwd"]) == 0 {
		return fmt.Errorf("key Cwd expected in tokens parameter")
	}

	tmplName := "devproxy.service"
	tmpl, err := template.ParseFiles(tmplName)

	if err != nil {
		return fmt.Errorf("error loading template %s, error %s", tmplName, err)
	}

	var tpl bytes.Buffer

	err = tmpl.Execute(&tpl, tokens)
	if err != nil {
		return err
	}

	err = writeUnit(name+".service", tpl.Bytes(), user)

	if err != nil {
		return err
	}
	return nil
}

func writeUnit(name string, data []byte, user bool) error {
	unitpath := "/lib/systemd/system"
	if user {
		unitpath = "/lib/systemd/user"
	}
	f, err := os.Create(filepath.Join(unitpath, name))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}
