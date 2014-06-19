package eventlog

import (
	"sync"
)

type Stage struct {
	log Log

	name string

	total         int
	currTaskIndex int

	indexLock sync.Mutex
}

func (s *Stage) BeginTask(name string) Task {
	s.indexLock.Lock()

	index := s.currTaskIndex
	s.currTaskIndex++

	s.indexLock.Unlock()

	task := Task{
		log: s.log,

		stageName: s.name,
		name:      name,

		total: s.total,
		index: index,
	}

	task.Start()

	return task
}
