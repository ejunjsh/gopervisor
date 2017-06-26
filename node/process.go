package node

import "time"

type process struct {
	pid string
	name string
	duration time.Duration
	launchTime time.Time
}

func (p *process) Start(){

}

func (p *process) Stop(){

}

func (p *process) Status(){

}