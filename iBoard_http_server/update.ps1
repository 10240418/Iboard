# Configuration
$REMOTE_USER = "root"
$REMOTE_HOST = "39.108.49.167"
$REMOTE_DIR = "/root/iboard"
$REMOTE_PASS = "1090119your@"
$DB_HOST = "localhost"
$DB_PORT = "3308"
$DB_USER = "root"
$DB_PASS = "1090119your"
$DB_NAME = "iboard_db"

Write-Host "Starting update process..."

# Function to handle errors
function Handle-Error {
  param($message)
  Write-Host "Error: $message" -ForegroundColor Red
  Write-Host "Attempting to rollback..."
    
  # Rollback command
  $rollbackCmd = @"
cd $REMOTE_DIR && 
cp -r backup/controller/base/* controller/base/ &&
cp -r backup/models/base/* models/base/ &&
cp -r backup/services/base/* services/base/ &&
docker-compose down && 
docker-compose build backend && 
docker-compose up -d
"@
  echo $REMOTE_PASS | ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST $rollbackCmd
    
  Write-Host "Rollback completed. Please check service status."
  exit 1
}

# 1. Check service status first
Write-Host "Checking current service status..."
$cmd = "cd $REMOTE_DIR && docker-compose ps"
$result = echo $REMOTE_PASS | ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST $cmd
if (-not $?) { Handle-Error "Failed to check service status" }

# 2. Create backup of current files
Write-Host "Creating backup of current files..."
$cmd = @"
cd $REMOTE_DIR && 
mkdir -p backup/controller/base &&
mkdir -p backup/models/base &&
mkdir -p backup/services/base &&
cp -r controller/base/building_controller.go backup/controller/base/ &&
cp -r models/base/building.go backup/models/base/ &&
cp -r services/base/building_service.go backup/services/base/ &&
echo 'Backup created'
"@
$result = echo $REMOTE_PASS | ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST $cmd
if (-not $?) { Handle-Error "Failed to create backup" }

# 3. Stop backend service gracefully
Write-Host "Stopping backend service..."
$cmd = "cd $REMOTE_DIR && docker-compose stop backend && sleep 5"
$result = echo $REMOTE_PASS | ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST $cmd
if (-not $?) { Handle-Error "Failed to stop backend service" }

# 3.5 Clean up misplaced files
Write-Host "Cleaning up misplaced files..."
$cmd = @"
cd $REMOTE_DIR &&
rm -f services/base/building_controller.go &&
rm -f services/base/router.go &&
echo 'Cleaned up misplaced files'
"@
$result = echo $REMOTE_PASS | ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST $cmd
if (-not $?) { Handle-Error "Failed to clean up misplaced files" }

# 4. Create temp directory and subdirectories
if (Test-Path "temp") {
  Remove-Item -Recurse -Force "temp"
}
New-Item -ItemType Directory -Path "temp" -Force
New-Item -ItemType Directory -Path "temp/controller" -Force
New-Item -ItemType Directory -Path "temp/controller/base" -Force
New-Item -ItemType Directory -Path "temp/models" -Force
New-Item -ItemType Directory -Path "temp/models/base" -Force
New-Item -ItemType Directory -Path "temp/services" -Force
New-Item -ItemType Directory -Path "temp/services/base" -Force

Write-Host "Creating update package..."

# 5. Copy modified files
Write-Host "Copying modified files..."
Copy-Item "controller/base/building_controller.go" -Destination "temp/controller/base/" -Force
Copy-Item "models/base/building.go" -Destination "temp/models/base/" -Force
Copy-Item "services/base/building_service.go" -Destination "temp/services/base/" -Force

# 6. Create update package
Write-Host "Creating update archive..."
Set-Location "temp"
tar -czf "../update.tar.gz" *
Set-Location ..

# 7. Upload files
Write-Host "Uploading update package..."
$result = echo $REMOTE_PASS | scp -o StrictHostKeyChecking=no "update.tar.gz" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/"
if (-not $?) { Handle-Error "Failed to upload update package" }

# 8. Extract files and verify
Write-Host "Extracting files and verifying..."
$cmd = @"
cd $REMOTE_DIR && 
tar -xzf update.tar.gz && 
rm update.tar.gz && 
ls -l controller/base/building_controller.go models/base/building.go services/base/building_service.go
"@
$result = echo $REMOTE_PASS | ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST $cmd
if (-not $?) { Handle-Error "Failed to extract files" }

# 9. Build and start backend service
Write-Host "Building and starting backend service..."
$cmd = @"
cd $REMOTE_DIR && 
docker-compose down && 
docker-compose build backend && 
docker-compose up -d
"@
$result = echo $REMOTE_PASS | ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST $cmd
if (-not $?) { Handle-Error "Failed to build and start service" }

# 10. Wait for service to start
Write-Host "Waiting for service to start..."
Start-Sleep -Seconds 10

# 11. Verify service status
Write-Host "Verifying service status..."
$cmd = @"
cd $REMOTE_DIR && 
docker-compose ps && 
echo 'Checking service health...' && 
curl -s http://localhost:10031/health || echo 'Service not responding'
"@
$result = echo $REMOTE_PASS | ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST $cmd
if ($result -match "Service not responding") { Handle-Error "Service failed to start properly" }

# 12. Clean up
Write-Host "Cleaning up temporary files..."
Remove-Item -Recurse -Force "temp"

Write-Host "Update completed successfully!"

# 13. Check final status
Write-Host ""
Write-Host "Checking container status..."
$cmd = "cd $REMOTE_DIR && docker-compose ps"
echo $REMOTE_PASS | ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST $cmd

Write-Host ""
Write-Host "Important notes:"
Write-Host "1. Backend service updated with new building management features"
Write-Host "2. Service port remains: 10031"
Write-Host "3. Using existing MySQL and Redis services"
Write-Host "4. Backup of original files created in $REMOTE_DIR/backup"
Write-Host ""
Write-Host "To view logs use: docker-compose logs -f backend"
Write-Host "To restart service use: docker-compose restart backend"
Write-Host "To rollback use: cd $REMOTE_DIR && cp -r backup/controller/base/* controller/base/ && cp -r backup/models/base/* models/base/ && cp -r backup/services/base/* services/base/ && docker-compose down && docker-compose build backend && docker-compose up -d"
Write-Host ""

Pause 