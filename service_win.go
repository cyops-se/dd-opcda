// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

var elog debug.Log

type myservice struct{}

func handlePanic() {
	if r := recover(); r != nil {
		elog.Error(1, fmt.Sprintf("recover: %#v", r))
		return
	}
}

func reportError(f string, args ...interface{}) {
	msg := fmt.Sprintf(f, args)
	log.Println(msg)
	if elog != nil {
		elog.Error(1, msg)
	}
}

func reportInfo(f string, args ...interface{}) {
	msg := fmt.Sprintf(f, args)
	log.Println(msg)
	if elog != nil {
		elog.Info(1, msg)
	}
}

func (m *myservice) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	defer handlePanic()

	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	changes <- svc.Status{State: svc.StartPending}
	fasttick := time.Tick(500 * time.Millisecond)
	slowtick := time.Tick(2 * time.Second)
	tick := fasttick
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	os.MkdirAll("/cyops/logs", 0755)
	if !ctx.trace {
		reportInfo("Logs are now redirected to '/cyops/logs/dd-opcda.*'")
		if stdout, err := os.OpenFile("/cyops/logs/dd-opcda.out.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND|os.O_SYNC, 0755); err != nil {
			reportError("Failed to open '/cyops/logs/dd-opcda.out.log', error; %s", err.Error())
		} else {
			os.Stdout = stdout
		}
		if stderr, err := os.OpenFile("/cyops/logs/dd-opcda.err.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND|os.O_SYNC, 0755); err != nil {
			reportError("Failed to open '/cyops/logs/dd-opcda.err.log', error; %s", err.Error())
		} else {
			os.Stderr = stderr
		}
	}

	reportInfo("starting engine")
	go runEngine()

	reportInfo("entering service control loop")

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
				// golang.org/x/sys/windows/svc.TestExample is verifying this output.
				testOutput := strings.Join(args, "-")
				testOutput += fmt.Sprintf("-%d", c.Context)
				reportInfo(testOutput)
				break loop
			case svc.Pause:
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
				tick = slowtick
			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
				tick = fasttick
			default:
				reportError("unexpected control request #%d", c)
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}

func runService(name string, isDebug bool) {
	var err error
	if isDebug {
		elog = debug.New(name)
	} else {
		elog, err = eventlog.Open(name)
		if err != nil {
			return
		}
	}
	defer elog.Close()

	reportInfo("starting %s service", name)
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	err = run(name, &myservice{})
	if err != nil {
		reportError("%s service failed: %v", name, err)
		return
	}
	reportInfo("%s service stopped", name)
}
