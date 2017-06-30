package node

import (
	"sort"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
)

type Node struct {
	name           string
	procs          map[string]*Process
	lock           sync.Mutex
}

func newNode() *Node {
	return &Node{procs: make(map[string]*Process)}
}


func (pm *Node) CreateProcess(supervisor_id string, config *ConfigEntry) *Process {
	procName := config.GetProgramName()

	proc, ok := pm.procs[procName]

	if !ok {
		proc = NewProcess(supervisor_id, config)
		pm.procs[procName] = proc
	}
	log.Info("create process:", procName)
	return proc
}


func (pm *Node) Add(name string, proc *Process) {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	pm.procs[name] = proc
	log.Info("add process:", name)
}

func (pm *Node) Remove(name string) *Process {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	proc, _ := pm.procs[name]
	delete(pm.procs, name)
	log.Info("remove process:", name)
	return proc
}

// return process if found or nil if not found
func (pm *Node) Find(name string) *Process {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	proc, ok := pm.procs[name]
	if ok {
		log.Debug("succeed to find process:", name)
	} else {
		//remove group field if it is included
		if pos := strings.Index(name, ":"); pos != -1 {
			proc, ok = pm.procs[name[pos+1:]]
		}
		if !ok {
			log.Info("fail to find process:", name)
		}
	}
	return proc
}

// clear all the processes
func (pm *Node) Clear() {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	pm.procs = make(map[string]*Process)
}

func (pm *Node) ForEachProcess(procFunc func(p *Process)) {
	pm.lock.Lock()
	defer pm.lock.Unlock()

	procs := pm.getAllProcess()
	for _, proc := range procs {
		procFunc(proc)
	}
}

func (pm *Node) getAllProcess() []*Process {
	tmpProcs := make([]*Process, 0)
	for _, proc := range pm.procs {
		tmpProcs = append(tmpProcs, proc)
	}
	sort.Sort(ProcessSortByPriority(tmpProcs))
	return tmpProcs
}


