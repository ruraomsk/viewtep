package mdbus

import (
	"encoding/json"
	"github.com/ruraomsk/TLServer/logger"
	"time"
)

//JInfoModbuses json for all modbuses
type JInfoModbuses struct {
	Name     string    `json:"name"`
	Modbuses []JModbus `json:"modbuses"`
}

//JModbus jsom one modbus
type JModbus struct {
	Name   string      `json:"name"`
	IP     []string    `json:"ips"`
	Port   []int       `json:"ports"`
	LastOp []time.Time `json:"lastop"`
}

//JFullInfo для возврата JSON с полным описанием всех переменных Модбаса
type JFullInfo struct {
	Name      string   `json:"name"`
	IP        []string `json:"ips"`
	Registers []Reg    `json:"registers"`
}

//Reg один регистр модбаса
type Reg struct {
	Name        string `json:"name"`
	Type        int    `json:"type"`
	Format      int    `json:"format"`
	Description string `json:"desc"`
	Address     int    `json:"address"`
	Size        int    `json:"size"`
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
func (d *Driver) JSONGetVal() ([]byte, error) {
	j := new(JValues)
	j.Name = d.Name
	coils := make([]bool, d.lenCoil)
	di := make([]bool, d.lenDI)
	ir := make([]uint16, d.lenIR)
	hr := make([]uint16, d.lenHR)
	coils, di, ir, hr = d.tr[d.chanel].get()

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
	return res, err

}

//GetFullInfo json all modbus info
func (d *Driver) GetFullInfo() ([]byte, error) {
	j := new(JFullInfo)
	j.Name = d.Name
	j.IP = make([]string, len(d.IP))
	for i := 0; i < len(d.IP); i++ {
		j.IP[i] = d.IP[i]
	}
	for name, reg := range d.registers {
		r := new(Reg)
		r.Name = name
		r.Description = reg.description
		r.Format = reg.format
		r.Type = reg.regtype
		r.Size = reg.size
		r.Address = reg.address
		j.Registers = append(j.Registers, *r)
	}
	res, err := json.Marshal(j)
	return res, err
}

//GetInfoModbuses return json define all modbuses
func GetInfoModbuses(mbs map[string]*Driver) ([]byte, error) {
	j := new(JInfoModbuses)
	j.Name = "modbuses"
	for _, mb := range mbs {
		jm := new(JModbus)
		jm.IP = make([]string, len(mb.tr))
		jm.Port = make([]int, len(mb.tr))
		jm.LastOp = make([]time.Time, len(mb.tr))
		jm.Name = mb.Name
		for i := 0; i < len(mb.tr); i++ {
			jm.IP[i] = mb.IP[i]
			jm.Port[i] = mb.Port[i]
			jm.LastOp[i] = mb.LastOp(i)

		}
		j.Modbuses = append(j.Modbuses, *jm)
	}
	res, err := json.Marshal(j)
	return res, err
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
	d.setChanel()
	coils, di, ir, hr = d.tr[d.chanel].get()

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
func (d *Driver) setChanel() {
	last := time.Unix(0, 0)
	chanel := 0
	for i := 0; i < len(d.tr); i++ {
		if d.tr[i].lastop().After(last) {
			last = d.tr[i].lastop()
			chanel = i
		}
	}
	d.chanel = chanel

}

//GetNamedValues возвращает запрошенные переменные в символьном виде
func (d *Driver) GetNamedValues(names map[string]interface{}) map[string]string {
	r := make(map[string]string)
	coils := make([]bool, d.lenCoil)
	di := make([]bool, d.lenDI)
	ir := make([]uint16, d.lenIR)
	hr := make([]uint16, d.lenHR)
	d.setChanel()
	coils, di, ir, hr = d.tr[d.chanel].get()
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
		for i := 0; i < len(d.tr); i++ {
			err := d.tr[i].writeVariable(reg, value)
			if err != nil {
				logger.Error.Println(err)
			}

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
