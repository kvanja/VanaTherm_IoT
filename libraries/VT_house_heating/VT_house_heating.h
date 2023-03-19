/*
  Bme280.cpp - Library for House heating.
  Created by Kristijan Vanja, February 13, 2021.
  Released into the public domain.
*/

#ifndef house_heating_h
#define house_heating_h

#include "Arduino.h"

class VT_house_heating {
  public:
    VT_house_heating();

    void set_heating_on(uint8_t pin = NULL);
    void set_heating_off(uint8_t pin = NULL);
    bool is_heating_on();
    void set_desired_temp(float temperature);
    void turn_on_heating_on_min_temp(float temperature);

  private:
    bool _heating_on = false;
    bool _heating_intervals_on = false;
    float _desired_temperature;
    float _minimum_temperature;
};
#endif
