package mdbus

import (
	"encoding/json"
	"fmt"
	"time"
)

//JInfoModbuses json for all modbuses
type JInfoModbuses struct {
	Name     string    `json:"name"`
	Modbuses []JModbus `json:"modbuses"`
}

//JModbus jsom one modbus
type JModbus struct {
	Name    string    `json:"name"`
	IP1     string    `json:"ip1"`
	IP2     string    `json:"ip2"`
	Port    int       `json:"port"`
	LastOp1 time.Time `json:"lastop1"`
	LastOp2 time.Time `json:"lastop2"`
}

//JFullInfo для возврата JSON с полным описанием всех переменных Модбаса
type JFullInfo struct {
	Name      string `json:"name"`
	IP1       string `json:"ip1"`
	IP2       string `json:"ip2"`
	Registers []Reg  `json:"registers"`
}

//Reg один регистр модбаса
type Reg struct {
	Name        string `json:"name"`
	Type        int    `json:"type"`
	Format      int    `json:"format"`
	Description string `json:"desc"`
}

//JValues all value from ModBus
type JValues struct {
	Name   string  `json:"name"`
	Values []Value `json:"values"`
}

//Value one value
type Value struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

//JSONGetVal return json all value modbus
func (d *Driver) JSONGetVal() (string, error) {
	j := new(JValues)
	j.Name = d.Name
	coils := make([]bool, d.lenCoil)
	di := make([]bool, d.lenDI)
	ir := make([]uint16, d.lenIR)
	hr := make([]uint16, d.lenHR)
	if d.chanel == 0 {
		coils, di, ir, hr = d.tr.get()
	} else {
		coils, di, ir, hr = d.tr2.get()
	}

	for name, reg := range d.registers {
		val := new(Value)
		val.Name = name
		switch reg.regtype {
		case 0:
			v := "false "
			if reg.GetBool(coils, 0) {
				v = "true "
			}
			val.Value = v
		case 1:
			v := "false "
			if reg.GetBool(di, 0) {
				v = "true "
			}
			val.Value = v
		case 2:
			val.Value = reg.GetValue(ir)
		case 3:
			val.Value = reg.GetValue(hr)
		}
		j.Values = append(j.Values, *val)
	}
	res, err := json.Marshal(j)
	return string(res), err

}

//GetFullInfo json all modbus info
func (d *Driver) GetFullInfo() (string, error) {
	j := new(JFullInfo)
	j.Name = d.Name
	j.IP1 = d.IP
	j.IP2 = d.IP2
	for name, reg := range d.registers {
		r := new(Reg)
		r.Name = name
		r.Description = reg.description
		r.Format = reg.format
		r.Type = reg.regtype
		j.Registers = append(j.Registers, *r)
	}
	res, err := json.Marshal(j)
	return string(res), err
}

//GetInfoModbuses return json define all modbuses
func GetInfoModbuses(mbs map[string]*Driver) (string, error) {
	j := new(JInfoModbuses)
	j.Name = "modbuses"
	for _, mb := range mbs {
		jm := new(JModbus)
		jm.Name = mb.Name
		jm.IP1 = mb.IP
		jm.IP2 = mb.IP2
		jm.Port = mb.Port
		jm.LastOp1 = mb.LastOp(1)
		jm.LastOp2 = mb.LastOp(2)
		j.Modbuses = append(j.Modbuses, *jm)
	}
	res, err := json.Marshal(j)
	return string(res), err
}

//GetNames возвращает имена переменных с типом переменной
func (d *Driver) GetNames() map[string]int {
	r := make(map[string]int)
	for name, reg := range d.registers {
		r[name] = reg.Regtype()
	}
	return r
}

//GetDescription   возвращает имена переменных с описанием
func (d *Driver) GetDescription() map[string]string {
	r := make(map[string]string)
	for name, reg := range d.registers {
		r[name] = reg.description
	}
	return r
}

//GetTypes возвращает имена переменных с типом регистра
func (d *Driver) GetTypes() map[string]int {
	r := make(map[string]int)
	for name, reg := range d.registers {
		r[name] = reg.regtype
	}
	return r
}

//GetValues возвращает все переменные в символьном виде
func (d *Driver) GetValues() map[string]string {
	r := make(map[string]string)
	coils := make([]bool, d.lenCoil)
	di := make([]bool, d.lenDI)
	ir := make([]uint16, d.lenIR)
	hr := make([]uint16, d.lenHR)
	if d.chanel == 0 {
		coils, di, ir, hr = d.tr.get()
	} else {
		coils, di, ir, hr = d.tr2.get()
	}

	//fmt.Printf("%d %d %d %d ", len(coils), len(di), len(ir), len(hr))
	for name, reg := range d.registers {
		val := ""
		switch reg.regtype {
		case 0:
			for i := 0; i < reg.size; i++ {
				v := "false "
				if reg.GetBool(coils, i) {
					v = "true "
				}
				val += v
			}
		case 1:
			for i := 0; i < reg.size; i++ {
				v := "false "
				if reg.GetBool(di, i) {
					v = "true "
				}
				val += v
			}
		case 2:
			val = reg.GetValue(ir)
		case 3:
			val = reg.GetValue(hr)
		}
		r[name] = val
	}
	return r
}

//GetNamedValues возвращает запрошенные переменные в символьном виде
func (d *Driver) GetNamedValues(names map[string]interface{}) map[string]string {
	r := make(map[string]string)
	coils := make([]bool, d.lenCoil)
	di := make([]bool, d.lenDI)
	ir := make([]uint16, d.lenIR)
	hr := make([]uint16, d.lenHR)
	if d.chanel == 0 {
		coils, di, ir, hr = d.tr.get()
	} else {
		coils, di, ir, hr = d.tr2.get()
	}
	//fmt.Printf("%d %d %d %d ", len(coils), len(di), len(ir), len(hr))
	for name, reg := range d.registers {
		if _, ok := names[name]; !ok {
			continue
		}
		val := ""
		switch reg.regtype {
		case 0:
			for i := 0; i < reg.size; i++ {
				v := "false"
				if reg.GetBool(coils, i) {
					v = "true"
				}
				val += v
			}
		case 1:
			for i := 0; i < reg.size; i++ {
				v := "false"
				if reg.GetBool(di, i) {
					v = "true"
				}
				val += v
			}
		case 2:
			val = reg.GetValue(ir)
		case 3:
			val = reg.GetValue(hr)
		}
		r[name] = val
	}
	return r
}

//SetNamedValues записывает на вывод запрошенные переменные в символьном виде
func (d *Driver) SetNamedValues(names map[string]string) {
	// need insert code
	for name, value := range names {
		// fmt.Printf("dev:%s name:%s value=%s \n", d.name, name, value)
		reg, _ := d.registers[name]

		err := d.tr.writeVariable(reg, value)
		if err != nil {
			fmt.Println(err)
		}
		err = d.tr2.writeVariable(reg, value)
		if err != nil {
			fmt.Println(err)
		}

	}
}

// SwapBuffer переворачивает порядок байтов для Модбас
func SwapBuffer(buffer []byte) []byte {
	for i := 0; i < len(buffer)-1; i += 2 {
		buffer[i], buffer[i+1] = buffer[i+1], buffer[i]
	}
	return buffer
}

//SwapUint16 переворачивает байты внутри регистров
func SwapUint16(buffer []uint16) []uint16 {
	for i, val := range buffer {
		// val := buffer[i]
		buffer[i] = ((val & 0xff) << 8) | ((val >> 8) & 0xff)
	}
	return buffer
}
