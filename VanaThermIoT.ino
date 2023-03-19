#include <VT_Esp8266.h>

VT_Esp8266 esp(BAUD_RATE);
VT_Bme280 bme(9600, 0x76);
VT_house_heating heating();
VT_Mqtt mqtt;

char msg[MSG_BUFFER_SIZE];
const char* device = "dev1";
void setup() {
  esp.set_relay_pin(RELAY_PIN);
  esp.set_interval_in_sec(5);
  esp.set_led_pin(D6);
  mqtt.connect_mqtt(device);
  mqtt.sub_to_mqtt("iot/kvanja/Kuća/floor8/room/1/dev/1/relayOn");
}

void loop() {
  esp.connect_to_WiFi(SSID_NAME, PWD);
  if (bme.is_available()) {
    mqtt.mqtt_on();
    if (esp.interval_reached()) {
      bme.print_sensor_values();
      float t = bme.get_temperature();
      float p = bme.get_pressure();
      float h = bme.get_humidity();

      if (!isnan(t) && !isnan(p) && !isnan(h)) {
        snprintf(msg, 6, "%f", t);
        mqtt.push_to_mqtt("iot/kvanja/Kuća/floor8/room/1/dev/1/temp", msg);
        snprintf(msg, 8, "%f", p);
        mqtt.push_to_mqtt("iot/kvanja/Kuća/floor8/room/1/dev/1/pressure", msg);
        snprintf(msg, 6, "%f", h);
        mqtt.push_to_mqtt("iot/kvanja/Kuća/floor8/room/1/dev/1/humidity", msg);
      }
      /*if (!esp.is_led_on()) {
        for (int i = 0; i < 3; i++) {
            esp.chng_pin_mode_w_delay(D6, LOW, MILISECONDS, 200, 100, "Palim", "Gasim");
          }
        }*/
    }
  } else {
    bme.print_sensor_unavailable();
    esp.set_relay_status(false);
  }
}
