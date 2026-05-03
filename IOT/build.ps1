@echo off
REM build.sh - Build pi-go-service on the Pi (handles CGO properly)
REM Usage: from Windows: bash build.sh

set PI_HOST=192.168.50.216
set PI_USER=tristrac
set PI_PASS=OuroKronii314-
set PI_DIR=/home/tristrac/scarrow

echo Building pi-go-service on Pi (CGO-enabled)...

REM Copy source files to Pi
echo Copying source to Pi...
sshpass -p %PI_PASS% scp -o StrictHostKeyChecking=no -r pi-go-service\* %PI_USER%@%PI_HOST%:%PI_DIR%/src/

REM Build on Pi
echo Building on Pi (this may take a few minutes)...
sshpass -p %PI_PASS% ssh -o StrictHostKeyChecking=no %PI_USER%@%PI_HOST% "cd %PI_DIR%/src && go build -o ../scarrow-hub ."

REM Make executable
sshpass -p %PI_PASS% ssh -o StrictHostKeyChecking=no %PI_USER%@%PI_HOST% "chmod +x %PI_DIR%/scarrow-hub"

echo Build complete!
sshpass -p %PI_PASS% ssh -o StrictHostKeyChecking=no %PI_USER%@%PI_HOST% "ls -lh %PI_DIR%/scarrow-hub"