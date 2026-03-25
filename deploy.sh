#!/bin/bash

ECR="264105488314.dkr.ecr.us-east-1.amazonaws.com/electric-backend"
CLUSTER="electric-backend-cluster"
SERVICE="electric-backend-service"

echo "🔐 Login ECR..."
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $ECR

echo "🔨 Build imagen..."
docker build -t $ECR:latest .

echo "📤 Push a ECR..."
docker push $ECR:latest

echo "🚀 Deploy ECS..."
aws ecs update-service --cluster $CLUSTER --service $SERVICE --force-new-deployment --region us-east-1

echo "✅ Deploy iniciado! Espera ~2 min en ECS"
