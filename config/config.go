package config

import (
	"io/ioutil"
	"log"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Name string
	Processes []struct{
		Name string   `yaml:"name"`
		StartCmd string `yaml:"startCmd"`
		StatusCmd string `yaml:"statusCmd"`
		StopCmd string `yaml:"stopCmd"`
		HealthCmd string `yaml:"healthCmd"`
	}
}

func LoadConfig() (*Config,error){
	t:= Config{}
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
