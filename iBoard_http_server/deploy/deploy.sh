#!/bin/bash

# Configuration
REMOTE_USER="root"
REMOTE_HOST="39.108.49.167"
REMOTE_DIR="/root/iboard"
APP_NAME="iboard_http_service"

echo "Starting deployment process..."

# 1. Create deployment package
echo "Creating deployment package..."
rm -rf temp
mkdir -p temp
cp -r ../migrations temp/
cp -r ../controller temp/
cp -r ../database temp/
cp -r ../middleware temp/
cp -r ../models temp/
cp -r ../router temp/
cp -r ../services temp/
cp -r ../utils temp/
cp ../go.mod temp/
cp ../go.sum temp/
cp ../main.go temp/
cp docker-compose.yml temp/
cp ../.env temp/
cp ../Dockerfile temp/

# 2. Create deployment archive
echo "Creating deployment archive..."
cd temp
tar -czf ../deploy.tar.gz *
cd ..

# 3. Upload to server
echo "Uploading to server..."
ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST "mkdir -p $REMOTE_DIR"
scp -o StrictHostKeyChecking=no deploy.tar.gz $REMOTE_USER@$REMOTE_HOST:$REMOTE_DIR/

# 4. Deploy on server
echo "Deploying on server..."
ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST "cd $REMOTE_DIR && \
    tar -xzf deploy.tar.gz && \
    rm deploy.tar.gz && \
    docker-compose down && \
    docker-compose up -d --build"

# 5. Run database migrations
echo "Running database migrations..."
ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST "cd $REMOTE_DIR && \
    docker exec iboard_mysql mysql -u root -p1090119your iboard_db < migrations/*.sql"

# 6. Clean up
echo "Cleaning up..."
rm -rf temp
rm -f deploy.tar.gz

echo "Deployment completed!"

# 7. Show service status
echo "Service status:"
ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST "cd $REMOTE_DIR && docker-compose ps"

# 8. Show logs
echo "Recent logs:"
ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST "cd $REMOTE_DIR && docker-compose logs --tail=50 backend" 