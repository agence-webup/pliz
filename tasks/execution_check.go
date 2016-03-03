package tasks

import (
	"fmt"
	"os"
	"time"
)

type TaskExecutionCheck interface {
	CanExecute() bool
	PostExecute()
}

type ModificationDateTaskExecutionCheck struct {
	UpdatedFile string
	CompareTo   string
}

// Check if modification time of the file 'UpdatedFile' is newer than the file 'CompareTo'
func (chk ModificationDateTaskExecutionCheck) CanExecute() bool {
	updatedFileStat, err := os.Stat(chk.UpdatedFile)
	if err != nil {
		fmt.Println(err)
		return false
	}
	compareToStat, err := os.Stat(chk.CompareTo)
	if err != nil {
		// if the 'compareTo' file doesn't exist, then we consider we have to execute the task
		if os.IsNotExist(err) {
			return true
		}

		fmt.Println(err)
		return false
	}

	if updatedFileStat.ModTime().Before(compareToStat.ModTime()) {
		return false
	}

	return true
}

func (chk ModificationDateTaskExecutionCheck) PostExecute() {
	currentTime := time.Now().Local()
	os.Chtimes(chk.CompareTo, currentTime, currentTime)
}
