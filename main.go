package main

import (
	"flag"
	"github.com/ejunjsh/gopervisor/node"
	"github.com/ejunjsh/gopervisor/config"
	log "github.com/Sirupsen/logrus"
	"github.com/ejunjsh/gorest"
)

type statusJson struct {
	Name string `json:"name"`
	Status string `json:"status"`
}

func main(){
	var confFile,address string
	flag.StringVar(&confFile, "configuration", "", "configuration file path")
	flag.StringVar(&confFile, "c", "", "configuration file path.")
	flag.StringVar(&address, "address", "", "Network host to listen on.")
	flag.StringVar(&address, "a", "", "Network host to listen on.")

	flag.Parse()

	c,err:=config.LoadConfig(confFile)
	if err!=nil{
		log.Error(err)
	}

	n:=node.NewNode(c)
	n.StartSupervisor()

	r:=gorest.NewApp()
	r.Get("/p/:op", func(r *gorest.HttpRequest, w gorest.HttpResponse) error {
		statusr:= make([]statusJson,0)
		if op,ok:=r.PathParams["op"];ok{
			switch op {
			case "status":
				n.ForEachProcess(func(p *node.Process) {
					statusr=append(statusr, statusJson{p.GetName(),p.GetStatus()})
				})
				return w.WriteJson(statusr)
			case "start":
				n.ForEachProcess(func(p *node.Process) {
					p.Start(false)
				})
				return w.WriteString("OK")
			case "stop":
				n.ForEachProcess(func(p *node.Process) {
					p.Stop(false)
				})
				return w.WriteString("OK")
			case "restart":
				n.ForEachProcess(func(p *node.Process) {
					p.Restart(false)
				})
				return w.WriteString("OK")
			}
		}
		return nil
	})
	r.Get("/p/:name/:op", func(r *gorest.HttpRequest, w gorest.HttpResponse) error {
		statusr:= make([]statusJson,0)
		if name,ok:=r.PathParams["name"];ok{
			p:=n.Find(name)
			if p!=nil{
				if op,ok:=r.PathParams["op"];ok {
					switch op {
					case "status":
						statusr = append(statusr, statusJson{p.GetName(), p.GetStatus()})
						return w.WriteJson(statusr)
					case "start":
						p.Start(false)
						return w.WriteString("OK")
					case "stop":
						p.Stop(false)
						return w.WriteString("OK")
					case "restart":
						p.Restart(false)
						return w.WriteString("OK")
					}
				}
			}
		}
		return nil
	})
	r.Run(address)
}

