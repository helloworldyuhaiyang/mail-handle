package internal

type Scheduler struct {
}

func NewScheduler() *Scheduler {
	return &Scheduler{}
}

func (s *Scheduler) Start() error {
	return nil
}

func (s *Scheduler) Stop() error {
	return nil
}
