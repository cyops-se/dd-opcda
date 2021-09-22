// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package main

import (
	"log"
	"os"
	runtimedebug "runtime/debug"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
)

type myservice struct{}

// Windows services can't print to console, so we have to explicitly handle any panics.
func HandleAnyError() {
	if r := recover(); r != nil {
		log.Printf("Panic: %v\n%s\n", r, string(runtimedebug.Stack()))
	}
}

func (m *myservice) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	defer HandleAnyError()
	if ctx.cmd == "debug" {
		log.Println("Logs are now redirected to '/cyops/logs/dd-opcda.*'")
		os.Stdout, _ = os.OpenFile("/cyops/logs/dd-opcda.out.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND|os.O_SYNC, 0755)
		os.Stderr, _ = os.OpenFile("/cyops/logs/dd-opcda.err.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND|os.O_SYNC, 0755)
	}

	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	changes <- svc.Status{State: svc.StartPending}
	fasttick := time.Tick(500 * time.Millisecond)
	slowtick := time.Tick(2 * time.Second)
	tick := fasttick
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	go runEngine()

	log.Println("entering loop")

loop:
	for {
		select {
		case <-tick:
			// beep()
			// elog.Info(1, "beep")
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				// Testing deadlock from https://code.google.com/p/winsvc/issues/detail?id=4
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				changes <- svc.Status{State: svc.StopPending}
				log.Println("Shutting down service")
				break loop
			case svc.Pause:
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
				tick = slowtick
			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
				tick = fasttick
			default:
				log.Printf("unexpected control request #%d\n", c)
			}
		}
	}
	log.Println("Exiting service")
	return
}

func runService(name string, isDebug bool) {
	var err error
	log.Printf("starting %s service\n", name)
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	err = run(name, &myservice{})
	if err != nil {
		log.Printf("%s service failed: %v\n", name, err)
		return
	}
	log.Printf("%s service stopped\n", name)
}
