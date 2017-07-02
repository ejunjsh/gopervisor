package config

import (
	"testing"

)

func TestLoadConfig(t *testing.T) {
	n,err:= LoadConfig("./node.yaml")
	if err!=nil{
		t.Error(err)
	}

	if len(*n.Processes)!=2{
		t.Error("wrong expectation.")
	}

	if (*n.Processes)[0].Env["env1"]!="env1"{
		t.Error("wrong env.")
	}

	if (*n.Processes)[0].Std.Errfile!="/opt/err"{
		t.Error("wrong std.")
	}

    if 	(*n.Processes)[0].GetEnvStringArray()[0]!="env1=env1"{
		t.Error("wrong env convertion.")
	}

}
