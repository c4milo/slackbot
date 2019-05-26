package slackbot

// Task defines the interface to implement by all modules
type Task interface {
	// Init initializes the task's module
	Init() error
	// Validate checks for invalid or missing parameters during configuration decoding.
	Validate() error
	// Apply applies the state declared by the task
	Apply() ([]byte, error)
	// Notify returns the name of a task being notified. The SlackBook runner will
	// delay running the notified task until the very end of the SlackBook definition, and in the same order
	// they were declared by the user. Notifications are only run if the notifier task's state changed.
	Notify() string
	// Changed tells whether or not the module changed any state. It is used for notifications.
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

// Notify returns the name of the task to notify
func (t task) Notify() string {
	return t.NotifyTask
}

// Changed returns whether the task made state change or not
func (t task) Changed() bool {
	return t.changed
}

// IsHandler returns whether the task is a handler or not
func (t task) IsHandler() bool {
	return t.Handler
}
