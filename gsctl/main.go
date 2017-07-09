package main

import (
	"os"
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"regexp"
	"strings"
)

func main(){
	if len(os.Args)==1 {
	   fmt.Println("usage:<export remote=127.0.0.1:9443;> gsctl [start|restart|stop|status] <processname>")
		fmt.Println("[...] is madatory, <...> is optional")

	}
    var address string

	for _,env:=range os.Environ(){
		if strings.Contains(env,"remote="){
			address=strings.Split(env,"=")[1]
		}
	}

	if len(os.Args)>1{
		if address==""{
			address="localhost:9443"
		}else if regexp.MustCompile(`:\d+`).Match([]byte(address)){
			address="localhost"+address
		}

		if len(os.Args)==2 {
			switch os.Args[1] {
			case "status":
				getResult("http://"+address+"/p/status")
				return
			case "start":
				getResult("http://"+address+"/p/start")
				return
			case "restart":
				getResult("http://"+address+"/p/restart")
				return
			case "stop":
				getResult("http://"+address+"/p/stop")
				return
			}
		}else {
			switch os.Args[1] {
			case "status":
				getResult("http://"+address+"/p/"+os.Args[2]+"/status")
				return
			case "start":
				getResult("http://"+address+"/p/"+os.Args[2]+"/start")
				return
			case "restart":
				getResult("http://"+address+"/p/"+os.Args[2]+"/restart")
				return
			case "stop":
				getResult("http://127.0.0.1:8081/p/"+os.Args[2]+"/stop")
				return
			}
		}



	}

}

type statusJson struct {
	Name string `json:"name"`
	Status string `json:"status"`
}

func getResult(url string){
	fmt.Println("connecting...")
	res, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}

	result, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	if err != nil {
		fmt.Println(err)
		return
	}

	var status []statusJson
	err=json.Unmarshal(result,&status)

	if err!=nil{
		fmt.Println(string(result))
	}else {
		for _,s:=range status{
			fmt.Println(s.Name+"	"+s.Status)
		}
	}
}
