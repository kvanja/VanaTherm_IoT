package mqtt

import (
	"fmt"
	"iot-mqtt-api/config"

	mqttClient "github.com/eclipse/paho.mqtt.golang"
)

// MQTT handles connection to broker
type MQTT struct {
	client   mqttClient.Client
	Messages chan mqttClient.Message
	Config   config.MQTT
}

type MQTTPathToDevice struct {
	UserName     string `json:"username"`
	BuildingName string `json:"buildingName"`
	FloorId      int    `json:"floorId"`
	RoomId       int    `json:"roomId"`
	DeviceId     uint   `json:"deviceId"`
	RelayOn      string `json:"relayOn"`
}

// NewMQTT creates new MQTT value
func NewMQTT(c config.MQTT) *MQTT {
	m := &MQTT{
		Config:   c,
		Messages: make(chan mqttClient.Message),
	}

	opts := mqttClient.NewClientOptions()
	opts.AddBroker(mqttBrokerURL(c))
	opts.SetClientID(c.ClientID)
	opts.SetUsername(c.User)
	opts.SetPassword(c.Password)
	opts.SetDefaultPublishHandler(m.handleMessage)
	opts.AutoReconnect = true

	m.client = mqttClient.NewClient(opts)
	return m
}

func (m *MQTT) Connect() error {
	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	for _, topic := range m.Config.Subscriptions {
		token := m.client.Subscribe(topic, 1, nil)
		token.Wait()
	}

	return nil
}

// Disconnect disconnects from MQTT broker
func (m *MQTT) Disconnect() {
	m.client.Disconnect(250)
}

// Publish publishes message to MQTT broker
func (m *MQTT) Publish(topic string, message string) {
	m.client.Publish(topic, 0, false, message)
}

func mqttBrokerURL(c config.MQTT) string {
	return fmt.Sprintf("tcp://%s:%s",
		c.Host,
		c.Port,
	)
}

func (m *MQTT) handleMessage(_ mqttClient.Client, message mqttClient.Message) {
	m.Messages <- message
}
