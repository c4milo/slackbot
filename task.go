package slackbot

type Task interface {
	// Init initializes the task's module
	Init() error
	// Validate returns an error if the task's module is declared with invalid or missing parameters
	Validate() error
	// Apply applies the state declared through a given module
	Apply() ([]byte, error)
	// Notify returns the name of a task being notified. The SlackBot runner will
	// delay running the notified task until the very end of the SlackBook definition, and in the same order
	// they were declared or queued up by the user. Notifications are only run if the notifier task's state changed.
	Notify() string
	// Changed tells whether or not the state of the module changed
	Changed() bool
	// IsHandler returns whether the task is a handler task. Handler tasks can only run through notifications
	IsHandler() bool
}

type task struct {
	Vars       map[string][]string `yaml:"vars"`
	NotifyTask string              `yaml:"notify"`
	Name       string              `yaml:"name"`
	State      string              `yaml:"state"`
	Handler    bool                `yml:"handler"`
	changed    bool
}

func (t task) Notify() string {
	return t.NotifyTask
}

func (t task) Changed() bool {
	return t.changed
}

func (t task) IsHandler() bool {
	return t.Handler
}
