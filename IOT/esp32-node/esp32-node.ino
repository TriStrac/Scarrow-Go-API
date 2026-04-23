#include <Preferences.h>

HardwareSerial UART1(1);

#define RADAR1_OT1 16

uint8_t openCfg[]  = {0xFD, 0xFC, 0xFB, 0xFA, 0x04, 0x00, 0xFF, 0x00, 0x01, 0x00, 0x04, 0x03, 0x02, 0x01};
uint8_t setRunMode[] = {0xFD, 0xFC, 0xFB, 0xFA, 0x08, 0x00, 0x12, 0x00, 0x00, 0x00, 0x64, 0x00, 0x00, 0x00, 0x04, 0x03, 0x02, 0x01};
uint8_t closeCfg[] = {0xFD, 0xFC, 0xFB, 0xFA, 0x02, 0x00, 0xFE, 0x00, 0x04, 0x03, 0x02, 0x01};

Preferences preferences;
String nodeId = "";
String nodeSecret = "";
String hubFilter = "";
bool isMotionDetected = false;
int lastDistanceCm = -1;

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

void parseDistance(String& data, int* lastDistanceCm, const char* label) {
  if (data.startsWith("Range")) {
    int idx = data.indexOf(' ');
    if (idx > 0) {
      *lastDistanceCm = data.substring(idx + 1).toInt();
      Serial.print(label); Serial.print(*lastDistanceCm); Serial.println(" cm");
    }
  }
}

void setup() {
  disableCore1WDT();

  Serial.begin(115200);
  Serial.println("[DEBUG] 1: Serial done");

  preferences.begin("scarrow", true);
  nodeId = preferences.getString("node_id", "");
  nodeSecret = preferences.getString("node_secret", "");
  hubFilter = preferences.getString("hub_filter", "");
  preferences.end();
  Serial.println("[DEBUG] 2: Preferences done");

  nodeId = "test-node";
  nodeSecret = "test-secret";
  hubFilter = "test-hub";
  Serial.println("[DEBUG] 3: forced credentials");

  Serial.println("[DEBUG] 4: past setup mode check");
  pinMode(RADAR1_OT1, INPUT);
  Serial.println("[DEBUG] 5: pinMode done");

  UART1.begin(115200, SERIAL_8N1, 18, 19);
  Serial.println("[DEBUG] 6: UART1 begin done");

  delay(2000);
  Serial.println("\n🚀 Scarrow Node - RADAR 1 ONLY TEST");

  sendCmd("Open", openCfg, sizeof(openCfg));
  delay(100);
  sendCmd("Mode", setRunMode, sizeof(setRunMode));
  delay(100);
  sendCmd("Close", closeCfg, sizeof(closeCfg));

  Serial.println("\n--- Radar 1 Active ---");
}

void loop() {
  int r1 = digitalRead(RADAR1_OT1);

  static int lastR1 = -1;

  if (r1 != lastR1) {
    Serial.print("[R1 OT1] "); Serial.println(r1 ? "HIGH" : "LOW");
    lastR1 = r1;
  }

  if (r1 == HIGH) {
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

  if (UART1.available()) {
    String data = UART1.readStringUntil('\n');
    data.trim();
    if (data.length() > 0) {
      Serial.print("[R1] "); Serial.println(data);
      parseDistance(data, &lastDistanceCm, "Distance: ");
    }
  }
}