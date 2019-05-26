package slackbot

import (
	"bytes"
	"fmt"
	"os/exec"
)

// Service implements a module for the service utility found in RedHat/Fedora
// systems as well as in Ubuntu versions before 15.04
// Available states are: started, restarted, reloaded and stopped
type Service struct {
	task   `yaml:",inline"`
	path   string
	states map[string]string
}

// Init initializes the service module
func (s *Service) Init() error {
	path, err := exec.LookPath("service")
	if err != nil {
		return fmt.Errorf("service: %s", err)
	}
	s.path = path
	s.states = map[string]string{
		"restarted": "restart",
		"started":   "start",
		"reloaded":  "reload",
		"stopped":   "stop",
	}

	if s.State == "" {
		s.State = "started"
	}
	return nil
}

// Validate does nothing in this module
func (s *Service) Validate() error {
	return nil
}

// Apply applies the declared state for a given service
func (s *Service) Apply() ([]byte, error) {
	running, err := s.state()
	if err != nil {
		return nil, err
	}

	action, ok := s.states[s.State]
	if !ok {
		return nil, fmt.Errorf("service: invalid state: %s", s.State)
	}

	if running && action == "start" ||
		!running && action == "stop" {
		return nil, nil
	}

	if !running && (action == "reload" || action == "restarted") {
		action = "start"
	}

	cmd := exec.Command(s.path, s.Name, action)
	return cmd.CombinedOutput()
}

// state returns the current state of the service
func (s *Service) state() (bool, error) {
	cmd := exec.Command(s.path, s.Name, "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, nil
	}

	if bytes.ContainsAny(output, "start/running") ||
		bytes.ContainsAny(output, "is running") {
		return true, nil
	}

	return false, nil
}
