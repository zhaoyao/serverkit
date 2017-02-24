package serverkit

import (
	"os"
	"os/signal"
	"syscall"
)

type Server struct {
	ctx *Context
}

func NewServer(config interface{}) (*Server, error) {
	return &Server{ctx: NewContext()}, nil
}

func (s *Server) Register(svc Service) {
	s.ctx.Register(svc)
}

func (s *Server) Start() error {
	return s.ctx.Start()
}

func (s *Server) StopOnSignal() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	//log.Info("Server exiting", "signal", sig)
	s.Stop()
}

func (s *Server) Stop() error {
	//s.lifecycle.Stop()
	s.ctx.Stop()
	return nil
}
