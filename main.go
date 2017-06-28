package main

import (
	"github.com/ejunjsh/gopervisor/config"
	"log"
)

func main(){
    c,err:=config.LoadConfig()
	if err!=nil{
		log.Fatal(err)
		return
	}


}
