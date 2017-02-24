package tcp

import (
	"fmt"
	"net"
	"sync"
)

type TCPService struct {
	Bind     string
	Port     int
	InitConn func(net.Conn) (TCPConn, error)
	nl       net.Listener
	stopOnce sync.Once
}

type TCPConn interface {
	Serve() error
}

func (tcp *TCPService) Start() error {
	service := fmt.Sprintf("%s:%d", tcp.Bind, tcp.Port)
	l, err := net.Listen("tcp", service)
	if err != nil {
		return err
	}
	tcp.nl = l

	go tcp.serve()
	return nil
}

func (tcp *TCPService) Stop() {
	tcp.stopOnce.Do(func() {
		if tcp.nl != nil {
			tcp.nl.Close()
		}
	})
}

func (tcp *TCPService) serve() error {
	ch := make(chan net.Conn, 4096)
	defer close(ch)

	go func() {
		for netc := range ch {
			var (
				err error
				//d   Duplexer
			)

			c, err := tcp.InitConn(netc)
			if err != nil {
				continue
			}

			go func() {
				if err := c.Serve(); err != nil {
					//
				}
			}()
		}
	}()

	wg := &sync.WaitGroup{}
	wg.Add(10)

	doAccept := func(ch chan net.Conn) {
		for {
			c, err := tcp.nl.Accept()
			if err != nil {
				if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
					//log.Warn("[transport:tcp] Listen failed, skip", "error", nerr)
					continue
				}
				wg.Done()
				return
			}
			ch <- c
		}
	}

	for i := 0; i < 10; i++ {
		go doAccept(ch)
	}

	wg.Wait()
	return nil
}
