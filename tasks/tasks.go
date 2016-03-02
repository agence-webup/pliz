package tasks

type Task struct {
	Name           string
	Description    string
	ExecutionCheck *TaskExecutionCheck
}

type TaskExecutionCheck struct {
	UpdatedFile string
	CompareTo   string
}
