package vars

import (
	"encoding/json"
	"rura/codetep/project"
	"strings"
)

//JFullInfo для возврата JSON с полным описанием всех переменных
type JFullInfo struct {
	Name string             `json:"name"`
	IPs  []string           `json:"ips"`
	Vars []project.Variable `json:"variables"`
}

//JValues all value from
type JValues struct {
	Name   string  `json:"name"`
	Values []Value `json:"values"`
}

//Value one value
type Value struct {
	Name  string   `json:"name"`
	Value []string `json:"value"`
}

//JFullRout all routers info
type JFullRout struct {
	Name  string  `json:"name"`
	Routs []JRout `json:"routs"`
}

//JRout one router info
type JRout struct {
	Name    string   `json:"name"`
	IPs     []string `json:"ips"`
	Connect []bool   `json:"connect"`
	Status  []string `json:"status"`
}

//JSONGetVal return json all value modbus
func (r *Router) JSONGetVal() ([]byte, error) {
	j := new(JValues)
	j.Name = r.Name
	r.mu.Lock()
	for name, v := range r.Values {
		val := new(Value)
		val.Name = name
		val.Value = v
		j.Values = append(j.Values, *val)
	}
	res, err := json.Marshal(j)
	r.mu.Unlock()
	return res, err

}

//GetFullInfo json all modbus info
func GetFullInfo(r []*Router) ([]byte, error) {
	j := new(JFullInfo)
	name := strings.Split(r[0].Name, ":")
	j.Name = name[0]
	j.IPs = make([]string, 0)
	for _, rr := range r {
		j.IPs = append(j.IPs, rr.IP)
	}
	for _, v := range r[0].Variables {
		j.Vars = append(j.Vars, v)
	}
	res, err := json.Marshal(j)
	return res, err
}

//GetInfoRouters make json for all routers
func GetInfoRouters(rs map[string]*Router) ([]byte, error) {
	j := new(JFullRout)
	j.Name = "routers"
	for _, r := range rs {
		found := false
		name := strings.Split(r.Name, ":")
		for i := 0; i < len(j.Routs); i++ {
			rr := j.Routs[i]
			if name[0] == rr.Name {
				rr.IPs = append(rr.IPs, r.IP)
				rr.Connect = append(rr.Connect, r.Connect)
				rr.Status = append(rr.Status, r.Status)
				j.Routs[i] = rr
				found = true
				break
			}
		}
		if !found {
			jr := new(JRout)
			jr.Name = name[0]
			jr.IPs = make([]string, 0)
			jr.IPs = append(jr.IPs, r.IP)
			jr.Connect = make([]bool, 0)
			jr.Connect = append(jr.Connect, r.Connect)
			jr.Status = make([]string, 0)
			jr.Status = append(jr.Status, r.Status)
			j.Routs = append(j.Routs, *jr)

		}
	}
	res, err := json.Marshal(j)
	return res, err
}
