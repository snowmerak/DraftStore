# DraftStore

DraftStore is a cloud-native file upload service that provides a two-stage upload mechanism using draft and main buckets. It's designed for scenarios where you need to upload files temporarily before confirming their final placement.

## 🏗️ Architecture

### System Architecture

```mermaid
graph TB
    subgraph "Client Layer"
        WebClient[Web Client]
        gRPCClient[gRPC Client]
        MobileApp[Mobile App]
    end
    
    subgraph "Kubernetes Cluster"
        subgraph "Load Balancer"
            LB[Load Balancer Service]
        end
        
        subgraph "Application Pods"
            Server1[Server Pod 1]
            Server2[Server Pod 2]
            Server3[Server Pod 3]
        end
        
        subgraph "Background Jobs"
            CronJob[Cleanup CronJob]
        end
        
        subgraph "Configuration"
            ConfigMap[ConfigMap]
            Secret[Secret]
        end
    end
    
    subgraph "AWS S3"
        DraftBucket[Draft Bucket]
        MainBucket[Main Bucket]
    end
    
    WebClient --> LB
    gRPCClient --> LB
    MobileApp --> LB
    
    LB --> Server1
    LB --> Server2
    LB --> Server3
    
    Server1 --> DraftBucket
    Server1 --> MainBucket
    Server2 --> DraftBucket
    Server2 --> MainBucket
    Server3 --> DraftBucket
    Server3 --> MainBucket
    
    CronJob --> DraftBucket
    
    ConfigMap --> Server1
    ConfigMap --> Server2
    ConfigMap --> Server3
    ConfigMap --> CronJob
    
    Secret --> Server1
    Secret --> Server2
    Secret --> Server3
    Secret --> CronJob
```

### Component Architecture

```mermaid
graph TB
    subgraph "API Layer"
        gRPCController[gRPC Controller]
        WebAPIController[Web API Controller]
    end
    
    subgraph "Service Layer"
        DraftService[Draft Service]
        CleanerService[Cleaner Service]
    end
    
    subgraph "Storage Layer"
        StorageInterface[Storage Interface]
        S3Implementation[S3 Implementation]
    end
    
    subgraph "External"
        S3Bucket[AWS S3 Buckets]
    end
    
    gRPCController --> DraftService
    WebAPIController --> DraftService
    DraftService --> StorageInterface
    CleanerService --> StorageInterface
    StorageInterface --> S3Implementation
    S3Implementation --> S3Bucket
```

### Use Case Diagram

```mermaid
graph LR
    subgraph "Actors"
        User[User]
        System[System/Cron]
    end
    
    subgraph "Use Cases"
        UC1[Get Upload URL]
        UC2[Upload File to Draft]
        UC3[Get Download URL]
        UC4[Download Draft File]
        UC5[Confirm Upload]
        UC6[Move to Main Bucket]
        UC7[Cleanup Expired Files]
        UC8[Health Check]
    end
    
    User --> UC1
    User --> UC2
    User --> UC3
    User --> UC4
    User --> UC5
    UC1 --> UC2
    UC2 --> UC3
    UC3 --> UC4
    UC2 --> UC5
    UC5 --> UC6
    System --> UC7
    System --> UC8
```

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
├── manifest/               # Kubernetes manifests
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

## 📊 Workflow Diagrams

### Complete Upload Workflow

```mermaid
sequenceDiagram
    participant Client
    participant DraftStore
    participant DraftBucket as Draft S3 Bucket
    participant MainBucket as Main S3 Bucket
    
    Note over Client, MainBucket: Two-Stage Upload Process
    
    Client->>DraftStore: 1. Request Upload URL
    Note right of Client: POST /api/v1/upload-url<br/>{"object_name": "file.jpg"}
    
    DraftStore->>DraftBucket: 2. Generate Presigned URL
    DraftBucket-->>DraftStore: 3. Return Presigned URL
    DraftStore-->>Client: 4. Return Upload URL + Metadata
    Note left of DraftStore: {"upload_url": "https://...",<br/>"object_name": "file.jpg",<br/>"expires_at": "2024-01-01T12:00:00Z"}
    
    Client->>DraftBucket: 5. Upload File Directly
    Note right of Client: PUT to presigned URL<br/>with file data
    DraftBucket-->>Client: 6. Upload Success
    
    Client->>DraftStore: 7. Request Download URL (Optional)
    Note right of Client: POST /api/v1/download-url<br/>{"object_name": "file.jpg"}
    
    DraftStore->>DraftBucket: 8. Generate Download URL
    DraftBucket-->>DraftStore: 9. Return Download URL
    DraftStore-->>Client: 10. Return Download URL
    
    Client->>DraftBucket: 11. Download/Preview File (Optional)
    DraftBucket-->>Client: 12. File Data
    
    Client->>DraftStore: 13. Confirm Upload
    Note right of Client: POST /api/v1/confirm-upload<br/>{"object_name": "file.jpg"}
    
    DraftStore->>DraftBucket: 14. Copy Object
    DraftStore->>MainBucket: 15. Move to Main Bucket
    DraftStore->>DraftBucket: 16. Delete from Draft
    DraftStore-->>Client: 17. Confirmation Success
    Note left of DraftStore: {"success": true,<br/>"final_location": "main/file.jpg"}
```

