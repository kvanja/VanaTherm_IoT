package datacontroller

import (
	"iot-mqtt-api/db"
	"iot-mqtt-api/logger"
	"iot-mqtt-api/user"
	"log"
	"net/http"
)

type Datacontroller struct {
	database *db.Database
	log      *logger.Logger
	User     user.User
}

func NewDataController(d *db.Database, l *logger.Logger) *Datacontroller {
	return &Datacontroller{database: d, log: l}
}

func (dc *Datacontroller) WebServer() {
	// available URL-s for actions to be executed
	http.HandleFunc("/printUser", dc.User.PrintUser)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
