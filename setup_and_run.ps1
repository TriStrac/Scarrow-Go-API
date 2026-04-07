# setup_and_run.ps1
$ErrorActionPreference = "Stop"

Write-Host "==================================================" -ForegroundColor Cyan
Write-Host "🚀 Scarrow-Go-API 1-Click Setup & Run 🚀" -ForegroundColor Cyan
Write-Host "==================================================" -ForegroundColor Cyan

# 1. Cleanup old instances
Write-Host "`n🧹 Cleaning up old processes..." -ForegroundColor Yellow
Stop-Process -Name "scarrow-api" -Force -ErrorAction SilentlyContinue
Stop-Process -Name "cloudflared" -Force -ErrorAction SilentlyContinue

# 2. Check/Create .env file
Write-Host "`n⚙️  Checking Environment Configurations..." -ForegroundColor Cyan
if (-not (Test-Path ".env")) {
    Write-Host "   📄 .env not found. Creating a default one..." -ForegroundColor Yellow
    $envContent = @"
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=root
MYSQL_ALLOW_EMPTY_PASSWORD=yes
DB_NAME=scarrow_db
PORT=8080
GIN_MODE=debug
JWT_SECRET=supersecretjwtkey_change_in_production
"@
    Set-Content -Path ".env" -Value $envContent
    Write-Host "   ✅ Created .env with default values." -ForegroundColor Green
} else {
    Write-Host "   ✅ .env file found." -ForegroundColor Green
}

# 3. Check for Go
Write-Host "`n🐹 Checking Go Installation..." -ForegroundColor Cyan
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "   ⏳ Go is not installed. Installing via winget..." -ForegroundColor Yellow
    winget install GoLang.Go --accept-package-agreements --accept-source-agreements
    Write-Host "   ⚠️ Go was installed. Please restart this script to load Go into the environment PATH." -ForegroundColor Red
    exit
} else {
    $goVer = go version
    Write-Host "   ✅ Go is installed ($goVer)." -ForegroundColor Green
}

# 4. Check for MySQL and Setup Database
Write-Host "`n🐬 Checking Database Setup..." -ForegroundColor Cyan
$mysqlCmd = Get-Command mysql -ErrorAction SilentlyContinue
if (-not $mysqlCmd) {
    Write-Host "   ⚠️ 'mysql' CLI not found. Ensure MySQL Server is running." -ForegroundColor Yellow
    Write-Host "   We will skip auto-creating the database. The API will auto-migrate tables if the DB exists." -ForegroundColor Yellow
} else {
    # Extract credentials safely
    $dbUserMatches = Select-String -Path ".env" -Pattern "^DB_USER=(.*)$"
    $dbPassMatches = Select-String -Path ".env" -Pattern "^DB_PASSWORD=(.*)$"
    $dbNameMatches = Select-String -Path ".env" -Pattern "^DB_NAME=(.*)$"

    $dbUser = if ($dbUserMatches) { $dbUserMatches.Matches.Groups[1].Value.Trim() } else { "root" }
    $dbPass = if ($dbPassMatches) { $dbPassMatches.Matches.Groups[1].Value.Trim() } else { "" }
    $dbName = if ($dbNameMatches) { $dbNameMatches.Matches.Groups[1].Value.Trim() } else { "scarrow_db" }

    $createDbSql = "CREATE DATABASE IF NOT EXISTS \`$dbName\` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
    
    try {
        if ($dbPass -eq "") {
            # Execute without password flag
            Start-Process -FilePath "mysql" -ArgumentList "-u $dbUser", "-e `"$createDbSql`"" -NoNewWindow -Wait
        } else {
            # Execute with password flag
            Start-Process -FilePath "mysql" -ArgumentList "-u $dbUser", "-p$dbPass", "-e `"$createDbSql`"" -NoNewWindow -Wait
        }
        Write-Host "   ✅ Database '$dbName' is ready/created." -ForegroundColor Green
    } catch {
        Write-Host "   ❌ Failed to execute DB creation script. Is MySQL running?" -ForegroundColor Red
    }
}

# 5. Build the Application
Write-Host "`n🔨 Building the Go application..." -ForegroundColor Cyan
Write-Host "   Downloading dependencies..." -ForegroundColor DarkGray
go mod tidy

Write-Host "   Compiling binary..." -ForegroundColor DarkGray
go build -o scarrow-api.exe ./cmd/api

if (Test-Path "scarrow-api.exe") {
    Write-Host "   ✅ Build successful! Executable created: scarrow-api.exe" -ForegroundColor Green
} else {
    Write-Host "   ❌ Build failed! Please check the terminal output for errors." -ForegroundColor Red
    exit
}

# 6. Check for Cloudflared
Write-Host "`n☁️  Checking Cloudflared Tunnel Installation..." -ForegroundColor Cyan
$cloudflaredCmd = Get-Command cloudflared -ErrorAction SilentlyContinue

if (-not $cloudflaredCmd) {
    # Check default install path just in case it's not in PATH yet
    if (Test-Path "C:\Program Files (x86)\cloudflared\cloudflared.exe") {
        $env:Path += ";C:\Program Files (x86)\cloudflared"
        $cloudflaredCmd = $true
    }
}

if (-not $cloudflaredCmd) {
    Write-Host "   ⏳ Cloudflared is not installed. Installing via winget..." -ForegroundColor Yellow
    winget install Cloudflare.cloudflared --accept-package-agreements --accept-source-agreements
    
    # Reload path locally for this session
    $env:Path += ";C:\Program Files (x86)\cloudflared"
    
    if (-not (Get-Command cloudflared -ErrorAction SilentlyContinue)) {
         Write-Host "   ⚠️ Cloudflared installed but path wasn't updated. Please restart the script." -ForegroundColor Red
         exit
    }
}
Write-Host "   ✅ Cloudflared is installed." -ForegroundColor Green


# 7. Run the API and Tunnel
Write-Host "`n==================================================" -ForegroundColor Cyan
Write-Host "🟢 Starting Services" -ForegroundColor Green
Write-Host "==================================================" -ForegroundColor Cyan

Write-Host "   -> Starting Scarrow API Server on http://localhost:8080" -ForegroundColor Yellow
$apiProcess = Start-Process -FilePath ".\scarrow-api.exe" -PassThru -WindowStyle Minimized

Write-Host "   ⏳ Waiting for server to initialize..." -ForegroundColor DarkGray
Start-Sleep -Seconds 3

Write-Host "   -> Starting Cloudflare Tunnel..." -ForegroundColor Yellow
Write-Host "`n🌐 Your public URL will appear below:`n" -ForegroundColor Cyan

# Run cloudflared in the foreground so the user sees the generated URL
cloudflared tunnel --url http://localhost:8080

# If the user closes cloudflared with Ctrl+C, cleanup the API
Write-Host "`n🛑 Shutting down API Server..." -ForegroundColor Yellow
Stop-Process -Id $apiProcess.Id -Force -ErrorAction SilentlyContinue
Write-Host "✅ All processes stopped gracefully." -ForegroundColor Green
