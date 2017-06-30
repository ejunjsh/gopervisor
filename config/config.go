package config

import (
	"io/ioutil"
	"log"
	"gopkg.in/yaml.v2"
)


type ProcConfig struct {
	Name string   `yaml:"name"`
	Cmd string `yaml:"cmd"`
}

type NodeConfig struct {
	Name string
	Processes *[]ProcConfig
}

func LoadConfig() (*NodeConfig,error){
	t:= NodeConfig{}
	b,err:=ioutil.ReadFile("./node.yaml")
	if err!=nil{
		log.Print(err)
		return nil,err
	}
	err=yaml.Unmarshal(b,&t)
	if err!=nil{
		log.Print(err)
		return nil,err
	}
	return &t,nil
}

func (n *NodeConfig)GetProcConfig( name string) *ProcConfig {
	for _,proc:=range *n.Processes{
		if proc.Name==name{
			return &proc
		}
	}
	return nil
}

