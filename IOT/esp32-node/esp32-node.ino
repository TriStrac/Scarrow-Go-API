#include <BLEDevice.h>
#include <BLEUtils.h>
#include <BLEServer.h>
#include <BLE2902.h>
#include <Preferences.h>
#include <ArduinoJson.h>
// --- HardwareSerial ---
HardwareSerial UART1(1);
HardwareSerial UART2(2);

// --- Pin Definitions ---
// Radar 1 (Bird detection - top)
#define RADAR1_OT1 16
#define RADAR1_TX 17

// Radar 2 (Rat detection - bottom)
#define RADAR2_OT1 13
#define RADAR2_RX 12
#define RADAR2_TX 10

// Deterrence (not yet connected)
// #define MOSFET_PIN 4
// #define AUDIO_PIN 5
// #define PEST_FREQ 20000

// --- BLE UUIDs ---
#define SERVICE_UUID        "4fafc201-1fb5-459e-8fcc-c5c9c331914b"
#define CHARACTERISTIC_UUID "beb5483e-36e1-4688-b7f5-ea07361b26a8"

// --- LD2420 Config ---
uint8_t openCfg[]  = {0xFD, 0xFC, 0xFB, 0xFA, 0x04, 0x00, 0xFF, 0x00, 0x01, 0x00, 0x04, 0x03, 0x02, 0x01};
uint8_t setRunMode[] = {0xFD, 0xFC, 0xFB, 0xFA, 0x08, 0x00, 0x12, 0x00, 0x00, 0x00, 0x64, 0x00, 0x00, 0x00, 0x04, 0x03, 0x02, 0x01};
uint8_t closeCfg[] = {0xFD, 0xFC, 0xFB, 0xFA, 0x02, 0x00, 0xFE, 0x00, 0x04, 0x03, 0x02, 0x01};

// --- Global ---
Preferences preferences;
String nodeId = "";
String nodeSecret = "";
String hubFilter = "";
bool isMotionDetected = false;
int lastDistanceCm = -1;
int lastDistance2Cm = -1;
bool isSetupMode = false;

// --- BLE Callbacks ---
class MyCallbacks: public BLECharacteristicCallbacks {
    void onWrite(BLECharacteristic *pCharacteristic) {
      String value = pCharacteristic->getValue();
      if (value.length() > 0) {
        Serial.println("Received Provisioning Data...");
        JsonDocument doc;
        DeserializationError error = deserializeJson(doc, value);
        if (error) { Serial.println(error.f_str()); return; }
        const char* nId = doc["node_id"];
        const char* nSecret = doc["node_secret"];
        const char* hFilter = doc["hub_filter"];
        if (nId && nSecret && hFilter) {
          preferences.begin("scarrow", false);
          preferences.putString("node_id", nId);
          preferences.putString("node_secret", nSecret);
          preferences.putString("hub_filter", hFilter);
          preferences.end();
          Serial.println("Config Saved! Restarting...");
          delay(1000);
          ESP.restart();
        }
      }
    }
};

void enterSetupMode() {
  isSetupMode = true;
  Serial.println("Entering SETUP MODE...");
  BLEDevice::init("Scarrow_Node_Setup");
  BLEDevice::setMTU(512);
  BLEServer *pServer = BLEDevice::createServer();
  BLEService *pService = pServer->createService(SERVICE_UUID);
  BLECharacteristic *pCharacteristic = pService->createCharacteristic(
    CHARACTERISTIC_UUID,
    BLECharacteristic::PROPERTY_READ | BLECharacteristic::PROPERTY_WRITE
  );
  pCharacteristic->setCallbacks(new MyCallbacks());
  pService->start();
  BLEAdvertising *pAdvertising = BLEDevice::getAdvertising();
  pAdvertising->addServiceUUID(SERVICE_UUID);
  BLEDevice::startAdvertising();
  Serial.println("Advertising 'Scarrow_Node_Setup'...");
}

void clearSerial() {
  while (UART1.available()) UART1.read();
}

bool sendCmd(const char* label, uint8_t* cmd, int len) {
  clearSerial();
  Serial.print("["); Serial.print(label); Serial.print("] ");
  UART1.write(cmd, len);
  UART1.flush();
  unsigned long start = millis();
  uint8_t resp[64];
  int idx = 0;
  while (millis() - start < 300 && idx < 64) {
    if (UART1.available()) resp[idx++] = UART1.read();
  }
  if (idx > 0) { Serial.println("OK"); return true; }
  else { Serial.println("TIMEOUT"); return false; }
}

