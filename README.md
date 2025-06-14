# DraftStore

DraftStore is a cloud-native file upload service that provides a two-stage upload mechanism using draft and main buckets. It's designed for scenarios where you need to upload files temporarily before confirming their final placement.

## 🏗️ Architecture

### Project Structure

```
DraftStore/
├── proto/                     # Protocol Buffer definitions
│   └── draft/v1/
│       └── draft.proto       # gRPC service definitions
├── gen/                      # Generated code
├── lib/                      # Core library components
│   ├── storage/             # Storage abstraction layer
│   │   └── s3/             # AWS S3 implementation
│   ├── service/            # Business logic services
│   │   ├── draft/          # Draft upload service
│   │   └── cleaner/        # Cleanup service
│   └── controller/         # API controllers
│       ├── grpc/           # gRPC server implementation
│       └── webapi/         # REST API implementation
├── cmd/                     # Application entry points
│   ├── server/             # Main server (gRPC + HTTP)
│   └── cronjob/            # Cleanup cronjob
├── buf.yaml                # Buf configuration
├── buf.gen.yaml            # Code generation configuration
└── go.mod                  # Go module dependencies
```

### Core Components

#### 1. Storage Layer (`lib/storage/`)
- **Interface**: Abstract storage interface for cloud provider flexibility
- **S3 Implementation**: AWS S3-specific implementation with presigned URLs
- **Operations**: Bucket management, object operations, cleanup

#### 2. Services (`lib/service/`)
- **Draft Service**: Manages two-stage upload workflow
- **Cleaner Service**: Handles automatic cleanup of expired draft objects

#### 3. API Layer (`lib/controller/`)
- **gRPC Server**: High-performance binary protocol
- **Web API Server**: RESTful HTTP API for web clients

#### 4. Applications (`cmd/`)
- **Server**: Combined gRPC and HTTP server
- **Cronjob**: Kubernetes Job for periodic cleanup

## 🚀 Features

- **Two-Stage Upload**: Upload to draft bucket, then confirm to move to main bucket
- **Presigned URLs**: Secure direct-to-storage uploads without proxying files
- **Automatic Cleanup**: Configurable cleanup of expired draft objects
- **Dual APIs**: Both gRPC and REST APIs available
- **Cloud Native**: Designed for Kubernetes deployment
- **Storage Agnostic**: Interface-based design allows different storage backends

## 📋 Prerequisites

- Go 1.24.4+
- AWS Account with S3 access
- Docker (for containerized deployment)
- Kubernetes cluster (for production deployment)
- Buf CLI (for protocol buffer generation)

## 🛠️ Installation

### Local Development

1. **Clone the repository**
   ```bash
   git clone https://github.com/snowmerak/DraftStore.git
   cd DraftStore
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Generate code from protobuf**
   ```bash
   go tool buf generate
   ```

4. **Set up AWS credentials**
   ```bash
   export AWS_ACCESS_KEY_ID=your-access-key
   export AWS_SECRET_ACCESS_KEY=your-secret-key
   export AWS_REGION=us-east-1
   ```

5. **Configure environment variables**
   ```bash
   export BUCKET_NAME=your-main-bucket
   export GRPC_PORT=50051
   export HTTP_PORT=8080
   export UPLOAD_TTL=3600    # 1 hour
   export DOWNLOAD_TTL=3600  # 1 hour
   ```

6. **Run the server**
   ```bash
   go run cmd/server/main.go
   ```

### Building Binaries

```bash
# Build server
go build -o bin/server cmd/server/main.go

# Build cronjob
go build -o bin/cronjob cmd/cronjob/main.go
```

## 🐳 Docker Deployment

### Build Docker Images

```dockerfile
# Dockerfile.server
FROM golang:1.24.4-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
CMD ["./server"]
```

```dockerfile
# Dockerfile.cronjob
FROM golang:1.24.4-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o cronjob cmd/cronjob/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/cronjob .
CMD ["./cronjob"]
```

### Build and Push

```bash
# Build images
docker build -f Dockerfile.server -t your-registry/draftstore-server:latest .
docker build -f Dockerfile.cronjob -t your-registry/draftstore-cronjob:latest .

# Push to registry
docker push your-registry/draftstore-server:latest
docker push your-registry/draftstore-cronjob:latest
```

## ☸️ Kubernetes Deployment

### 1. ConfigMap for Configuration

```yaml
# config/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: draftstore-config
  namespace: draftstore
data:
  BUCKET_NAME: "main"
  AWS_REGION: "us-east-1"
  GRPC_PORT: "50051"
  HTTP_PORT: "8080"
  UPLOAD_TTL: "3600"
  DOWNLOAD_TTL: "3600"
  OBJECT_LIFETIME: "86400"  # 24 hours
```

### 2. Secret for AWS Credentials

```yaml
# config/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws-credentials
  namespace: draftstore
type: Opaque
data:
  AWS_ACCESS_KEY_ID: <base64-encoded-access-key>
  AWS_SECRET_ACCESS_KEY: <base64-encoded-secret-key>
