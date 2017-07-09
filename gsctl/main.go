package main

import (
	"os"
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"
)

func main(){
	if len(os.Args)==2{
		switch os.Args[1] {
		case "status":
			getResult("http://127.0.0.1:8081/p/status")
			return
		case "start":
			getResult("http://127.0.0.1:8081/p/start")
			return
		case "restart":
			getResult("http://127.0.0.1:8081/p/restart")
			return
		case "stop":
			getResult("http://127.0.0.1:8081/p/stop")
			return
		}
	}

	if len(os.Args)==3{

		switch os.Args[1] {
		case "status":
			getResult("http://127.0.0.1:8081/p/"+os.Args[2]+"/status")
			return
		case "start":
			getResult("http://127.0.0.1:8081/p/"+os.Args[2]+"/start")
			return
		case "restart":
			getResult("http://127.0.0.1:8081/p/"+os.Args[2]+"/restart")
			return
		case "stop":
			getResult("http://127.0.0.1:8081/p/"+os.Args[2]+"/stop")
			return
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
	}

	result, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	if err != nil {
		fmt.Println(err)
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
