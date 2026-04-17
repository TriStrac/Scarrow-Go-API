// --- Pin Definitions ---
#define RADAR_RX 16 
#define RADAR_TX 17 

// --- LD2420 Protocol Constants ---
uint8_t openCfg[]  = {0xFD, 0xFC, 0xFB, 0xFA, 0x04, 0x00, 0xFF, 0x00, 0x01, 0x00, 0x04, 0x03, 0x02, 0x01};
// 0x64 is "Running/Normal Mode" (ASCII/Simple output)
uint8_t setRunMode[] = {0xFD, 0xFC, 0xFB, 0xFA, 0x08, 0x00, 0x12, 0x00, 0x00, 0x00, 0x64, 0x00, 0x00, 0x00, 0x04, 0x03, 0x02, 0x01};
uint8_t closeCfg[] = {0xFD, 0xFC, 0xFB, 0xFA, 0x02, 0x00, 0xFE, 0x00, 0x04, 0x03, 0x02, 0x01};

// --- State Tracking ---
int lastDistanceCm = -1;
bool isMotionDetected = false;

void clearRX() {
  while(Serial1.available()) {
    Serial1.read();
  }
}

bool sendCommand(const char* label, uint8_t* cmd, int len) {
  clearRX(); 
  
  Serial.print("[CMD] "); Serial.print(label); Serial.print("... ");
  Serial1.write(cmd, len);
  Serial1.flush();
  
  unsigned long start = millis();
  uint8_t resp[64];
  int idx = 0;
  
  while (millis() - start < 300 && idx < 64) {
    if (Serial1.available()) {
      resp[idx++] = Serial1.read();
    }
  }

  if (idx > 0) {
    Serial.print("ACK: ");
    for(int i=0; i<idx; i++) {
      if(resp[i] < 0x10) Serial.print("0");
      Serial.print(resp[i], HEX); Serial.print(" ");
    }
    Serial.println();
    return true; 
  } else {
    Serial.println("TIMEOUT!");
    return false;
  }
}

void setup() {
  Serial.begin(115200);
  while(!Serial); 

  Serial1.begin(115200, SERIAL_8N1, RADAR_RX, RADAR_TX);
  
  delay(2000); 
  Serial.println("\n\n=========================================");
  Serial.println("🚀 Scarrow Node v2.6 - DISTANCE TRACKER");
  Serial.println("=========================================");
  
  sendCommand("Open Config", openCfg, sizeof(openCfg));
  delay(100);
  
  sendCommand("Set Normal Mode", setRunMode, sizeof(setRunMode));
  delay(100);

  sendCommand("Close Config", closeCfg, sizeof(closeCfg));
  
  Serial.println("\n--- Listening for ASCII Motion Events ---");
}

void loop() {
  if (Serial1.available()) {
    String data = Serial1.readStringUntil('\n');
    data.trim(); // Remove \r
    
    if (data.length() > 0) {
      if (data == "ON") {
        if (!isMotionDetected) {
           Serial.println("🚨 MOTION DETECTED! (Target Entered Zone)");
           isMotionDetected = true;
           // TODO: Turn on Horn
        }
      } 
      else if (data == "OFF") {
        if (isMotionDetected) {
           Serial.println("✅ CLEAR (Target Left Zone).");
           isMotionDetected = false;
           lastDistanceCm = -1; // Reset distance tracking
           // TODO: Turn off Horn
        }
      } 
      else if (data.startsWith("Range")) {
        int spaceIdx = data.indexOf(' ');
        if (spaceIdx > 0) {
          int distance_mm = data.substring(spaceIdx + 1).toInt();
          int distance_cm = distance_mm / 10;
          
          // Only log if distance changes by more than 2cm (to prevent spam from micro-jitter)
          if (abs(distance_cm - lastDistanceCm) > 2) {
             Serial.print("🎯 Target Distance updated: "); 
             Serial.print(distance_cm); 
             Serial.println(" cm");
             lastDistanceCm = distance_cm;
          }
        }
      }
    }
  }
}
