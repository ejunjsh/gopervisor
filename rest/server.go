package rest

import (
	"net/http"
	"regexp"
	"fmt"
)

type Server struct {
	Adress string

}

func (s *Server) Run(){
	http.ListenAndServe(s.Adress,s)
}


var pattern=regexp.MustCompile(`/gopervisor/(\w+)/(\d+)/(\w+)`)

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request){

	if pattern.MatchString(r.URL.Path){
		matches:=pattern.FindSubmatch([]byte(r.URL.Path))
		node:=string(matches[1])
        pid:=string(matches[2])
		operation:=string(matches[3])
		w.Write([]byte(fmt.Sprintf("node:%s \n pid:%s \n operation:%s \n",node,pid,operation)))
		return
	}

	w.WriteHeader(404)
	w.Write([]byte("nothing found!"))

}