void clearSerial2() {
  while (UART2.available()) UART2.read();
}

bool sendCmd2(const char* label, uint8_t* cmd, int len) {
  clearSerial2();
  Serial.print("[R2 "); Serial.print(label); Serial.print("] ");
  UART2.write(cmd, len);
  UART2.flush();
  unsigned long start = millis();
  uint8_t resp[64];
  int idx = 0;
  while (millis() - start < 300 && idx < 64) {
    if (UART2.available()) resp[idx++] = UART2.read();
  }
  if (idx > 0) { Serial.println(" OK"); return true; }
  else { Serial.println(" TIMEOUT"); return false; }
}

void parseDistance(String& data, int* lastDistanceCm, const char* label) {
  if (data.startsWith("Range")) {
    int idx = data.indexOf(' ');
    if (idx > 0) {
      *lastDistanceCm = data.substring(idx + 1).toInt();
      Serial.print(label); Serial.print(*lastDistanceCm); Serial.println(" cm");
    }
  }
}

void processRadar2() {
  if (UART2.available()) {
    String data = UART2.readStringUntil('\n');
    data.trim();
    if (data.length() > 0) {
      Serial.print("[R2] "); Serial.println(data);
      parseDistance(data, &lastDistance2Cm, "Distance2: ");
    }
  }
}

void setup() {
  Serial.begin(115200);
  
  preferences.begin("scarrow", true);
  nodeId = preferences.getString("node_id", "");
  nodeSecret = preferences.getString("node_secret", "");
  hubFilter = preferences.getString("hub_filter", "");
  preferences.end();

  if (nodeId == "" || nodeSecret == "") { enterSetupMode(); return; }

  Serial.println("Node ID: " + nodeId);
  pinMode(RADAR1_OT1, INPUT);
  pinMode(RADAR2_OT1, INPUT);

  UART1.begin(115200, SERIAL_8N1, 17, 16);
  delay(200);
  UART2.begin(115200, SERIAL_8N1, RADAR2_RX, RADAR2_TX);
  delay(200);
  
  delay(2000);
  Serial.println("\n🚀 Scarrow Node - FIELD MODE");
  
  sendCmd("Open", openCfg, sizeof(openCfg));
  delay(100);
  sendCmd("Mode", setRunMode, sizeof(setRunMode));
  delay(100);
  sendCmd("Close", closeCfg, sizeof(closeCfg));
  
  sendCmd2("Open", openCfg, sizeof(openCfg));
  delay(100);
  sendCmd2("Mode", setRunMode, sizeof(setRunMode));
  delay(100);
  sendCmd2("Close", closeCfg, sizeof(closeCfg));
  Serial.println("Radar 2 config sent");
  
  Serial.println("\n--- Radars Active ---");
}

void loop() {
  if (isSetupMode) { delay(1000); return; }

  // Read OT1 for motion detection
  int r1 = digitalRead(RADAR1_OT1);
  int r2 = digitalRead(RADAR2_OT1);
  
  static int lastR1 = -1, lastR2 = -1;
  
  if (r1 != lastR1) {
    Serial.print("[R1 OT1] "); Serial.println(r1 ? "HIGH" : "LOW");
    lastR1 = r1;
  }
  if (r2 != lastR2) {
    Serial.print("[R2 OT1] "); Serial.println(r2 ? "HIGH" : "LOW");
    lastR2 = r2;
  }
  
  // Motion detection state (deterrence not yet connected)
  if (r1 == HIGH || r2 == HIGH) {
    if (!isMotionDetected) {
      Serial.println("MOTION DETECTED");
      isMotionDetected = true;
    }
  } else {
    if (isMotionDetected) {
      Serial.println("CLEAR");
      isMotionDetected = false;
    }
  }
  
  // Read serial distance from Radar 1
  if (UART1.available()) {
    String data = UART1.readStringUntil('\n');
    data.trim();
    if (data.length() > 0) {
      Serial.print("[R1] "); Serial.println(data);
      parseDistance(data, &lastDistanceCm, "Distance: ");
    }
  }
  
  // Process Radar 2 serial data
  processRadar2();
}