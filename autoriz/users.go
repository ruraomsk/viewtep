package autoriz

import (
	"encoding/xml"
	"github.com/ruraomsk/TLServer/logger"
	"io/ioutil"
)

//Users list all users
type Users struct {
	HeadUsers xml.Name `xml:"users"`
	Users     []User   `xml:"user"`
}

//User define one user
type User struct {
	Login    string `xml:"login,attr"`
	Password string `xml:"password,attr"`
	Name     string `xml:"name,attr,omitempty"`
}

//LoadUsers загружает всех пользователей
func LoadUsers(path string) (Users, error) {
	var us Users
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Error.Println("Error! " + err.Error())
		return us, err
	}
	err = xml.Unmarshal(buf, &us)
	return us, err
}

//ChekUser return true if login and password is correct
func (us *Users) ChekUser(user string, password string) bool {
	for _, u := range us.Users {
		if (u.Login == user) && (u.Password == password) {
			return true
		}
	}
	return false
}
