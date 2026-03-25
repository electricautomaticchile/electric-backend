$ECR = "264105488314.dkr.ecr.us-east-1.amazonaws.com/electric-backend"
$CLUSTER = "electric-backend-cluster"
$SERVICE = "electric-backend-service"

Write-Host "Login ECR..."
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $ECR

Write-Host "Build imagen..."
docker build -t "${ECR}:latest" .

Write-Host "Push a ECR..."
docker push "${ECR}:latest"

Write-Host "Deploy ECS..."
aws ecs update-service --cluster $CLUSTER --service $SERVICE --force-new-deployment --region us-east-1

Write-Host "Deploy iniciado! Espera 2 min en ECS"
