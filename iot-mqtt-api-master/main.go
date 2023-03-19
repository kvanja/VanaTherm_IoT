package main

import (
	"iot-mqtt-api/config"
	"iot-mqtt-api/db"
	"iot-mqtt-api/iot"
	"iot-mqtt-api/logger"
	"iot-mqtt-api/mqtt"
	"log"
)

func main() {
	c, err := config.Load("configuration.toml")
	if err != nil {
		log.Fatal(err)
	}

	l, err := logger.New(c.Log.Filename)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	d, err := db.Connect(c.Database)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	defer d.Disconnect()

	m := mqtt.NewMQTT(c.MQTT)
	err = m.Connect()
	if err != nil {
		log.Fatal("Unable to connect to mqtt:", err)
	}
	defer m.Disconnect()

	i := iot.NewIOT(d, m, l, c.OwApiURL.URL)
	go i.WebServer()
	i.ReceiveData()

}