### gRPC Workflow

```mermaid
sequenceDiagram
    participant gRPCClient as gRPC Client
    participant gRPCServer as gRPC Server
    participant DraftService as Draft Service
    participant StorageLayer as Storage Layer
    participant S3 as AWS S3
    
    gRPCClient->>gRPCServer: CreateDraftBucket()
    gRPCServer->>DraftService: CreateBucket()
    DraftService->>StorageLayer: CreateBucket()
    StorageLayer->>S3: CreateBucket API
    S3-->>StorageLayer: Success
    StorageLayer-->>DraftService: Success
    DraftService-->>gRPCServer: BucketCreated
    gRPCServer-->>gRPCClient: CreateDraftBucketResponse
    
    gRPCClient->>gRPCServer: GetUploadURL(object_name)
    gRPCServer->>DraftService: GenerateUploadURL()
    DraftService->>StorageLayer: GetPresignedURL()
    StorageLayer->>S3: Generate Presigned URL
    S3-->>StorageLayer: Presigned URL
    StorageLayer-->>DraftService: URL + Metadata
    DraftService-->>gRPCServer: UploadURLResponse
    gRPCServer-->>gRPCClient: GetUploadURLResponse
    
    Note over gRPCClient, S3: Direct upload to S3 (bypassing server)
    
    gRPCClient->>gRPCServer: ConfirmUpload(object_name)
    gRPCServer->>DraftService: ConfirmUpload()
    DraftService->>StorageLayer: MoveObject()
    StorageLayer->>S3: Copy + Delete Operations
    S3-->>StorageLayer: Success
    StorageLayer-->>DraftService: Success
    DraftService-->>gRPCServer: UploadConfirmed
    gRPCServer-->>gRPCClient: ConfirmUploadResponse
```

### Cleanup Process

```mermaid
sequenceDiagram
    participant CronJob as Cleanup CronJob
    participant CleanerService as Cleaner Service
    participant StorageLayer as Storage Layer
    participant DraftBucket as Draft S3 Bucket
    
    Note over CronJob, DraftBucket: Daily Cleanup Process (2 AM UTC)
    
    CronJob->>CleanerService: Start Cleanup Process
    CleanerService->>StorageLayer: ListExpiredObjects()
    StorageLayer->>DraftBucket: List Objects with Metadata
    DraftBucket-->>StorageLayer: Object List + Timestamps
    
    loop For Each Expired Object
        StorageLayer-->>CleanerService: Expired Object Info
        CleanerService->>StorageLayer: DeleteObject(object_name)
        StorageLayer->>DraftBucket: Delete Object
        DraftBucket-->>StorageLayer: Delete Success
        StorageLayer-->>CleanerService: Deletion Confirmed
    end
    
    CleanerService-->>CronJob: Cleanup Complete
    Note right of CleanerService: Log: "Cleaned up N expired objects"
```

### Error Handling Flow

```mermaid
flowchart TD
    A[Client Request] --> B{Valid Request?}
    B -->|No| C[Return 400 Bad Request]
    B -->|Yes| D{AWS Credentials Valid?}
    D -->|No| E[Return 500 Internal Error]
    D -->|Yes| F{S3 Bucket Exists?}
    F -->|No| G[Return 404 Not Found]
    F -->|Yes| H{Object Exists? (for download/confirm)}
    H -->|No| I[Return 404 Object Not Found]
    H -->|Yes| J[Process Request]
    J --> K{Operation Success?}
    K -->|No| L[Return 500 Internal Error]
    K -->|Yes| M[Return Success Response]
    
    C --> N[Log Error]
    E --> N
    G --> N
    I --> N
    L --> N
    M --> O[Log Success]
```

### Deployment Flow

```mermaid
flowchart TB
    subgraph "Development"
        A[Code Changes]
        B[Build Docker Images]
        C[Push to Registry]
    end
    
    subgraph "Kubernetes Cluster"
        D[Apply Manifests]
        E[Rolling Update]
        F[Health Checks]
        G[Service Ready]
    end
    
    subgraph "Monitoring"
        H[Pod Status]
        I[Resource Usage]
        J[Application Logs]
        K[CronJob Status]
    end
    
    A --> B
    B --> C
    C --> D
    D --> E
    E --> F
    F --> G
    
    G --> H
    G --> I
    G --> J
    G --> K
    
    style A fill:#e1f5fe
    style G fill:#c8e6c9
    style H fill:#fff3e0
    style I fill:#fff3e0
    style J fill:#fff3e0
    style K fill:#fff3e0
```

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
