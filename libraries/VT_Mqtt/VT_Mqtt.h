/*
  VT_Mqtt.h - Library for MQTT.
  Created by Kristijan Vanja, February 13, 2021.
  Released into the public domain.
*/

#ifndef MQTT_h
#define MQTT_h

#include "Arduino.h"
#include <ESP8266WiFi.h>
#include <PubSubClient.h>
#include <VT_includes.h>

class VT_Mqtt {
  public:
      VT_Mqtt();
      void connect_mqtt(const char* mqtt_client);
      bool mqtt_on();
      void push_to_mqtt(const char* path, const char* value);
      bool sub_to_mqtt(const char* path);

  private:
      bool _mqtt_on = false;
};
#endif