#include <Wire.h>
#include <SPI.h>
#include "MQTTCred.h"
#include "WiFiCred.h"

#define D0 16
#define D1 5
#define D2 4
#define D3 0
#define D4 2
#define D5 14
#define D6 12
#define D7 13
#define D8 15
#define D9 16
#define RX 3
#define TX 1
#define S3 10
#define S2 9
#define D14 16
#define D15 16
#define D16 16

#define RELAY_PIN D7
#define LED_PIN D6

#define BME_ADRESS 0x76
#define SEALEVELPRESSURE_HPA (1013.25)
#define BAUD_RATE 9600
#define DEFAULT_DELAY 2000
#define MILISECONDS 0
#define SECONDS 1
#define MINUTES 2
#define HOURS 3
