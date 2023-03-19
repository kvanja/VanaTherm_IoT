/*
  Bme280.cpp - Library for House heating.
  Created by Kristijan Vanja, February 13, 2021.
  Released into the public domain.
*/

#include "Arduino.h"
#include "VT_house_heating.h"

VT_house_heating::VT_house_heating() {}

void VT_house_heating::set_heating_on(uint8_t pin) {
    _heating_on = true;
}

void VT_house_heating::set_heating_off(uint8_t pin) {
    _heating_on = false;
}

bool VT_house_heating::is_heating_on() {
    return _heating_on;
}

void VT_house_heating::set_desired_temp(float temperature) {
    _desired_temperature = temperature;
}

void VT_house_heating::turn_on_heating_on_min_temp(float temperature) {
    _minimum_temperature = temperature;
}