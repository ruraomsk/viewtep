package main

import (
	"fmt"
	"os"
	"rura/codetep/project"
	"rura/viewtep/mdbus"
	"rura/viewtep/vars"
	"time"
)

func main() {

	fmt.Println("Начало работы...")
	prPath := ""
	if len(os.Args) == 1 {
		prPath = "/home/rura/dataSimul/prnew"
	} else {
		prPath = os.Args[1]
	}
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
	pr.Models, err = project.LoadAllModels(prPath + "/settings/models")
	drivers := make(map[string]*mdbus.Driver)
	routers := make(map[string]*vars.Router)
	for _, s := range pr.Subs {
		sub := pr.Subsystems[s.Name]
		for _, mb := range sub.Modbuses {
			if mb.Type == "master" {
				continue
			}
			dbus, err := mdbus.Init(sub.Name+":"+mb.Name, mb, s)
			if err != nil {
				fmt.Println("Modbus " + sub.Name + ":" + mb.Name + " Error " + err.Error())
			} else {
				drivers[dbus.Name] = dbus
			}
		}
		vr, err := vars.Init(sub, s, s.Main)
		if err != nil {
			fmt.Println("Rout " + sub.Name + " Error " + err.Error())
		} else {
			routers[vr.Name] = vr
		}
		if s.Main != s.Second {
			vr, err := vars.Init(sub, s, s.Second)
			if err != nil {
				fmt.Println("Rout " + sub.Name + " Error " + err.Error())
			} else {
				routers[vr.Name] = vr
			}
		}

	}
	fmt.Println("Modbuses.....")
	for _, drv := range drivers {
		drv.Run()
		js, err := drv.GetFullInfo()
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println(js)
		}
	}
	fmt.Println("Variables.....")

	for _, rout := range routers {
		rout.Start()
		js, err := rout.GetFullInfo()
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println(js)
		}
	}
	// for _, rout := range routers {
	// }
	for true {
		time.Sleep(10 * time.Second)
		for _, drv := range drivers {
			js, err := drv.JSONGetVal()
			if err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println(js)
			}
		}
		for _, rout := range routers {
			js, err := rout.JSONGetVal()
			if err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println(js)
			}
		}
		js, err := vars.GetInfoRouters(routers)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println(js)
		}
		js, err = mdbus.GetInfoModbuses(drivers)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println(js)
		}
		break
	}
	fmt.Println("Конец работы")
}
