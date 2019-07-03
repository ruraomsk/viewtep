package vars

import (
	"encoding/json"
	"rura/codetep/project"
)

//JFullInfo для возврата JSON с полным описанием всех переменных
type JFullInfo struct {
	Name string             `json:"name"`
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
	Name    string `json:"name"`
	Connect bool   `json:"connect"`
	Status  string `json:"status"`
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
func (r *Router) GetFullInfo() ([]byte, error) {
	j := new(JFullInfo)
	j.Name = r.Name
	for _, v := range r.Variables {
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
		jr := new(JRout)
		jr.Name = r.Name
		jr.Connect = r.Connect
		jr.Status = r.Status
		j.Routs = append(j.Routs, *jr)
	}
	res, err := json.Marshal(j)
	return res, err
}
