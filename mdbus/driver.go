package mdbus

import (
	"bytes"
	"fmt"
	"rura/codetep/project"
	"rura/teprol/logger"
	"strconv"
	"time"
)

//DeviceInOut интерфейс всех внешних драйверов ввода вывода
type DeviceInOut interface {
	start()
	get() ([]bool, []bool, []uint16, []uint16)
	writeVariable(*Register, string) error
	worked() bool
	lock()
	unlock()
	stop()
	lastop() time.Time
}

// Driver драйвер одного устройства
type Driver struct {
	Name        string
	Description string
	lenCoil     int
	lenDI       int
	lenIR       int
	lenHR       int
	registers   map[string]*Register
	connect     bool
	work        bool
	chanel      int
	tr          []DeviceInOut
	Step        int
	Restart     int
	IP          []string
	Port        []int
}

// Init подготавливает драйвер к работе
func Init(name string, dev project.Modbus, sub project.Sub) (*Driver, error) {
	driver := new(Driver)
	driver.Name = name
	driver.Description = dev.Description
	driver.Step = 500
	driver.Restart = 5
	driver.IP = make([]string, 2)
	driver.IP[0] = sub.Main
	driver.IP[1] = sub.Second
	driver.Port = make([]int, 2)
	driver.Port[0], _ = strconv.Atoi(dev.Port)
	driver.Port[1], _ = strconv.Atoi(dev.Port)
	// fmt.Println("driver " + driver.name)
	driver.registers = make(map[string]*Register, len(dev.Registers))
	//fmt.Println("")
	// Загружаем регистры
	//fmt.Println(DT.Name)
	for _, rec := range dev.Registers {
		reg := new(Register)
		reg.name = rec.Name
		reg.description = rec.Description
		reg.regtype = rec.Type
		reg.format = rec.Format
		reg.address = rec.Address
		reg.size = rec.Size
		reg.unitID = rec.UID
		reg.count(driver)
		driver.registers[reg.name] = reg
	}
	//fmt.Println("count",driver.lenCoil, driver.lenDI, driver.lenIR, driver.lenHR)
	driver.tr = make([]DeviceInOut, len(driver.IP))
	for ip := 0; ip < len(driver.IP); ip++ {
		con := fmt.Sprintf("%s:%d", driver.IP[ip], dev.Port[ip])
		d, err := master(driver, con)
		if err != nil {
			logger.Error.Println(err.Error())
			return driver, err
		}
		driver.tr[ip] = d
	}
	driver.work = true
	return driver, nil
}

func (d *Driver) loop() {

	step := 10 * time.Second
	if d.Restart != 0 {
		step = time.Duration(d.Restart) * time.Second
	}
	for {
		//start := time.Now()
		if !d.work {
			return
		}
		for canel := 0; canel < len(d.tr); canel++ {
			if d.tr[canel].worked() {
				d.chanel = canel
				continue
			}
			d.tr[canel].lock()
			con := fmt.Sprintf("%s:%d", d.IP[canel], d.Port[canel])
			dv, err := master(d, con)
			d.tr[canel].unlock()
			if err != nil {
				logger.Error.Println(err.Error())
			} else {
				d.tr[canel] = dv
				dv.start()
			}
		}
		time.Sleep(step)

	}
}

// Run запускает драйвер
func (d *Driver) Run() {
	for i := 0; i < len(d.tr); i++ {
		d.tr[i].start()
	}
	go d.loop()
	//fmt.Println("Запустили Драйвер " + d.name)
}

// Stop останавливает драйвер
func (d *Driver) Stop() {
	for i := 0; i < len(d.tr); i++ {
		d.tr[i].stop()
	}
	d.work = false
}

// Status возвращает статус драйвера
func (d *Driver) Status() bool {
	return d.work
}

//ToString выводит описание драйвера
func (d *Driver) ToString() string {
	s := bytes.NewBufferString("driver:" + d.Name + fmt.Sprintf(" rgs=%d ", len(d.registers)) + "\n")
	ss := fmt.Sprintf("%d %d %d %d \n", d.lenCoil, d.lenDI, d.lenIR, d.lenHR)
	s.WriteString(ss)
	for _, reg := range d.registers {
		s.WriteString(reg.ToString() + " \n")
	}
	return s.String()
}

//LastOp return time last operation on driver
func (d *Driver) LastOp(chanel int) time.Time {
	last := time.Unix(0, 0)
	for i := 0; i < len(d.tr); i++ {
		if d.tr[i].lastop().After(last) {
			last = d.tr[i].lastop()
		}
	}
	return last
}

//WriteVariable write value to out
func (d *Driver) WriteVariable(name string, value string) {
	reg, ok := d.registers[name]
	if !ok {
		return
	}
	for i := 0; i < len(d.tr); i++ {
		d.tr[i].writeVariable(reg, value)
	}
}
