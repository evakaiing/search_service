package storage

import (
	"encoding/xml"
	"io"
	"os"
)

var DatasetPath string = "dataset.xml"

type UserXML struct {
	Id            int    `xml:"id"`
	Guid          string `xml:"guid"`
	IsActive      bool   `xml:"isActive"`
	Balance       string `xml:"balance"`
	Picture       string `xml:"picture"`
	Age           int    `xml:"age"`
	EyeColor      string `xml:"eyeColor"`
	FirstName     string `xml:"first_name"`
	LastName      string `xml:"last_name"`
	Gender        string `xml:"gender"`
	Company       string `xml:"company"`
	Email         string `xml:"email"`
	Phone         string `xml:"phone"`
	Address       string `xml:"address"`
	About         string `xml:"about"`
	RegisterTime  string `xml:"registered"`
	FavoriteFruit string `xml:"favoriteFruit"`
}

type UsersXML struct {
	Users []UserXML `xml:"row"`
}

func LoadUsers(datasetPath string) (*UsersXML, error) {
	usersXML := new(UsersXML)
	file, err := os.Open(datasetPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	xmlData, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	err = xml.Unmarshal(xmlData, &usersXML)
	if err != nil {
		return nil, err
	}

	return usersXML, nil
}


