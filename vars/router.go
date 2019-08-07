package vars

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net"
	"rura/codetep/project"
	"rura/teprol/logger"
	"strconv"
	"strings"
	"sync"
	"time"
)

//Router one rout for ip subsystem
type Router struct {
	Name      string
	IP        string
	Port      int
	Variables map[string]project.Variable
	NumValue  map[int]string
	Values    map[string][]string
	Arrays    []string
	MaxID     int
	LastOp    time.Time
	mu        sync.Mutex
	Connect   bool
	Status    string
}

//XMLRet for respnoce
type XMLRet struct {
	// Header xml.Name `xml:"vals"`
	Vals []Val `xml:"val"`
}

//Val one value
type Val struct {
	ID    int    `xml:"id,attr"`
	Value string `xml:"value,attr"`
}

//Init make one router
func Init(sub *project.Subsystem, sop project.Sub, ip string) (*Router, error) {
	r := new(Router)
	r.Name = sub.Name + ":" + ip
	r.IP = ip
	r.Port = 1080
	r.Values = make(map[string][]string)
	// r.Variables = make(map[string]project.Variable)
	r.Variables = sub.Variables
	r.NumValue = make(map[int]string)
	MaxID := 0
	for _, vr := range sub.Variables {
		if vr.ID > MaxID {
			MaxID = vr.ID
		}
		size, _ := strconv.Atoi(vr.Size)
		if size > 1 {
			r.Arrays = append(r.Arrays, vr.Name)
		}
		r.Values[vr.Name] = make([]string, size)
		r.NumValue[vr.ID] = vr.Name
	}
	r.MaxID = MaxID
	return r, nil
}

//Start start router
func (r *Router) Start() {
	go r.run()
}
func (r *Router) readArrays() {
	for _, name := range r.Arrays {
		v, ok := r.Variables[name]
		if !ok {
			fmt.Println(r.Name, name, " not found")
			return
		}
		conn, err := net.Dial("tcp4", r.IP+":"+strconv.Itoa(r.Port))
		if err != nil {
			r.Status = fmt.Sprintln(err.Error())
			r.Connect = false
			return
		}
		defer conn.Close()
		var result bytes.Buffer
		buffer := make([]byte, 130000)
		str := "A" + strconv.Itoa(v.ID) + " " + v.Size + " \000"
		_, err = conn.Write([]byte(str))
		if err != nil {
			fmt.Println(r.Name, err.Error())
			return
		}
		loop := true
		for loop {
			lenr, err := conn.Read(buffer)
			if err != nil {
				if err.Error() != "EOF" {
					fmt.Println(r.Name, err.Error())
					return
				}
			}
			if lenr == 0 {
				break
			}
			for index := 0; index < lenr; index++ {
				if buffer[index] == 0 {
					loop = false
				}
			}
			result.Write(buffer[:lenr])
		}
		conn.Close()
		// st := result.String()
		// fmt.Println(r.Name, v.Name, st)
		res := strings.Split(strings.TrimRight(strings.TrimLeft(result.String(), " "), " "), " ")
		size, _ := strconv.Atoi(v.Size)
		if len(res) != size {
			// fmt.Println(r.Name, v.Name, len(res), size, " bad size ")
			// fmt.Println(res)
			continue
		}
		r.mu.Lock()
		r.Values[v.Name] = res
		r.mu.Unlock()
	}
	if len(r.Arrays) > 0 {
		r.Connect = true
		r.Status = ""
	}
}
func (r *Router) readVariables() {
	loop := true
	stp := 500
	i := 1
	for loop {
		conn, err := net.Dial("tcp4", r.IP+":"+strconv.Itoa(r.Port))
		if err != nil {
			r.Status = fmt.Sprintln(err.Error())
			r.Connect = false
			return
		}
		defer conn.Close()
		j := i + stp
		if j > r.MaxID {
			j = r.MaxID
			loop = false
		}
		var result bytes.Buffer
		buffer := make([]byte, 130000)
		str := "R" + strconv.Itoa(i) + " " + strconv.Itoa(j) + "\000"
		_, err = conn.Write([]byte(str))
		if err != nil {
			fmt.Println(r.Name, err.Error())
			return
		}
		for true {
			lenr, err := conn.Read(buffer)
			if err != nil {
				if err.Error() != "EOF" {
					r.Status = fmt.Sprintln(err.Error())
					r.Connect = false
					return
				}
			}
			if lenr == 0 {
				break
			}
			for index := 0; index < lenr; index++ {
				if buffer[index] == 0 {
					buffer[index] = ' '
				}
			}
			result.Write(buffer[:lenr])

			if strings.Contains(result.String(), "</vals>\n") {
				break
			}
		}
		// fmt.Println(r.Name, result.String())
		conn.Close()

		x := new(XMLRet)
		err = xml.Unmarshal(result.Bytes(), &x)
		if err != nil {
			fmt.Println(r.Name, err.Error())
			return
		}
		r.mu.Lock()
		for _, v := range x.Vals {
			if v.ID == 0 {
				continue
			}
			name, ok := r.NumValue[v.ID]
			if !ok {
				fmt.Println(r.Name, v.ID, " not found")
				return
			}
			mas := r.Values[name]
			mas[0] = v.Value
			r.Values[name] = mas
		}
		r.mu.Unlock()
		i += stp

	}
	r.Connect = true
	r.Status = ""
}
func (r *Router) run() {
	step := time.Duration(1000) * time.Millisecond
	for {
		r.readVariables()
		r.readArrays()
		time.Sleep(step)
	}
}

//WriteVariable send value to subsystem
func (r *Router) WriteVariable(name string, value string) {
	v, ok := r.Variables[name]
	if !ok {
		logger.Error.Println("В", r.Name, "нет", name)
		return
	}

	conn, err := net.Dial("tcp4", r.IP+":"+strconv.Itoa(r.Port))
	if err != nil {
		logger.Error.Println(err.Error())
		return
	}
	defer conn.Close()
	str := "W" + strconv.Itoa(v.ID) + " " + value + "\000"
	_, err = conn.Write([]byte(str))
	if err != nil {
		logger.Error.Println(r.Name, err.Error())
		return
	}

}
