package slackbot

import (
	"bytes"
	"fmt"
	"os/exec"
)

type Service struct {
	task   `yaml:",inline"`
	path   string
	states map[string]string
}

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

func (s *Service) Validate() error {
	return nil
}

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
