#!/bin/bash
# This script is used to build the project and copy the build to the remote server
supervisor_name="iboard_http_service"
filename="iboard_http_service"
directory="/root/iboard/http_service"  # 修改为您想要的目录

# Build the project
export GOOS=linux
export GOARCH=amd64  # 如果您的服务器是x86架构，使用amd64；如果是arm架构，使用arm64

# echo "Building the project..."
go build -o $filename
echo "Build completed."

# Copy the build to the remote server
echo "Copying the build to the remote server..."
sftp root@39.108.49.167 << EOF
cd $directory
put $filename __build_temp
EOF
echo "Copy completed."

# Clean up
echo "Cleaning up..."
rm $filename
echo "Clean up completed."

# Run the service on the remote server
echo "Restarting the service on the remote server..."
echo "1090119your@" | ssh root@39.108.49.167 "mkdir -p ${directory}"
echo "1090119your@" | ssh root@39.108.49.167 "mv ${directory}/__build_temp ${directory}/${filename}"
echo "1090119your@" | ssh root@39.108.49.167 "sudo -S supervisorctl restart ${supervisor_name}"
echo "Service restarted."