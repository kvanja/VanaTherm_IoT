/*
  Bme280.cpp - Library for Bme280 sensor.
  Created by Kristijan Vanja, February 13, 2021.
  Released into the public domain.
*/

#include "Arduino.h"
#include "VT_Bme280.h"

VT_Bme280::VT_Bme280(uint32_t baud_rate, int bme_adress) {
  status = _bme.begin(!!bme_adress ? bme_adress : BME_ADRESS);
  _bme_adress = !!bme_adress ? bme_adress : BME_ADRESS;

  is_available();
}

bool VT_Bme280::is_available() {
  if (!_bme.begin(_bme_adress)) {
    return false;
  }
  return true;
}

void VT_Bme280::print_sensor_unavailable() {
  Serial.println("Sensor offline! Please check wiring, address, sensor ID!");
  while (!is_available()) delay(200);
}

void VT_Bme280::print_sensor_values() {
  Serial.print("Temperatura: ");
  Serial.print(get_temperature());
  Serial.println("°C");

  Serial.print("Pritisak: ");

  Serial.print(get_pressure());
  Serial.println("hPa");

  Serial.print("Nadmorska visina: ");
  Serial.print(get_altitude());
  Serial.println("m");

  Serial.print("Vlažnost zraka: ");
  Serial.print(get_humidity());
  Serial.println("%");

  Serial.println();
  delay(_Read_delay);
}

float VT_Bme280::get_temperature() {
  return _Temperature = _bme.readTemperature();
}

float VT_Bme280::get_pressure() {
  return _Pressure = _bme.readPressure() / 100.0F;
}

float VT_Bme280::get_altitude() {
  return _Aprox_altitude = _bme.readAltitude(SEALEVELPRESSURE_HPA);
}

float VT_Bme280::get_humidity() {
  return _Humidity = _bme.readHumidity();
}

int VT_Bme280::set_read_delay_in_seconds(unsigned int in_seconds) {
  return _Read_delay = in_seconds * 1000;
}

int VT_Bme280::get_read_delay() {
  return _Read_delay;
}

void VT_Bme280::set_bme_adress(int adress) {
  _bme_adress = adress;
}