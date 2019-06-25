package mdbus

import (
	"bytes"
	"fmt"
	"rura/codetep/project"
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
	tr          DeviceInOut
	tr2         DeviceInOut
	Step        int
	Restart     int

	IP    string
	Port  int
	IP2   string
	Port2 int
}

// Init подготавливает драйвер к работе
func Init(name string, dev project.Modbus, sub project.Sub) (*Driver, error) {
	driver := new(Driver)
	driver.Name = name
	driver.Description = dev.Description
	// driver.Step = 1000
	// driver.Restart = 10000
	driver.IP = sub.Main
	driver.IP2 = sub.Second
	driver.Port, _ = strconv.Atoi(dev.Port)
	driver.Port2, _ = strconv.Atoi(dev.Port)
	// fmt.Println("driver " + driver.name)
	driver.registers = make(map[string]*Register, len(dev.Registers))
	//fmt.Println("")
	// Загружаем регистры
	var err error
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
	con := fmt.Sprintf("%s:%s", driver.IP, dev.Port)
	d, err := master(driver, con)
	if err != nil {
		fmt.Println(err.Error())
		return driver, err
	}
	driver.tr = d
	con = fmt.Sprintf("%s:%s", driver.IP2, dev.Port)
	d, err = master(driver, con)
	if err != nil {
		fmt.Println(err.Error())
		return driver, err
	}
	driver.tr2 = d
	driver.work = true
	return driver, nil
}

func (d *Driver) loop() {

	step := 60 * time.Second
	if d.Restart != 0 {
		step = time.Duration(d.Restart) * time.Second
	}

	for {
		//start := time.Now()
		if !d.work {
			return
		}
		if d.tr.worked() {
			d.chanel = 0
		} else {
			d.tr.lock()
			con := fmt.Sprintf("%s:%d", d.IP, d.Port)
			dv, err := master(d, con)
			d.tr.unlock()
			if err != nil {
				fmt.Println(err.Error())
			} else {
				d.tr = dv
				dv.start()
			}
		}
		if d.tr2.worked() {
			d.chanel = 1
		} else {
			d.tr2.lock()
			con := fmt.Sprintf("%s:%d", d.IP2, d.Port2)
			dv, err := master(d, con)
			d.tr2.unlock()
			if err != nil {
				fmt.Println(err.Error())
			} else {
				d.tr2 = dv
				dv.start()
			}
		}
		//stop := time.Now()
		//elapsed := stop.Sub(start)
		// fmt.Println("driver "+d.name)
		time.Sleep(step)

	}
}

// Run запускает драйвер
func (d *Driver) Run() {
	d.tr.start()
	d.tr2.start()
	go d.loop()
	//fmt.Println("Запустили Драйвер " + d.name)
}

// Stop останавливает драйвер
func (d *Driver) Stop() {
	d.tr.stop()
	d.tr2.stop()
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
	if chanel == 1 {
		return d.tr.lastop()
	}
	return d.tr2.lastop()
}
