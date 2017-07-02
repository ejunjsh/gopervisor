package main

import (
	"github.com/ejunjsh/gopervisor/node"
	"github.com/ejunjsh/gopervisor/config"
	"sync"
)

func main(){
	wg:= sync.WaitGroup{}
	wg.Add(1)
	process:=node.NewProcess(&config.ProcConfig{Name:"test",Cmd:"sh ./test.sh1",Std:config.StdConfig{Outfile:"/Users/zhouff/abc",Errfile:"/Users/zhouff/abc"},Autorestart:true})



	process.Start(true)
	wg.Wait()
}

