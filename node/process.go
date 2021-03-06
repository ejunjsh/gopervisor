package node

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"sync"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ejunjsh/gopervisor/config"
)

type ProcessState int

const (
	STOPPED  ProcessState = iota
	STARTING              = 10
	RUNNING               = 20
	BACKOFF               = 30
	STOPPING              = 40
	EXITED                = 100
	FATAL                 = 200
	UNKNOWN               = 1000
)

func (p ProcessState) String() string {
	switch p {
	case STOPPED:
		return "STOPPED"
	case STARTING:
		return "STARTING"
	case RUNNING:
		return "RUNNING"
	case BACKOFF:
		return "BACKOFF"
	case STOPPING:
		return "STOPPING"
	case EXITED:
		return "EXITED"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

type Process struct {
	supervisor_id string
	config        *config.ProcConfig
	cmd           *exec.Cmd
	startTime     time.Time
	stopTime      time.Time
	state         ProcessState
	//true if process is starting
	inStart bool
	//true if the process is stopped by user
	stopByUser bool
	retryTimes int
	lock       sync.RWMutex
	stdin      io.WriteCloser
	stdoutLog  Logger
	stderrLog  Logger
	exitc      chan struct{}
}

func NewProcess(config *config.ProcConfig) *Process {
	proc := &Process{
		config:     config,
		cmd:        nil,
		startTime:  time.Unix(0, 0),
		stopTime:   time.Unix(0, 0),
		state:      STOPPED,
		inStart:    false,
		stopByUser: false,
		retryTimes: 0,
	}

	//start the process if autostart is set to true
	if proc.isAutoStart() {
		proc.Start(false)
	}

	return proc
}

func (p *Process) Start(wait bool) {
	log.WithFields(log.Fields{"program": p.GetName()}).Info("try to start program")
	p.lock.Lock()
	if p.exitc != nil {
		close(p.exitc)
	}
	p.exitc = make(chan struct{}, 1)
	if p.inStart || p.state == RUNNING {
		log.WithFields(log.Fields{"program": p.GetName()}).Info("Don't start program again, program is already started")
		p.lock.Unlock()
		return
	}

	p.inStart = true
	p.stopByUser = false
	p.lock.Unlock()

	var m sync.Mutex
	runCond := sync.NewCond(&m)
	m.Lock()

	go func() {
		p.retryTimes = 0

		for {
			p.run(runCond)
			if (p.stopTime.Unix() - p.startTime.Unix()) < int64(p.getStartSeconds()) {
				p.retryTimes++
			} else {
				p.retryTimes = 0
			}
			if p.stopByUser {
				log.WithFields(log.Fields{"program": p.GetName()}).Info("Stopped by user, don't start it again")
				break
			}
			if !p.isAutoRestart() {
				log.WithFields(log.Fields{"program": p.GetName()}).Info("Don't start the stopped program because its autorestart flag is false")
				break
			}
			if p.retryTimes >= p.getStartRetries() && p.getStartRetries() > 0 {
				log.WithFields(log.Fields{"program": p.GetName()}).Info("Don't start the stopped program because its retry times ", p.retryTimes, " is greater than start retries ", p.getStartRetries())
				break
			}
		}
		p.lock.Lock()
		p.inStart = false
		p.exitc <- struct{}{}
		p.lock.Unlock()
	}()
	if wait {
		runCond.Wait()
	}
}

func (p *Process) GetName() string {
	return p.config.Name
}

func (p *Process) GetDescription() string {
	if p.state == RUNNING {
		d := time.Now().Sub(p.startTime)
		return fmt.Sprintf("pid %d, uptime %02d:%02d:%02d", p.cmd.Process.Pid, int(d.Hours()), int(d.Minutes()), int(d.Seconds()))
	} else if p.state != STOPPED {
		return p.stopTime.String()
	}
	return ""
}

func (p *Process) GetExitstatus() int {
	if p.state == EXITED || p.state == BACKOFF {
		status, ok := p.cmd.ProcessState.Sys().(syscall.WaitStatus)
		if ok {
			return status.ExitStatus()
		}
	}
	return 0
}

func (p *Process) GetPid() int {
	if p.state == STOPPED {
		return 0
	}
	return p.cmd.Process.Pid
}

// Get the process state
func (p *Process) GetState() ProcessState {
	return p.state
}

func (p *Process) GetStartTime() time.Time {
	return p.startTime
}

func (p *Process) GetStopTime() time.Time {
	switch p.state {
	case STARTING:
		fallthrough
	case RUNNING:
		fallthrough
	case STOPPING:
		return time.Unix(0, 0)
	default:
		return p.stopTime
	}
}

func (p *Process) GetStdoutLogfile() string {
	if p.config.Std.Outfile == "" {
		return "/dev/null"
	} else {
		return p.config.Std.Outfile
	}
}

func (p *Process) GetStderrLogfile() string {
	if p.config.Std.Errfile == "" {
		return "/dev/null"
	} else {
		return p.config.Std.Errfile
	}
}

func (p *Process) getStartSeconds() int {
	return p.config.Startsecs
}

func (p *Process) getStartRetries() int {
	return p.config.Startretries
}

func (p *Process) isAutoStart() bool {
	return p.config.Autostart
}

func (p *Process) SendProcessStdin(chars string) error {
	if p.stdin != nil {
		_, err := p.stdin.Write([]byte(chars))
		return err
	}
	return fmt.Errorf("NO_FILE")
}

// check if the process should be
func (p *Process) isAutoRestart() bool {
	return p.config.Autorestart

	//if !autoRestart  {
	//	return false
	//} else if autoRestart  {
	//	return true
	//} else {
	//	p.lock.Lock()
	//	defer p.lock.Unlock()
	//	if p.cmd != nil && p.cmd.ProcessState != nil {
	//		exitCode, err := p.getExitCode()
	//		return err == nil && p.inExitCodes(exitCode)
	//	}
	//}
	//return false

}

//func (p *Process) inExitCodes(exitCode int) bool {
//	for _, code := range p.getExitCodes() {
//		if code == exitCode {
//			return true
//		}
//	}
//	return false
//}

func (p *Process) getExitCode() (int, error) {
	if status, ok := p.cmd.ProcessState.Sys().(syscall.WaitStatus); ok {
		return status.ExitStatus(), nil
	}

	return -1, fmt.Errorf("no exit code")

}

//
//func (p *Process) getExitCodes() []int {
//	strExitCodes := strings.Split(p.config.GetString("exitcodes", "0,2"), ",")
//	result := make([]int, 0)
//	for _, val := range strExitCodes {
//		i, err := strconv.Atoi(val)
//		if err == nil {
//			result = append(result, i)
//		}
//	}
//	return result
//}

func (p *Process) run(runCond *sync.Cond) {
	args, err := parseCommand(p.config.Cmd)

	if err != nil {
		log.Error("the command is empty string")
		return
	}
	p.lock.Lock()
	if p.cmd != nil && p.cmd.ProcessState != nil {
		status := p.cmd.ProcessState.Sys().(syscall.WaitStatus)
		if status.Continued() {
			log.WithFields(log.Fields{"program": p.GetName()}).Info("Don't start program because it is running")
			p.lock.Unlock()
			return
		}
	}
	p.cmd = exec.Command(args[0])
	if len(args) > 1 {
		p.cmd.Args = args
	}
	if p.setUser() != nil {
		log.WithFields(log.Fields{"user": p.config.User}).Error("fail to run as user")
		p.lock.Unlock()
		return
	}
	p.cmd.SysProcAttr = &syscall.SysProcAttr{}
	set_deathsig(p.cmd.SysProcAttr)
	p.setEnv()
	//p.setDir()
	p.setLog()

	p.stdin, _ = p.cmd.StdinPipe()
	p.startTime = time.Now()
	p.changeStateTo(STARTING)
	err = p.cmd.Start()
	if err != nil {
		log.WithFields(log.Fields{"program": p.config.Name}).Error("fail to start program")
		p.changeStateTo(FATAL)
		p.stopTime = time.Now()
		p.lock.Unlock()
		runCond.Signal()
	} else {
		if p.stdoutLog != nil {
			p.stdoutLog.SetPid(p.GetPid())
		}
		if p.stderrLog != nil {
			p.stderrLog.SetPid(p.GetPid())
		}
		log.WithFields(log.Fields{"program": p.config.Name}).Info("success to start program")
		startSecs := p.getStartSeconds()
		//Set startsec to 0 to indicate that the program needn't stay
		//running for any particular amount of time.
		if startSecs > 0 {
			p.changeStateTo(RUNNING)
		}
		p.lock.Unlock()
		log.WithFields(log.Fields{"program": p.config.Name}).Debug("wait program exit")
		runCond.Signal()
		p.cmd.Wait()
		p.lock.Lock()
		p.stopTime = time.Now()
		if p.stopTime.Unix()-p.startTime.Unix() < int64(startSecs) {
			p.changeStateTo(BACKOFF)
		} else {
			p.changeStateTo(EXITED)
		}
		p.lock.Unlock()
	}

}

func (p *Process) changeStateTo(procState ProcessState) {
	//if p.config.IsProgram() {
	//	progName := p.config.GetProgramName()
	//	groupName := p.config.GetGroupName()
	//	if procState == STARTING {
	//		emitEvent(createPorcessStartingEvent(progName, groupName, p.state.String(), p.retryTimes))
	//	} else if procState == RUNNING {
	//		emitEvent(createPorcessRunningEvent(progName, groupName, p.state.String(), p.GetPid()))
	//	} else if procState == BACKOFF {
	//		emitEvent(createPorcessBackoffEvent(progName, groupName, p.state.String(), p.retryTimes))
	//	} else if procState == STOPPING {
	//		emitEvent(createPorcessStoppingEvent(progName, groupName, p.state.String(), p.GetPid()))
	//	} else if procState == EXITED {
	//		exitCode, err := p.getExitCode()
	//		expected := 0
	//		if err == nil && p.inExitCodes(exitCode) {
	//			expected = 1
	//		}
	//		emitEvent(createPorcessExitedEvent(progName, groupName, p.state.String(), expected, p.GetPid()))
	//	} else if procState == FATAL {
	//		emitEvent(createPorcessFatalEvent(progName, groupName, p.state.String()))
	//	} else if procState == STOPPED {
	//		emitEvent(createPorcessStoppedEvent(progName, groupName, p.state.String(), p.GetPid()))
	//	} else if procState == UNKNOWN {
	//		emitEvent(createPorcessUnknownEvent(progName, groupName, p.state.String()))
	//	}
	//}
	p.state = procState
}

func (p *Process) Signal(sig os.Signal) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.cmd != nil && p.cmd.Process != nil {
		return p.cmd.Process.Signal(sig)
	}

	return fmt.Errorf("process is not started")
}

func (p *Process) setEnv() {
	env := p.config.GetEnvStringArray()
	if len(env) != 0 {
		p.cmd.Env = append(os.Environ(), env...)
	} else {
		p.cmd.Env = os.Environ()
	}
}

//func (p *Process) setDir() {
//	dir := p.config.GetString("directory", "")
//	if dir != "" {
//		p.cmd.Dir = dir
//		fmt.Printf("Directory has been set to: %s\n", dir)
//	}
//}

func (p *Process) setLog() {
	var maxbytes, backups int
	if p.config.Std.Maxbytes == 0 {
		maxbytes = 50 * 1024 * 1024
	}
	if p.config.Std.Backups == 0 {
		backups = 10
	}

	p.stdoutLog = p.createLogger(p.config.Std.Outfile, int64(maxbytes), backups)

	p.cmd.Stdout = p.stdoutLog

	p.stderrLog = p.createLogger(p.config.Std.Errfile, int64(maxbytes), backups)

	p.cmd.Stderr = p.stderrLog
}

func (p *Process) createLogger(logFile string, maxBytes int64, backups int) Logger {
	var logger Logger
	logger = NewNullLogger()

	if logFile == "/dev/stdout" {
		logger = NewStdoutLogger()
	} else if logFile == "/dev/stderr" {
		logger = NewStderrLogger()
	} else if len(logFile) > 0 {
		logger = NewFileLogger(logFile, maxBytes, backups, NewNullLocker())
	}
	return logger
}

func (p *Process) setUser() error {
	userName := p.config.User
	if len(userName) == 0 {
		return nil
	}
	u, err := user.Lookup(userName)
	if err != nil {
		return err
	}
	p.cmd.SysProcAttr = &syscall.SysProcAttr{}
	uid, err := strconv.ParseUint(u.Uid, 10, 32)
	if err != nil {
		return err
	}
	gid, err := strconv.ParseUint(u.Gid, 10, 32)
	if err != nil {
		return err
	}
	p.cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
	return nil
}

//convert a signal name to signal
func toSignal(signalName string) os.Signal {
	if signalName == "HUP" {
		return syscall.SIGHUP
	} else if signalName == "INT" {
		return syscall.SIGINT
	} else if signalName == "QUIT" {
		return syscall.SIGQUIT
	} else if signalName == "KILL" {
		return syscall.SIGKILL
	} else if signalName == "USR1" {
		return syscall.SIGUSR1
	} else if signalName == "USR2" {
		return syscall.SIGUSR2
	} else {
		return syscall.SIGTERM

	}

}

//send signal to process to stop it
func (p *Process) Stop(wait bool) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.stopByUser = true
	if p.cmd != nil && p.cmd.Process != nil {
		log.WithFields(log.Fields{"program": p.GetName()}).Info("stop the program")
		p.cmd.Process.Signal(toSignal(p.config.Stopsignal))
		if wait {
			p.cmd.Process.Wait()
		}
	}
}

func (p *Process) Restart(wait bool) {
	if wait {
		p.Stop(wait)
		<-p.exitc
		p.Start(wait)
	} else {
		go func() {
			p.Stop(!wait)
			if _, ok := <-p.exitc; ok {
				p.Start(!wait)
			}
		}()
	}
}

func (p *Process) GetStatus() string {
	if p.cmd.ProcessState != nil {
		return p.cmd.ProcessState.String()
	}
	return p.state.String()
}
