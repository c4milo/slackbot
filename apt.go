package slackbot

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"regexp"
)

// Apt implements a module for apt-get, found in debian-based systems.
// Available states are: present or absent
type Apt struct {
	task          `yaml:",inline"`
	path          string
	dpkgqueryPath string
	// Package is the name of the package to install or remove
	Package string `yaml:"package"`
	// UpdateIndex determines whether to update the local package index cache
	UpdateIndex bool `yaml:"update_index"`
}

func (a *Apt) Init() error {
	dpkgqueryPath, err := exec.LookPath("dpkg-query")
	if err != nil {
		return fmt.Errorf("apt: %s", err)
	}
	a.dpkgqueryPath = dpkgqueryPath

	path, err := exec.LookPath("apt-get")
	if err != nil {
		return fmt.Errorf("apt: %s", err)
	}

	a.path = path
	if a.State == "" {
		a.State = "present"
	}
	return nil
}

func (a *Apt) Validate() error {
	if a.Package == "" && !a.UpdateIndex {
		return errors.New("apt: at least one package name is required")
	}

	return nil
}

func (a *Apt) Apply() ([]byte, error) {
	installed, err := a.state()
	if err != nil {
		return nil, err
	}

	action := "install"
	if a.State == "absent" {
		action = "remove"
	}

	if action == "install" && installed ||
		action == "remove" && !installed {
		return nil, nil
	}

	if a.UpdateIndex {
		output, err := a.updateIndex()
		if err != nil || a.Package == "" {
			return output, err
		}

		if len(output) > 0 {
			log.Println(string(output[:]))
		}
	}

	cmd := exec.Command(a.path, "-y", action, a.Package)
	a.changed = true
	return cmd.CombinedOutput()
}

// state determines whether the package is installed or not
func (a *Apt) state() (bool, error) {
	cmd := exec.Command(a.dpkgqueryPath, "--status", a.Package)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, nil
	}

	regex := regexp.MustCompile(`Status: deinstall.+`)
	if regex.Match(output) {
		return false, nil
	}
	return true, nil
}

func (a *Apt) updateIndex() ([]byte, error) {
	fmt.Printf("Updating package index ...\n")
	cmd := exec.Command(a.path, "update")
	return cmd.CombinedOutput()
}
