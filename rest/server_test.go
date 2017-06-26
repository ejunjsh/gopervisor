package rest

import (
	"testing"
	"net/http"
	"io/ioutil"
	"fmt"
)

func TestServer_Run(t *testing.T) {
	server:=Server{Adress:":8080"}
	go server.Run()

	res, err := http.Get("http://127.0.0.1:8080/gopervisor/abc/1/status")
	if err != nil {
		t.Fatal(err)
	}


	got, gerr := ioutil.ReadAll(res.Body)
	res.Body.Close()

	if gerr != nil {
		t.Fatal(gerr)
	}

	want:=fmt.Sprintf("node:%s \n pid:%s \n operation:%s \n","abc","1","status")
	if string(got) != want {
		t.Errorf("got = %s, want %s", got, want)
	}
}
