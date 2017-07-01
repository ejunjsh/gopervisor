package config

import (
	"io/ioutil"
	"log"
	"gopkg.in/yaml.v2"
	"fmt"
)


type ProcConfig struct {
	Name string   `yaml:"name"`
	Cmd string `yaml:"cmd"`
	Env map[string] string `yaml:"env"`
	Std struct{
		Outfile string  `yaml:"outfile"`
		Errfile string  `yaml:"errfile"`
	}
	Startsecs int `yaml:"startsecs"`
	Startretries int `yaml:"startretries"`
	Autostart bool `yaml:"autostart"`
	Autorestart bool `yaml:"autorestart"`
	User string `yaml:"user"`
}

type NodeConfig struct {
	Name string
	Processes *[]ProcConfig
}

func LoadConfig(filepath string) (*NodeConfig,error){
	t:= NodeConfig{}
	b,err:=ioutil.ReadFile(filepath)
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

func (n *NodeConfig)GetProcConfig(name string) *ProcConfig {
	for _,proc:=range *n.Processes{
		if proc.Name==name{
			return &proc
		}
	}
	return nil
}

func (p *ProcConfig) GetEnvStringArray() []string{
	if len(p.Env) >0{
		result :=make([]string,0)
		for key,value := range p.Env{
			s:=fmt.Sprintf("%s=%s",key,value)
			result=append(result,s)
		}
		return  result
	}
	return []string{""}
}

