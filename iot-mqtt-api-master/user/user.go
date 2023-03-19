package user

import (
	"fmt"
	"net/http"
)

type User struct {
	Id        int32  `json:"ID"`
	FirstName string `json:"FIRST_NAME"`
	LastName  string `json:"LAST_NAME"`
	Gender    string `json:"GENDER"`
	Username  string `json:"USERNAME"`
	Password  string `json:"PASSWORD"`
	Email     string `json:"EMAIL"`
	BirthDate string `json:"BIRTH_DATE"`
	Created   string `json:"CREATED"`
	Updated   string `json:"UPDATED"`
	CreatedBy int32  `json:"CREATED_BY"`
	UpdatedBy int32  `json:"UPDATED_BY"`
}

// func NewUser(w http.ResponseWriter, r *http.Request) *User {
// 	// _, err := database.Conn.Exec("INSERT INTO Temperatures (Value, InsertedAt) VALUES ($1,CURRENT_TIMESTAMP)", temperature)
// 	// if err != nil {
// 	// 	log.Println(err)
// 	// }
// 	return &User{Id: 1}
// }

func (user *User) NewUser(id int32, firstName string, lastName string, gender string, username string, password string, email string,
	birthDate string, created string, updated string, createdBy int32, updatedBy int32) *User {
	return &User{Id: id, FirstName: firstName, LastName: lastName, Gender: gender, Username: username, Password: password,
		Email: email, BirthDate: birthDate, Created: created, Updated: updated, CreatedBy: createdBy, UpdatedBy: updatedBy}
}

func (user *User) PrintUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("%", r)
}
