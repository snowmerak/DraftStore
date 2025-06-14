# DraftStore Kubernetes Manifests

This directory contains Kubernetes manifest files for deploying DraftStore to a Kubernetes cluster.

## Prerequisites

- Kubernetes cluster (v1.19+)
- kubectl configured to access your cluster
- Docker images built and available:
  - `draftstore-server:latest`
  - `draftstore-cronjob:latest`

## Quick Deployment

1. **Update AWS credentials in secret.yaml**
   ```bash
   # Encode your AWS credentials
   echo -n "your-access-key-id" | base64
   echo -n "your-secret-access-key" | base64
   
   # Update the values in secret.yaml
   ```

2. **Update configuration in configmap.yaml**
   - Set your S3 bucket name
   - Adjust region and other settings as needed

3. **Deploy using kubectl**
   ```bash
   # Deploy all resources
   kubectl apply -f .
   
   # Or deploy in order
   kubectl apply -f namespace.yaml
   kubectl apply -f configmap.yaml
   kubectl apply -f secret.yaml
   kubectl apply -f deployment.yaml
   kubectl apply -f service.yaml
   kubectl apply -f cronjob.yaml
   kubectl apply -f hpa.yaml
   ```

4. **Deploy using Kustomize**
   ```bash
   kubectl apply -k .
   ```

## Verification

```bash
# Check all resources
kubectl get all -n draftstore

# Check pods
kubectl get pods -n draftstore

# Check services
kubectl get services -n draftstore

# Check cronjobs
kubectl get cronjobs -n draftstore

# Check HPA
kubectl get hpa -n draftstore
```

## Accessing the Service

```bash
# Get external IP (for LoadBalancer)
kubectl get service draftstore-service -n draftstore

# Port forward for testing
kubectl port-forward service/draftstore-service 8080:80 -n draftstore
kubectl port-forward service/draftstore-service 50051:50051 -n draftstore
```

## Monitoring

```bash
# Check logs
kubectl logs -f deployment/draftstore-server -n draftstore

# Check cronjob logs
kubectl logs -f job/draftstore-cleanup-<job-id> -n draftstore

# Monitor resource usage
kubectl top pods -n draftstore
```

## Cleanup

```bash
# Remove all resources
kubectl delete -f .

# Or remove namespace (this removes everything)
kubectl delete namespace draftstore
```

## Customization

- Modify resource limits in `deployment.yaml` based on your needs
- Adjust HPA settings in `hpa.yaml` for your scaling requirements
- Change cronjob schedule in `cronjob.yaml` as needed
- Update image tags in `kustomization.yaml` for different versions
