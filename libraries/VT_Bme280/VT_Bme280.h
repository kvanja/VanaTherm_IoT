/*
  Bme280.cpp - Library for Bme280 sensor.
  Created by Kristijan Vanja, February 13, 2021.
  Released into the public domain.
*/

#ifndef Bme280_h
#define Bme280_h

#include "Arduino.h"
#include <Adafruit_Sensor.h>
#include <Adafruit_BME280.h>
#include <VT_Mqtt.h>

class VT_Bme280 {
  public:
    VT_Bme280(uint32_t baud_rate = BAUD_RATE, int bme_adress = BME_ADRESS);
    bool is_available();
    void print_sensor_unavailable();
    void print_sensor_values();
    float get_temperature();
    float get_pressure();
    float get_altitude();
    float get_humidity();
    int set_read_delay_in_seconds(unsigned int in_seconds);
    int get_read_delay();
    void set_bme_adress(int adress);
  private:
    Adafruit_BME280 _bme;
    int _bme_adress;
    unsigned int _Read_delay;
    unsigned status;
    float _Temperature;
    float _Pressure;
    float _Aprox_altitude;
    float _Humidity;
};
#endif
