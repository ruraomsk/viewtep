package vars

import (
	"encoding/json"
	"rura/codetep/project"
)

//JFullInfo для возврата JSON с полным описанием всех переменных Модбаса
type JFullInfo struct {
	Name string             `json:"name"`
	Vars []project.Variable `json:"variables"`
}

//JValues all value from ModBus
type JValues struct {
	Name   string  `json:"name"`
	Values []Value `json:"values"`
}

//Value one value
type Value struct {
	Name  string   `json:"name"`
	Value []string `json:"value"`
}

//JSONGetVal return json all value modbus
func (r *Router) JSONGetVal() (string, error) {
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
	return string(res), err

}

//GetFullInfo json all modbus info
func (r *Router) GetFullInfo() (string, error) {
	j := new(JFullInfo)
	j.Name = r.Name
	for _, v := range r.Variables {
		j.Vars = append(j.Vars, v)
	}
	res, err := json.Marshal(j)
	return string(res), err
}
