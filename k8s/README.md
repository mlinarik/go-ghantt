# Kubernetes Deployment for go-ghantt

This directory contains Kubernetes manifests for deploying the go-ghantt application.

## Files

- `namespace.yaml` - Creates the go-ghantt namespace
- `pvc.yaml` - PersistentVolumeClaim for data storage
- `deployment.yaml` - Deployment with container configuration
- `service.yaml` - ClusterIP service
- `ingress.yaml` - Ingress for external access

## Deployment

### Quick Deploy

Apply all manifests in order:

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/pvc.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/ingress.yaml
```

Or apply all at once:

```bash
kubectl apply -f k8s/
```

### Verify Deployment

```bash
# Check namespace
kubectl get namespace go-ghantt

# Check all resources in namespace
kubectl get all -n go-ghantt

# Check pod logs
kubectl logs -n go-ghantt -l app=go-ghantt

# Check ingress
kubectl get ingress -n go-ghantt
```

### Access the Application

The application will be available at: https://go-ghantt.mlinarik.com

## Configuration

### Image

The deployment uses: `harbor.mlinarik.com/mlinarik/go-ghantt:1.01`

To update to a new version:

```bash
kubectl set image deployment/go-ghantt go-ghantt=harbor.mlinarik.com/mlinarik/go-ghantt:1.02 -n go-ghantt
```

### Storage

The PVC uses `local-path` storage class by default. Adjust `storageClassName` in `pvc.yaml` if using a different storage provider.

### Resources

Default resource limits:
- Memory: 64Mi-256Mi
- CPU: 100m-500m

Adjust in `deployment.yaml` as needed.

### TLS/SSL

The ingress is configured for:
- cert-manager with Let's Encrypt
- Automatic SSL redirect
- Host: go-ghantt.mlinarik.com

Update the host and annotations in `ingress.yaml` for your environment.

## Cleanup

To remove the deployment:

```bash
kubectl delete -f k8s/
```

Or delete the namespace (removes everything):

```bash
kubectl delete namespace go-ghantt
```
