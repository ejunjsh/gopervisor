package config

import (
	"testing"
	"fmt"
)

func TestLoadConfig(t *testing.T) {
	n,err:= LoadConfig()
	if err!=nil{
		t.Error(err)
	}

	if len(*n.Processes)!=2{
		t.Error("wrong expectation.")
	}


	fmt.Println(n)
}
