package session_cleaner

import sessionServ "core/internal/service/session"

type Process struct {
	sessionService sessionServ.SessionService
}

func NewProcess(sessionService sessionServ.SessionService) *Process {
	return &Process{
		sessionService: sessionService,
	}
}

func (p *Process) Start() {

}
