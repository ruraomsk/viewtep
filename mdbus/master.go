package mdbus

import (
	"fmt"
	"sync"
	"time"

	"github.com/goburrow/modbus"
)

//Master структура для мастера TCP modbus
type Master struct {
	master      *modbus.TCPClientHandler
	client      modbus.Client
	description string
	name        string
	coils       []bool
	di          []bool
	ir          []uint16
	hr          []uint16
	connected   bool
	work        bool
	mu          sync.Mutex
	Step        int
	con         string
	lastop      time.Time
}

func master(d *Driver, con string) (*Master, error) {
	m := new(Master)
	m.Step = d.Step
	m.name = d.Name
	m.con = con
	m.connected = false
	m.coils = make([]bool, d.lenCoil)
	m.di = make([]bool, d.lenDI)
	m.ir = make([]uint16, d.lenIR)
	m.hr = make([]uint16, d.lenHR)
	m.master = modbus.NewTCPClientHandler(con)
	m.master.Timeout = time.Second
	m.master.SlaveId = 1
	m.connected = false
	m.lastop = time.Unix(0, 0)
	// m.master.Logger = fmt
	return m, nil
}
func (m *Master) worked() bool {
	return m.work
}

func (m *Master) start() {
	err := m.master.Connect()
	if err != nil {
		fmt.Println(m.name + " start " + err.Error())
		return
	}
	m.client = modbus.NewClient(m.master)
	m.work = true
	go m.run()
	//fmt.Println("Master :" + m.con)

}
func (m *Master) stop() {
	m.master.Close()
	m.work = false
	return
}

func (m *Master) readAllCoils() {
	m.mu.Lock()
	defer m.mu.Unlock()
	var coils = uint16(len(m.coils))
	if coils == 0 {
		return
	}
	buff, err := m.client.ReadCoils(0, coils)
	if err != nil {
		fmt.Println(m.name + " coils " + err.Error())
		m.stop()
		return
	}

	//cmb.SwapBuffer(buff)
	for i := range m.coils {
		m.coils[i] = getbool(buff, uint16(i))
	}
}

func (m *Master) readAllDI() {
	m.mu.Lock()
	defer m.mu.Unlock()
	var di = uint16(len(m.di))
	if di == 0 {
		return
	}
	buff, err := m.client.ReadDiscreteInputs(0, di)
	if err != nil {
		fmt.Println(m.name + " di " + err.Error())
		m.stop()
		return
	}
	//	cmb.SwapBuffer(buff)

	for i := range m.di {
		m.di[i] = getbool(buff, uint16(i))
	}
}

func (m *Master) readAllIR() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.ir) == 0 {
		return
	}
	ref := uint16(0)
	for count := len(m.ir); count > 0; count -= 125 {
		len := count
		if count > 125 {
			len = 125
		}
		buff, err := m.client.ReadInputRegisters(ref, uint16(len))
		if err != nil {
			fmt.Println(m.name + " ir " + err.Error())
			m.stop()
			return
		}
		pos := ref
		left := 0
		SwapBuffer(buff)
		for i := 0; i < len; i++ {
			m.ir[pos] = (uint16(buff[left+1]) << 8) | uint16(buff[left])
			pos++
			left += 2
		}
		ref += 125
	}
}

func (m *Master) readAllHR() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.hr) == 0 {
		return
	}

	ref := uint16(0)
	for count := len(m.hr); count > 0; count -= 125 {
		len := count
		if count > 125 {
			len = 125
		}
		buff, err := m.client.ReadHoldingRegisters(ref, uint16(len))
		if err != nil {
			fmt.Println(m.name + " hr " + err.Error())
			m.stop()
			return
		}
		pos := ref
		left := 0
		SwapBuffer(buff)
		for i := 0; i < len; i++ {
			m.hr[pos] = (uint16(buff[left+1]) << 8) | uint16(buff[left])
			pos++
			left += 2
		}
		ref += 125
	}
}
func getbool(buffer []byte, pos uint16) bool {
	b := buffer[(pos / 8)]
	b = b >> (pos % 8)
	b = b & 1
	if b > 0 {
		return true
	}
	return false
}
func (m *Master) get() (coils []bool, di []bool, ir []uint16, hr []uint16) {
	m.mu.Lock()
	defer m.mu.Unlock()
	coils = make([]bool, len(m.coils))
	for i, val := range m.coils {
		coils[i] = val
	}
	di = make([]bool, len(m.di))
	for i, val := range m.di {
		di[i] = val
	}
	ir = make([]uint16, len(m.ir))
	for i, val := range m.ir {
		ir[i] = val
	}
	hr = make([]uint16, len(m.hr))
	for i, val := range m.hr {
		hr[i] = val
	}
	return
}

func (m *Master) run() {
	step := time.Duration(m.Step) * time.Millisecond
	for {

		//start := time.Now()
		if !m.work {
			break
		}

		m.readAllCoils()
		m.readAllDI()
		m.readAllIR()
		m.readAllHR()
		m.lastop = time.Now()
		if !m.work {
			break
		}

		//stop := time.Now()
		//elapsed := stop.Sub(start)
		// fmt.Println("master " + m.name)
		time.Sleep(step)
	}
	m.stop()
}
func (m *Master) writeVariable(reg *Register, value string) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.work {
		return
	}
	buffer, err := reg.SetValue(value)
	if err != nil {
		return
	}
	buf := make([]byte, len(buffer)*2)
	pos := 0
	// print(reg.name + " [")
	for i := 0; i < len(buffer); i++ {
		// print(buffer[i], "->")
		buf[pos+1] = byte(buffer[i] & 0xff)
		buf[pos+0] = byte((buffer[i] >> 8) & 0xff)
		// fmt.Print(buf[pos], buf[pos+1], "-")
		pos += 2
	}
	if len(buffer) == 0 {
		fmt.Printf("Длина буфера нулевая [%s]\n", value)
		return
	}
	// println("]")
	// println(reg.name, "=", value, " len=", len(buffer))
	if int(m.master.SlaveId) != reg.unitID {
		m.master.SlaveId = byte(reg.unitID)
	}
	if len(buffer) == 1 {
		// buffer[0] = ((buffer[0] & 0xff) << 8) | ((buffer[0] >> 8) & 0xff)
		switch reg.regtype {
		case 0:
			// println(reg.name, buffer[0])
			_, err = m.client.WriteSingleCoil(uint16(reg.address&0xffff), buffer[0])
		case 3:
			_, err = m.client.WriteSingleRegister(uint16(reg.address&0xffff), buffer[0])
		}
		return
	}
	switch reg.regtype {
	case 0:
		_, err = m.client.WriteMultipleCoils(uint16(reg.address&0xffff), uint16(len(buffer)), buf)
	case 3:
		// println(buf, len(buffer))
		_, err = m.client.WriteMultipleRegisters(uint16(reg.address&0xffff), uint16(len(buffer)), buf)
	}
	return
}
func (m *Master) lock() {
	m.mu.Lock()
}
func (m *Master) unlock() {
	m.mu.Unlock()
}
