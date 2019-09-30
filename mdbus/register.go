package mdbus

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

//Register один регистр модбас
type Register struct {
	name        string
	description string
	regtype     int
	format      int
	address     int
	size        int
	unitID      int
}

//ToString Вывод в строку
func (r *Register) ToString() string {
	return fmt.Sprintf("%s %s %d %d %d %d %d", r.name, r.description, r.regtype, r.format, r.address, r.size, r.unitID)
}

//Regtype тип регистра для хранения 0-bool,1-int,2-float,3-long
func (r *Register) Regtype() int {
	if r.regtype < 2 {
		return 0
	}
	if r.format < 8 {
		return 1
	}
	if r.format >= 11 && r.format <= 13 {
		return 3
	}
	return 2
}

//GetValue Чтение переменной
func (r *Register) GetValue(buffer []uint16) (res string) {
	res = ""
	if r.regtype < 2 {
		return ""
	}
	for i := 0; i < r.size; i++ {

		switch r.Regtype() {
		case 1:
			val := r.GetInt(buffer, i)
			res += fmt.Sprintf("%d ", val)
		case 2:
			val := r.GetFloat(buffer, i)
			res += fmt.Sprintf("%f ", val)
		case 3:
			val := r.GetLong(buffer, i)
			res += fmt.Sprintf("%d ", val)
		}
	}
	return
}

//SetValue уставновить значение переменной
func (r *Register) SetValue(value string) (buffer []uint16, err error) {
	// fmt.Println(value)
	size := r.size
	if r.format >= 4 && r.format <= 9 {
		size *= 2
	}
	if r.format > 11 && r.format <= 15 {
		size *= 4
	}
	if r.format == 17 {
		size *= 2
	}
	buffer = make([]uint16, size)
	err = nil
	s := strings.Split(value, " ")
	// if len(s) != r.size {
	// 	err = fmt.Errorf("Несовпали размеры %s %s = %d", r.name, value, len(s))
	// 	return
	// }
	for i := 0; i < r.size; i++ {
		ss := s[i]
		for {
			j := strings.LastIndex(ss, " ")
			if j < 0 {
				break
			}
			ss = ss[0:j]
		}
		s[i] = ss
		// fmt.Printf("[%s] ", s[i])

		if r.regtype < 2 {
			b := false
			b, err = strconv.ParseBool(s[i])
			if err != nil {
				return
			}
			buffer[i] = 0
			if b {
				buffer[i] = 0xff00
				// println(r.name)
			}
		} else {

			switch r.Regtype() {
			case 1:
				var val int64
				val, err = strconv.ParseInt(s[i], 10, 32)
				//println(val)
				if err != nil {
					return
				}
				if r.format < 4 {
					//2 bytes
					buffer[i] = uint16(val & 0xffff)
					continue
				}
				if r.format < 8 {
					//4 bytes
					buffer[(i*2)+1] = uint16(val & 0xffff)
					buffer[i*2] = uint16((val >> 16) & 0xffff)
					// println(buffer[i], buffer[i+1])
				}
			case 2:
				var val float64
				if len(s[i]) == 0 {
					return
				}
				val, err = strconv.ParseFloat(s[i], 32)
				if err != nil {
					return
				}
				if r.format == 8 || r.format == 9 {
					//4 bytes
					b := math.Float32bits(float32(val))
					buffer[(i*2)+1] = uint16(b & 0xffff)
					buffer[i*2] = uint16((b >> 16) & 0xffff)
					continue
				}
				if r.format == 14 || r.format == 15 {
					//8 bytes
					b := math.Float64bits(val)
					buffer[(i*4)+3] = uint16((b >> 48) & 0xffff)
					buffer[(i*4)+2] = uint16((b >> 32) & 0xffff)
					buffer[(i*4)+1] = uint16((b >> 16) & 0xffff)
					buffer[(i*4)+0] = uint16(b & 0xffff)
				}

			case 3:
				var val int64
				val, err = strconv.ParseInt(s[i], 10, 64)
				if err != nil {
					return
				}
				if r.format >= 11 && r.format <= 13 {
					//4 bytes
					buffer[(i*4)+3] = uint16((val >> 48) & 0xffff)
					buffer[(i*4)+2] = uint16((val >> 32) & 0xffff)
					buffer[(i*4)+1] = uint16((val >> 16) & 0xffff)
					buffer[i*4] = uint16(val & 0xffff)
				}

			}

		}

	}
	return
}

//GetBool возвращает bool
func (r *Register) GetBool(buffer []bool, pos int) bool {
	if r.regtype > 1 {
		return false
	}
	return buffer[r.address+pos]
}

//GetInt возвращает int
func (r *Register) GetInt(buffer []uint16, pos int) int {

	if r.format < 4 {
		//fmt.Print(len(buffer), " getInt "+r.ToString()+"\n")
		//2 bytes
		if r.format == 2 {
			return int(uint16(buffer[r.address+pos]))

		}
		//fmt.Println(r.name, r.address, r.regtype, len(buffer))
		return int(int16(buffer[r.address+pos]))
	}
	if r.format < 8 {
		//4 bytes
		pos *= 2
		return int(buffer[r.address+pos])<<16 | int(buffer[r.address+pos+1])
	}
	return 0
}

//GetLong возвращает long
func (r *Register) GetLong(buffer []uint16, pos int) int64 {
	pos = pos * 4
	if r.format >= 11 && r.format <= 13 {
		//4 bytes
		return int64(buffer[r.address+pos])<<48 | int64(buffer[r.address+pos+1])<<32 | int64(buffer[r.address+pos+2])<<16 | int64(buffer[r.address+pos+3])
	}
	return 0
}

//GetFloat возвращает float
func (r *Register) GetFloat(buffer []uint16, pos int) float64 {
	if r.format == 8 || r.format == 9 {
		//4 bytes
		pos = pos * 2
		r := uint32(buffer[r.address+pos])<<16 | uint32(buffer[r.address+pos+1])
		return float64(math.Float32frombits(r))
	}
	if r.format == 14 || r.format == 15 {
		//4 bytes
		pos = pos * 4
		r := uint64(buffer[r.address+pos])<<48 | uint64(buffer[r.address+pos+1])<<32 | uint64(buffer[r.address+pos+2])<<16 | uint64(buffer[r.address+pos+3])
		return math.Float64frombits(r)
	}
	return 0.0
}
func (r *Register) count(d *Driver) {
	size := r.size
	if r.format >= 4 && r.format <= 9 {
		size *= 2
	}
	if r.format > 11 && r.format <= 15 {
		size *= 4
	}
	if r.format == 17 {
		size *= 2
	}
	right := r.address + size
	switch r.regtype {
	case 0:
		if right > d.lenCoil {
			d.lenCoil = right
		}
	case 1:
		if right > d.lenDI {
			d.lenDI = right
		}
	case 2:
		if right > d.lenIR {
			d.lenIR = right
		}
	case 3:
		if right > d.lenHR {
			d.lenHR = right
		}
	}
}
