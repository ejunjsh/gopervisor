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

	fmt.Println(n)
}
