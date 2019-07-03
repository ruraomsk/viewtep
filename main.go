package main

import (
	"fmt"
	"net/http"
	"os"
	"rura/codetep/project"
	"rura/viewtep/mdbus"
	"rura/viewtep/vars"
	"time"
)

var drivers map[string]*mdbus.Driver
var routers map[string]*vars.Router

func respAllSubsystems(w http.ResponseWriter, r *http.Request) {
	res, err := vars.GetInfoRouters(routers)
	if err != nil {
		fmt.Println("Запрос ", err.Error())
		return
	}
	w.Write(res)
}
func respSubsystemInfo(w http.ResponseWriter, r *http.Request) {
	for _, rout := range routers {
		res, err := rout.GetFullInfo()
		if err != nil {
			fmt.Println("Запрос ", err.Error())
			return
		}
		w.Write(res)
		w.Write([]byte("\n*****************************************\n"))
	}
}
func respSubsystemValue(w http.ResponseWriter, r *http.Request) {
	for _, rout := range routers {
		res, err := rout.JSONGetVal()
		if err != nil {
			fmt.Println("Запрос ", err.Error())
			return
		}
		w.Write(res)
		w.Write([]byte("\n*****************************************\n"))
	}
}
func respAllModbuses(w http.ResponseWriter, r *http.Request) {
	res, err := mdbus.GetInfoModbuses(drivers)
	if err != nil {
		fmt.Println("Запрос ", err.Error())
		return
	}
	w.Write(res)
}
func respModbusInfo(w http.ResponseWriter, r *http.Request) {

	for _, drv := range drivers {
		res, err := drv.GetFullInfo()
		if err != nil {
			fmt.Println("Запрос ", err.Error())
			return
		}
		w.Write(res)
		w.Write([]byte("\n*****************************************\n"))
	}
}
func respModbusValue(w http.ResponseWriter, r *http.Request) {

	for _, drv := range drivers {
		res, err := drv.JSONGetVal()
		if err != nil {
			fmt.Println("Запрос ", err.Error())
			return
		}
		w.Write(res)
		w.Write([]byte("\n*****************************************\n"))
	}
}
func gui() {
	http.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {
		http.ServeFile(response, request, "./index.html")
	})
	http.HandleFunc("/allSubs", respAllSubsystems)
	http.HandleFunc("/allModbuses", respAllModbuses)
	http.HandleFunc("/subinfo", respSubsystemInfo)
	http.HandleFunc("/modinfo", respModbusInfo)
	http.HandleFunc("/subvalue", respSubsystemValue)
	http.HandleFunc("/modvalue", respModbusValue)
	fmt.Println("Listering on port 8080")
	http.ListenAndServe(":8080", nil)
}
func main() {

	fmt.Println("Начало работы...")
	prPath := ""
	if len(os.Args) == 1 {
		prPath = "/home/rura/dataSimul/pr"
	} else {
		prPath = os.Args[1]
	}
	fmt.Println("Проект загружается из ", prPath)
	pr, err := project.LoadProject(prPath)
	if err != nil {
		fmt.Println("Найдены ошибки " + err.Error())
		return
	}
	pr.DefDrivers, err = project.LoadAllDrivers(prPath + "/settings/default")
	if err != nil {
		fmt.Println("Найдены ошибки " + err.Error())
		return
	}
	globalError := false
	pr.Models, err = project.LoadAllModels(prPath + "/settings/models")
	drivers = make(map[string]*mdbus.Driver)
	routers = make(map[string]*vars.Router)
	for _, s := range pr.Subs {
		sub := pr.Subsystems[s.Name]
		for _, mb := range sub.Modbuses {
			if mb.Type == "master" {
				continue
			}
			dbus, err := mdbus.Init(sub.Name+":"+mb.Name, mb, s)
			if err != nil {
				fmt.Println("Modbus " + sub.Name + ":" + mb.Name + " Error " + err.Error())
				globalError = true
			} else {
				drivers[dbus.Name] = dbus
			}
		}
		vr, err := vars.Init(sub, s, s.Main)
		if err != nil {
			fmt.Println("Rout " + sub.Name + " Error " + err.Error())
			globalError = true
		} else {
			routers[vr.Name] = vr
		}
		if s.Main != s.Second {
			vr, err := vars.Init(sub, s, s.Second)
			if err != nil {
				fmt.Println("Rout " + sub.Name + " Error " + err.Error())
				globalError = true
			} else {
				routers[vr.Name] = vr
			}
		}

	}
	if globalError {
		fmt.Println("Дальнейшая работа невозможна!")
		return
	}
	for _, r := range routers {
		r.Start()
	}
	for _, d := range drivers {
		d.Run()
	}
	go gui()
	for true {
		time.Sleep(10 * time.Second)
	}
	fmt.Println("Конец работы")
}
