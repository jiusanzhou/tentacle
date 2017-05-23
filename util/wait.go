package util

import "sync"

type OnceDone struct {
	wait *sync.WaitGroup
	once *sync.Once
}

func NewOnceDone() *OnceDone {
	return &OnceDone{&sync.WaitGroup{}, &sync.Once{}}
}

func (od *OnceDone) Add(n int) {
	od.wait.Add(n)
}

func (od *OnceDone) Done() {
	od.once.Do(od.wait.Done)
}

func (od *OnceDone) Wait() {
	od.wait.Wait()
}
