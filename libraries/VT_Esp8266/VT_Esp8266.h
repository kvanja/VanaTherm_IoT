/*
  Bme280.cpp - Library for House heating.
  Created by Kristijan Vanja, February 13, 2021.
  Released into the public domain.
*/

#ifndef Esp8266_h
#define Esp8266_h

#include <VT_Bme280.h>
#include <VT_house_heating.h>

class VT_Esp8266 {
  public:
    VT_Esp8266(int baud_rate = 9600);

    void set_is_heating_controller(bool is_heating_controller);
    void set_relay_pin(uint8_t pin);
    void set_relay_status(bool turn_on);
    bool get_relay_status();
    void set_pin_mode_high(uint8_t pin);
    void set_pin_mode_low(uint8_t pin);
    void set_pin_mode_output(uint8_t pin);
    void set_pin_mode_input(uint8_t pin);
    void begin_serial(int rate);
    void print_msg(char* message);
    void convert_and_make_delay(uint16_t delay_time, uint8_t measurement);
    void chng_pin_mode_w_delay(uint8_t pin, int current_state, uint8_t delay_measurement, uint16_t delay_1,
                              uint16_t delay_2, char* message_1 = "", char* message_2 = "");
    void set_interval_in_min(uint16_t interval = 0);
    void set_interval_in_sec(uint16_t interval = 0);
    void set_interval_in_ms(uint16_t interval = 0);
    bool interval_reached();
    void set_led_pin(uint8_t pin);
    void set_led_status(bool status);
    bool is_led_on();
    void connect_to_WiFi(const char* SSID, const char* password);

  private:
    bool _controller_on = false;
    bool _is_heating_controller = false;
    int _relay_pin;
    bool _relay_on = false;
    bool _led_on = false;
    int _led_pin;
    unsigned long _previous_Millis = 0;
    unsigned long _current_Millis = 0;
    unsigned int _interval = 0;
};
#endif