```

### 3. Deployment for Server

```yaml
# deploy/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: draftstore-server
  namespace: draftstore
spec:
  replicas: 3
  selector:
    matchLabels:
      app: draftstore-server
  template:
    metadata:
      labels:
        app: draftstore-server
    spec:
      containers:
      - name: server
        image: your-registry/draftstore-server:latest
        ports:
        - containerPort: 50051
          name: grpc
        - containerPort: 8080
          name: http
        envFrom:
        - configMapRef:
            name: draftstore-config
        - secretRef:
            name: aws-credentials
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### 4. Service for Load Balancing

```yaml
# deploy/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: draftstore-service
  namespace: draftstore
spec:
  selector:
    app: draftstore-server
  ports:
  - name: grpc
    port: 50051
    targetPort: 50051
  - name: http
    port: 80
    targetPort: 8080
  type: LoadBalancer
```

### 5. CronJob for Cleanup

```yaml
# deploy/cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: draftstore-cleanup
  namespace: draftstore
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: cleanup
            image: your-registry/draftstore-cronjob:latest
            envFrom:
            - configMapRef:
                name: draftstore-config
            - secretRef:
                name: aws-credentials
            resources:
              requests:
                memory: "64Mi"
                cpu: "50m"
              limits:
                memory: "256Mi"
                cpu: "200m"
          restartPolicy: OnFailure
```

### 6. Deploy to Kubernetes

```bash
# Create namespace
kubectl create namespace draftstore

# Apply configurations
kubectl apply -f config/
kubectl apply -f deploy/

# Check deployment status
kubectl get pods -n draftstore
kubectl get services -n draftstore
kubectl get cronjobs -n draftstore
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `BUCKET_NAME` | Main S3 bucket name | `main` | ✅ |
| `AWS_REGION` | AWS region | `us-east-1` | ✅ |
| `AWS_ACCESS_KEY_ID` | AWS access key | - | ✅ |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key | - | ✅ |
| `GRPC_PORT` | gRPC server port | `50051` | ❌ |
| `HTTP_PORT` | HTTP server port | `8080` | ❌ |
| `UPLOAD_TTL` | Upload URL TTL (seconds) | `3600` | ❌ |
| `DOWNLOAD_TTL` | Download URL TTL (seconds) | `3600` | ❌ |
| `OBJECT_LIFETIME` | Draft object lifetime (seconds) | `86400` | ❌ |

## 📊 Expected Behavior in Kubernetes

### Normal Operations

1. **Server Pods**: 3 replicas running for high availability
2. **Load Balancer**: Distributes traffic across server instances
3. **Health Checks**: Automatic pod restarts on failures
4. **Cleanup Job**: Runs daily to remove expired draft objects

### Scaling

```bash
# Scale up for high traffic
kubectl scale deployment draftstore-server --replicas=5 -n draftstore

# Horizontal Pod Autoscaler
kubectl autoscale deployment draftstore-server --cpu-percent=50 --min=3 --max=10 -n draftstore
```

### Monitoring

```bash
# Check logs
kubectl logs -f deployment/draftstore-server -n draftstore

# Check cronjob status
kubectl get jobs -n draftstore

# Monitor resource usage
kubectl top pods -n draftstore
```

## 🔌 API Usage

### gRPC API

```protobuf
service DraftService {
  rpc CreateDraftBucket(CreateDraftBucketRequest) returns (CreateDraftBucketResponse);
  rpc GetUploadURL(GetUploadURLRequest) returns (GetUploadURLResponse);
  rpc GetDownloadURL(GetDownloadURLRequest) returns (GetDownloadURLResponse);
  rpc ConfirmUpload(ConfirmUploadRequest) returns (ConfirmUploadResponse);
}
```

### REST API

```bash
# Get upload URL
curl -X POST http://localhost:8080/api/v1/upload-url \
  -H "Content-Type: application/json" \
  -d '{"object_name": "my-file.jpg"}'

# Get download URL
curl -X POST http://localhost:8080/api/v1/download-url \
  -H "Content-Type: application/json" \
  -d '{"object_name": "my-file.jpg"}'

# Confirm upload
curl -X POST http://localhost:8080/api/v1/confirm-upload \
  -H "Content-Type: application/json" \
  -d '{"object_name": "my-file.jpg"}'
```

## 🔍 Troubleshooting

### Common Issues

1. **AWS Credentials**: Ensure proper IAM permissions for S3 operations
2. **Bucket Names**: Must be globally unique in S3
3. **Network Policies**: Ensure pods can reach S3 endpoints
4. **Resource Limits**: Adjust based on traffic patterns

### Debug Commands

```bash
# Check pod logs
kubectl logs -l app=draftstore-server -n draftstore

# Check cronjob history
kubectl get jobs -n draftstore

# Test connectivity
kubectl exec -it deployment/draftstore-server -n draftstore -- wget -O- http://localhost:8080/health
```

## 📝 License

[Add your license information here]

## 🤝 Contributing

[Add contribution guidelines here]
