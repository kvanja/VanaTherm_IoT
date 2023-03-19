/*
  VT_Mqtt.cpp - Library for MQTT.
  Created by Kristijan Vanja, February 13, 2021.
  Released into the public domain.
*/

#include "VT_Mqtt.h"
WiFiClient _espClient;
PubSubClient _client(_espClient);

void callback(char* topic, byte* message, unsigned int length) {
  Serial.print("Message arrived on topic: ");
  Serial.print(topic);
  Serial.print(". Message: ");
  String messageTemp;
  
  for (int i = 0; i < length; i++) {
    Serial.print((char)message[i]);
    messageTemp += (char)message[i];
  }
  if (messageTemp == "true") {
    digitalWrite(D7, LOW);
  } else {
    digitalWrite(D7, HIGH);
  }
}

VT_Mqtt::VT_Mqtt() {
    _client.setServer(MQTT_SERVER, 1883);
}

void VT_Mqtt::connect_mqtt(const char* mqtt_client) {
    while (!_client.connected()) {
        Serial.print("Attempting MQTT connection...");
        Serial.println(mqtt_client);
        Serial.println(MQTT_USERNAME);
        Serial.println(MQTT_PASSW);
        if (_client.connect(mqtt_client, MQTT_USERNAME, MQTT_PASSW)){
            Serial.println("Connected.");
            _mqtt_on = true;
        } else {
            Serial.println("Reconnecting...");
            delay(DEFAULT_DELAY);
        }
    }
}

bool VT_Mqtt::mqtt_on() {
    _client.loop();
    return _mqtt_on;
}

void VT_Mqtt::push_to_mqtt(const char* path, const char* value) {
    Serial.print("Publishing...\n");
    _client.publish(path, value);
}

bool VT_Mqtt::sub_to_mqtt(const char* path) {
    _client.subscribe(path);
    _client.setCallback(callback);
    return true;
}