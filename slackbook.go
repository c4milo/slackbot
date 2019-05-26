package slackbot

import (
	"fmt"
	"log"
	"os"

	"github.com/c4milo/gotoolkit"
	yaml "gopkg.in/yaml.v3"
)

// SlackBook defines a list of tasks to run
type SlackBook struct {
	path string

	// Tasks contains a set of all the tasks defined by the user.
	// A map is used to efficiently implement notifications.
	Tasks map[string]Task

	// tasks contains task names in the order they were defined by the user.
	tasks []string

	// notifyQueue holds notify elements declared by user, in order. Notifications are deduplicated and sent out at the end of
	// the SlackBook run.
	notifyQueue gotoolkit.Queue
}

// Decode decodes SlackBook YAML configuration file into the current SlackBook receiver instance.
func (sb *SlackBook) Decode(fpath string) error {
	sb.path = fpath

	file, err := os.Open(sb.path)
	if err != nil {
		return err
	}
	defer file.Close()

	sb.Tasks = make(map[string]Task)
	sb.notifyQueue = new(gotoolkit.ListQueue)

	decoder := yaml.NewDecoder(file)

	if err := decoder.Decode(sb); err != nil {
		return err
	}

	return nil
}

// UnmarshalYAML allows to decode tasks while keeping the method Run simple thanks to the Task interface
func (sb *SlackBook) UnmarshalYAML(value *yaml.Node) error {
	for _, task := range value.Content {
		var t Task

		if len(task.Content) < 3 {
			return fmt.Errorf("invalid task declaration found in: %s, line %d.", sb.path, task.Line)
		}

		module := task.Content[2]
		switch module.Value {
		case "apt":
			t = new(Apt)
		case "file":
			t = new(File)
		case "service":
			t = new(Service)
		default:
			return fmt.Errorf("not supported module %q found in: %s, line %d", module.Value, sb.path, module.Line)
		}

		// decode task metadata
		if err := task.Decode(t); err != nil {
			return err
		}

		// decode declared module state
		if err := task.Content[3].Decode(t); err != nil {
			return err
		}

		if err := t.Validate(); err != nil {
			return fmt.Errorf("%s: %s, line %d", err, sb.path, task.Line)
		}

		if err := t.Init(); err != nil {
			return fmt.Errorf("%s: %s, line %d", err, sb.path, task.Line)
		}

		taskName := task.Content[1].Value
		if _, ok := sb.Tasks[taskName]; !ok {
			sb.Tasks[taskName] = t
		} else {
			return fmt.Errorf("duplicated task name %q found in: %s, line %d", taskName, sb.path, task.Line)
		}

		sb.tasks = append(sb.tasks, taskName)
	}
	return nil
}

// Run runs SlackBook tasks in user defined order, skipping running notified
// tasks in the declared order and instead running them at the end of the SlackBook.
func (s *SlackBook) Run() error {
	for _, name := range s.tasks {
		t := s.Tasks[name]

		// Handler tasks can only be run through notifications
		if t.IsHandler() {
			continue
		}

		fmt.Printf("-----> %s... \n", name)
		out, err := t.Apply()
		if len(out) > 0 {
			log.Println(string(out[:]))
		}

		if err != nil {
			return err
		}

		if t.Changed() {
			if taskName := t.Notify(); taskName != "" {
				notifiedTask, ok := s.Tasks[taskName]
				if !ok || !notifiedTask.IsHandler() {
					return fmt.Errorf("notified task does not exist or is not a handler: %q", taskName)
				}
				s.notifyQueue.Enqueue(notifiedTask)
			}
		}
	}

	// Run notified tasks
	fmt.Printf("-----> Running %d notified tasks... \n", s.notifyQueue.Size())
	for !s.notifyQueue.IsEmpty() {
		t, _ := s.notifyQueue.Dequeue()
		out, err := t.(Task).Apply()
		if len(out) > 0 {
			log.Println(string(out[:]))
		}

		if err != nil {
			return err
		}
	}
	return nil
}
