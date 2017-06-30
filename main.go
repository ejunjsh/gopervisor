package main

import (
	"github.com/ejunjsh/gopervisor/config"
	log "github.com/Sirupsen/logrus"
)

func main(){
    _,err:=config.LoadConfig()
	if err!=nil{
		log.Fatal(err)
		return
	}

}
