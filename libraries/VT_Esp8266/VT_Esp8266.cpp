/*
  Bme280.cpp - Library for House heating.
  Created by Kristijan Vanja, February 13, 2021.
  Released into the public domain.
*/

#include "Arduino.h"
#include "VT_Esp8266.h"

VT_Esp8266::VT_Esp8266(int baud_rate) {
  begin_serial(!!baud_rate ? baud_rate : BAUD_RATE);
}

void VT_Esp8266::set_is_heating_controller(bool is_heating_controller) {
  _is_heating_controller = is_heating_controller;
}

void VT_Esp8266::set_relay_pin(uint8_t pin) {
  _relay_pin = pin;
  set_pin_mode_output(pin);
}

void VT_Esp8266::set_relay_status(bool turn_on) {
  _relay_on = turn_on;
  turn_on ? set_pin_mode_high(_relay_pin) : set_pin_mode_low(_relay_pin);
}

bool VT_Esp8266::get_relay_status() {
  return _relay_on;
}

void VT_Esp8266::set_pin_mode_high(uint8_t pin) {
  digitalWrite(pin, HIGH);
}

void VT_Esp8266::set_pin_mode_low(uint8_t pin) {
  digitalWrite(pin, LOW);
}

void VT_Esp8266::set_pin_mode_output(uint8_t pin) {
  pinMode(pin, OUTPUT);
}

void VT_Esp8266::set_pin_mode_input(uint8_t pin) {
  pinMode(pin, INPUT);
}

void VT_Esp8266::begin_serial(int rate) {
  Serial.begin(rate);
  if (!Serial) {
    while (!Serial) {
      print_msg("Microcontroller offline, please check your microcontroller!");
    } 
  } else {
    _controller_on = true;
  }
}

void VT_Esp8266::print_msg(char* message) {
  if (message != "") {
    Serial.println(message);
  }
}

void VT_Esp8266::convert_and_make_delay(uint16_t delay_time, uint8_t measurement) {
  if (measurement == SECONDS) {
    delay(delay_time * 1000);
  } else if (measurement == MINUTES) {
    delay(delay_time * 60 * 1000);
  } else if (measurement == HOURS) {
    delay(delay_time * 60 * 60 * 1000);
  } else {
    delay(delay_time);
  }
}

void VT_Esp8266::chng_pin_mode_w_delay(uint8_t pin, int current_state, uint8_t delay_measurement, uint16_t delay_1,
                              uint16_t delay_2, char* message_1, char* message_2) {
  if (current_state == 0) {
    print_msg(message_1);
    set_pin_mode_high(pin);
    convert_and_make_delay(delay_1, delay_measurement);
    print_msg(message_2);
    set_pin_mode_low(pin);
    convert_and_make_delay(delay_2, delay_measurement);
  } else {
    print_msg(message_1);
    set_pin_mode_high(pin);
    convert_and_make_delay(delay_1, delay_measurement);
    print_msg(message_2);
    set_pin_mode_low(pin);
    convert_and_make_delay(delay_2, delay_measurement);
  }
}

void VT_Esp8266::set_interval_in_min(uint16_t interval) {
  _interval = interval ? interval * (60 * 1000) : 0;
}

void VT_Esp8266::set_interval_in_sec(uint16_t interval) {
  _interval = interval ? interval * 1000 : 0;
}

void VT_Esp8266::set_interval_in_ms(uint16_t interval) {
  _interval = interval ? interval : 0;
}

bool VT_Esp8266::interval_reached() {
  _current_Millis = millis();
  if (_current_Millis - _previous_Millis >= _interval) {
    _previous_Millis = _current_Millis;
    if (_interval == 0) {
      _interval = DEFAULT_DELAY;
    }
    return true;
  }
  return false;
}

void VT_Esp8266::set_led_pin(uint8_t pin) {
  _led_pin = pin;
  set_pin_mode_output(_led_pin);
}

void VT_Esp8266::set_led_status(bool status) {
  if (!!_led_pin) {
    if (status) {
      set_pin_mode_high(_led_pin);
    } else {
      set_pin_mode_low(_led_pin);
    }
  }
  _led_on = !!_led_pin ? status : false;
}

bool VT_Esp8266::is_led_on() {
  return _led_on;
}

void VT_Esp8266::connect_to_WiFi(const char* SSID, const char* password) {
  if (WiFi.status() != WL_CONNECTED) {
    if (SSID != "" && password != "") {
      Serial.println("Connecting to WiFi");
      Serial.println(SSID_NAME);
      WiFi.mode(WIFI_STA);
      WiFi.begin(SSID_NAME, PWD);

      Serial.println();
      Serial.print("Connecting...");
      while (WiFi.status() != WL_CONNECTED) {
        delay(1000);
        Serial.println("Trying to connect, please wait...");
      }

      Serial.println("WiFi connected!");
      Serial.print("NodeMcu IP adress:");
      Serial.println(WiFi.localIP());

    } else {
      Serial.println("No WiFi credentials!");
    }
  }
}
