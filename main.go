package main

import (
	"flag"
	"github.com/ejunjsh/gopervisor/node"
	"github.com/ejunjsh/gopervisor/config"
	log "github.com/Sirupsen/logrus"
	"github.com/ejunjsh/gopervisor/rest"
)



func main(){
	var confFile,address string
	flag.StringVar(&confFile, "configuration", "", "configuration file path")
	flag.StringVar(&confFile, "c", "", "configuration file path.")
	flag.StringVar(&address, "address", "", "Network host to listen on.")
	flag.StringVar(&address, "a", "", "Network host to listen on.")

	flag.Parse()

	c,err:=config.LoadConfig(confFile)
	if err!=nil{
		log.Error(err)
	}

	n:=node.NewNode(c)
	n.StartSupervisor()

	s:=rest.Server{Address:address,Node:n}
	s.Run()
}

