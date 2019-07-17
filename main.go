package main

import (
	"fmt"
	"net/http"
	"os"
	"rura/codetep/project"
	"rura/viewtep/autoriz"
	"rura/viewtep/mdbus"
	"rura/viewtep/vars"
	"strings"
	"time"
)

var drivers map[string]*mdbus.Driver
var routers map[string]*vars.Router
var users autoriz.Users
var aprov map[string]bool

func respAllSubsystems(w http.ResponseWriter, r *http.Request) {
	if !isLogged(r) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println("Need login ", r.RemoteAddr)
		return
	}
	res, err := vars.GetInfoRouters(routers)
	if err != nil {
		fmt.Println("Запрос ", err.Error())
		return
	}
	Sending(w, res)
}
func respSubsystemInfo(w http.ResponseWriter, r *http.Request) {
	if !isLogged(r) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println("Need login ", r.RemoteAddr)
		return
	}
	name := r.URL.Query().Get("name")
	rout := make([]*vars.Router, 0)
	ok := false
	for _, r := range routers {
		n := strings.Split(r.Name, ":")
		if n[0] == name {
			rout = append(rout, r)
			ok = true
		}
	}
	if !ok {
		fmt.Println("Запрос нет ", name)
		return

	}
	res, err := vars.GetFullInfo(rout)
	if err != nil {
		fmt.Println("Запрос ", err.Error())
		return
	}
	Sending(w, res)
}
func respSubsystemValue(w http.ResponseWriter, r *http.Request) {
	if !isLogged(r) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println("Need login ", r.RemoteAddr)
		return
	}
	name := r.URL.Query().Get("name")
	rout, ok := routers[name]
	if !ok {
		fmt.Println("Запрос нет ", name)
		return
	}
	res, err := rout.JSONGetVal()
	if err != nil {
		fmt.Println("Запрос ", err.Error())
		return
	}
	Sending(w, res)
}
func respAllModbuses(w http.ResponseWriter, r *http.Request) {
	if !isLogged(r) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println("Need login ", r.RemoteAddr)
		return
	}
	res, err := mdbus.GetInfoModbuses(drivers)
	if err != nil {
		fmt.Println("Запрос ", err.Error())
		return
	}
	Sending(w, res)
}
func respModbusInfo(w http.ResponseWriter, r *http.Request) {
	if !isLogged(r) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println("Need login ", r.RemoteAddr)
		return
	}
	name := r.URL.Query().Get("name")
	drv, ok := drivers[name]
	if !ok {
		fmt.Println("Запрос нет ", name)
		return

	}

	res, err := drv.GetFullInfo()
	if err != nil {
		fmt.Println("Запрос ", err.Error())
		return
	}
	Sending(w, res)
}
func respModbusValue(w http.ResponseWriter, r *http.Request) {
	if !isLogged(r) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println("Need login ", r.RemoteAddr)
		return
	}
	name := r.URL.Query().Get("name")
	drv, ok := drivers[name]
	if !ok {
		fmt.Println("Запрос нет ", name)
		return

	}

	res, err := drv.JSONGetVal()
	if err != nil {
		fmt.Println("Запрос ", err.Error())
		return
	}
	Sending(w, res)
}

//Sending send json to web
func Sending(w http.ResponseWriter, res []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(res)
}
func loginToSystem(w http.ResponseWriter, r *http.Request) {
	login := r.URL.Query().Get("login")
	password := r.URL.Query().Get("password")

	if !users.ChekUser(login, password) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println("Login bad ", login, password)

		return
	}
	ok, _ := aprov[r.RemoteAddr]
	if !ok {
		aprov[r.RemoteAddr] = true
	}
}
func respSetModbusValue(w http.ResponseWriter, r *http.Request) {
	if !isLogged(r) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println("Need login ", r.RemoteAddr)
		return
	}
	mname := r.URL.Query().Get("modbus")
	name := r.URL.Query().Get("name")
	value := r.URL.Query().Get("value")
	drv, ok := drivers[mname]
	if !ok {
		fmt.Println("Запрос нет ", mname, name, value)
		return

	}
	drv.WriteVariable(name, value)
}
func respSetSubsystemValue(w http.ResponseWriter, r *http.Request) {
	if !isLogged(r) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println("Need login ", r.RemoteAddr)
		return
	}
	sname := r.URL.Query().Get("subsystem")
	name := r.URL.Query().Get("name")
	value := r.URL.Query().Get("value")

	rout, ok := routers[sname]
	if !ok {
		fmt.Println("Запрос нет ", sname, name, value)
		return
	}
	rout.WriteVariable(name, value)
}
func isLogged(r *http.Request) bool {
	return true
	// ok, _ := aprov[r.RemoteAddr]
	// return ok
}
func gui() {
	aprov = make(map[string]bool)
	http.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {
		http.ServeFile(response, request, "./index.html")
	})
	http.HandleFunc("/login", loginToSystem)
	http.HandleFunc("/allSubs", respAllSubsystems)
	http.HandleFunc("/allModbuses", respAllModbuses)
	http.HandleFunc("/subinfo", respSubsystemInfo)
	http.HandleFunc("/modinfo", respModbusInfo)
	http.HandleFunc("/subvalue", respSubsystemValue)
	http.HandleFunc("/modvalue", respModbusValue)
	http.HandleFunc("/setsubval", respSetSubsystemValue)
	http.HandleFunc("/setmodval", respSetModbusValue)
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
	users, err = autoriz.LoadUsers("users.xml")
	if err != nil {
		globalError = true
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
