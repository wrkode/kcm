# Adopted Cluster Lifecycle Management

## Overview

KCM now supports full lifecycle management of adopted clusters, enabling you to manage existing Kubernetes clusters as if they were created by KCM. This feature provides infrastructure discovery, CAPI integration, scaling, upgrades, monitoring, and backup capabilities for adopted clusters.

## Features

### 1. Infrastructure Discovery

KCM can automatically discover and map existing infrastructure to CAPI resources:

- **Auto Discovery**: Automatically discovers nodes and infrastructure resources
- **Manual Discovery**: Manually specify node mappings for precise control
- **Provider Support**: Supports AWS, Azure, GCP, vSphere, and OpenStack

### 2. CAPI Integration

Generates and manages Cluster API (CAPI) resources for adopted clusters:

- **Cluster Resources**: Creates CAPI Cluster resources
- **Machine Resources**: Maps existing nodes to CAPI Machine resources
- **Infrastructure Resources**: Generates provider-specific infrastructure resources

### 3. Scaling Operations

Automatic and manual scaling of adopted clusters:

- **Auto Scaling**: Scale based on metrics (CPU, memory, etc.)
- **Node Groups**: Define different node groups with specific configurations
- **Min/Max Constraints**: Set scaling boundaries
- **Rolling Updates**: Safe scaling with rolling updates

### 4. Upgrade Management

Perform cluster and node upgrades:

- **Rolling Upgrades**: Upgrade nodes one by one to maintain availability
- **In-Place Upgrades**: Upgrade entire cluster at once
- **Maintenance Windows**: Schedule upgrades during maintenance windows
- **Version Management**: Track current and target versions

### 5. Monitoring

Comprehensive monitoring and health checks:

- **Health Checks**: Monitor cluster health status
- **Metrics Collection**: Collect CPU, memory, disk, and custom metrics
- **Status Reporting**: Real-time status updates
- **Alerting**: Integration with monitoring systems

### 6. Backup/Recovery

Automated backup and restore operations:

- **Scheduled Backups**: Configure backup schedules
- **Full/Incremental**: Support for full and incremental backups
- **Multiple Storage**: Support for S3, GCS, Azure, and local storage
- **Retention Policies**: Configurable backup retention

## Configuration

### Basic Configuration

Enable adopted cluster management by adding the `adoptedCluster` section to your ClusterDeployment:

```yaml
apiVersion: k0rdent.mirantis.com/v1beta1
kind: ClusterDeployment
metadata:
  name: my-adopted-cluster
spec:
  template: adopted-cluster
  adoptedCluster:
    enabled: true
    discoveryMode: "auto"  # or "manual"
    infrastructureMapping:
      provider: "aws"
      region: "us-west-2"
      credentials:
        secretName: "aws-credentials"
        namespace: "default"
    lifecycleManagement:
      enabled: true
      # ... additional configuration
```

### Discovery Configuration

#### Auto Discovery

For automatic infrastructure discovery:

```yaml
adoptedCluster:
  discoveryMode: "auto"
  infrastructureMapping:
    provider: "aws"
    region: "us-west-2"
    credentials:
      secretName: "aws-credentials"
      namespace: "default"
```

#### Manual Discovery

For manual node mapping:

```yaml
adoptedCluster:
  discoveryMode: "manual"
  infrastructureMapping:
    provider: "aws"
    region: "us-west-2"
    credentials:
      secretName: "aws-credentials"
      namespace: "default"
    nodeMapping:
      - nodeName: "worker-node-1"
        machineName: "worker-machine-1"
        infrastructureRef:
          apiVersion: "infrastructure.cluster.x-k8s.io/v1beta1"
          kind: "AWSMachine"
          name: "worker-machine-1"
```

### Scaling Configuration

Configure auto-scaling behavior:

```yaml
lifecycleManagement:
  scaling:
    enabled: true
    autoScaling: true
    minNodes: 2
    maxNodes: 10
    nodeGroups:
      - name: "worker-nodes"
        instanceType: "t3.medium"
        minNodes: 2
        maxNodes: 8
        labels:
          node-role.kubernetes.io/worker: "true"
        taints: []
```

### Upgrade Configuration

Configure upgrade behavior:

```yaml
lifecycleManagement:
  upgrades:
    enabled: true
    autoUpgrade: false
    maintenanceWindow: "0 2 * * 0"  # Sundays at 2 AM
    upgradeStrategy: "rolling"  # or "in-place"
    maxUnavailable: 1
```

### Monitoring Configuration

Configure monitoring and health checks:

```yaml
lifecycleManagement:
  monitoring:
    enabled: true
    metrics:
      - "cpu_usage_percent"
      - "memory_usage_percent"
      - "disk_usage_percent"
      - "node_count"
      - "pod_count"
    healthChecks:
      enabled: true
      interval: "30s"
      timeout: "10s"
      failureThreshold: 3
```

### Backup Configuration

Configure backup and restore operations:

```yaml
lifecycleManagement:
  backup:
    enabled: true
    schedule: "0 1 * * *"  # Daily at 1 AM
    retention: "30d"
    storage:
      type: "s3"  # or "gcs", "azure", "local"
      location: "s3://my-backup-bucket/cluster-backups/"
      credentials:
        secretName: "backup-credentials"
        namespace: "default"
```

## Status and Monitoring

### Status Fields

The ClusterDeployment status includes detailed information about adopted cluster management:

```yaml
status:
  adoptedClusterStatus:
    discoveryStatus:
      phase: "Completed"
      message: "Infrastructure discovery completed successfully"
      lastDiscoveryTime: "2024-01-15T10:30:00Z"
      discoveredResources:
        - type: "node"
          name: "worker-node-1"
          provider: "aws"
          region: "us-west-2"
          status: "Ready"
    capiStatus:
      phase: "Completed"
      message: "CAPI integration completed successfully"
      lastSyncTime: "2024-01-15T10:35:00Z"
      generatedResources:
        - apiVersion: "cluster.x-k8s.io/v1beta1"
          kind: "Cluster"
          name: "my-adopted-cluster"
          namespace: "default"
          status: "Ready"
    scalingStatus:
      phase: "Completed"
      currentNodes: 3
      desiredNodes: 3
      scalingOperations:
        - type: "scale-up"
          timestamp: "2024-01-15T09:00:00Z"
          fromNodes: 2
          toNodes: 3
          status: "Completed"
    upgradeStatus:
      phase: "NotStarted"
      currentVersion: "1.28.0"
      targetVersion: "1.29.0"
    monitoringStatus:
      phase: "Completed"
      healthStatus: "Healthy"
      lastHealthCheckTime: "2024-01-15T10:40:00Z"
      metrics:
        - name: "cpu_usage_percent"
          value: "75.5"
          unit: "%"
          timestamp: "2024-01-15T10:40:00Z"
    backupStatus:
      phase: "Completed"
      lastBackupTime: "2024-01-15T01:00:00Z"
      backupOperations:
        - type: "full"
          timestamp: "2024-01-15T01:00:00Z"
          status: "Completed"
          size: "2.5GB"
          location: "s3://my-backup-bucket/cluster-backups/"
```

## Examples

### AWS Adopted Cluster

```yaml
apiVersion: k0rdent.mirantis.com/v1beta1
kind: ClusterDeployment
metadata:
  name: aws-adopted-cluster
spec:
  template: adopted-cluster
  adoptedCluster:
    enabled: true
    discoveryMode: "auto"
    infrastructureMapping:
      provider: "aws"
      region: "us-west-2"
      credentials:
        secretName: "aws-credentials"
        namespace: "default"
    lifecycleManagement:
      enabled: true
      scaling:
        enabled: true
        autoScaling: true
        minNodes: 2
        maxNodes: 10
        nodeGroups:
          - name: "worker-nodes"
            instanceType: "t3.medium"
            minNodes: 2
            maxNodes: 8
      upgrades:
        enabled: true
        autoUpgrade: false
        upgradeStrategy: "rolling"
      monitoring:
        enabled: true
        metrics:
          - "cpu_usage_percent"
          - "memory_usage_percent"
      backup:
        enabled: true
        schedule: "0 1 * * *"
        storage:
          type: "s3"
          location: "s3://my-backup-bucket/"
```

### Azure Adopted Cluster

```yaml
apiVersion: k0rdent.mirantis.com/v1beta1
kind: ClusterDeployment
metadata:
  name: azure-adopted-cluster
spec:
  template: adopted-cluster
  adoptedCluster:
    enabled: true
    discoveryMode: "manual"
    infrastructureMapping:
      provider: "azure"
      region: "eastus"
      credentials:
        secretName: "azure-credentials"
        namespace: "default"
      nodeMapping:
        - nodeName: "aks-agentpool-12345678-vmss000000"
          machineName: "worker-machine-1"
    lifecycleManagement:
      enabled: true
      scaling:
        enabled: true
        nodeGroups:
          - name: "worker-pool"
            instanceType: "Standard_D2s_v3"
            minNodes: 3
            maxNodes: 12
      backup:
        enabled: true
        storage:
          type: "azure"
          location: "https://storageaccount.blob.core.windows.net/backup-container/"
```

## Prerequisites

### Credentials

Ensure you have the necessary credentials configured:

1. **Cloud Provider Credentials**: Create secrets with cloud provider credentials
2. **Backup Storage Credentials**: Configure backup storage credentials
3. **Cluster Access**: Ensure KCM can access the adopted cluster

### Example Credential Secrets

#### AWS Credentials

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws-credentials
  namespace: default
type: Opaque
data:
  aws_access_key_id: <base64-encoded-access-key>
  aws_secret_access_key: <base64-encoded-secret-key>
```

#### Azure Credentials

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: azure-credentials
  namespace: default
type: Opaque
data:
  azure_subscription_id: <base64-encoded-subscription-id>
  azure_client_id: <base64-encoded-client-id>
  azure_client_secret: <base64-encoded-client-secret>
  azure_tenant_id: <base64-encoded-tenant-id>
```

## Troubleshooting

### Common Issues

1. **Discovery Failures**: Check cloud provider credentials and permissions
2. **Scaling Issues**: Verify node group configurations and instance types
3. **Upgrade Failures**: Ensure maintenance windows are properly configured
4. **Backup Failures**: Verify backup storage credentials and permissions

### Status Checks

Monitor the status of adopted cluster operations:

```bash
# Check ClusterDeployment status
kubectl get clusterdeployment my-adopted-cluster -o yaml

# Check adopted cluster status
kubectl get clusterdeployment my-adopted-cluster -o jsonpath='{.status.adoptedClusterStatus}'
```

### Logs

Check controller logs for detailed information:

```bash
# Check adopted cluster controller logs
kubectl logs -n kcm-system deployment/kcm-controller -c manager
```

## Migration Guide

### From Basic Adoption

If you have existing adopted clusters without lifecycle management:

1. **Enable Lifecycle Management**: Add the `lifecycleManagement` section
2. **Configure Discovery**: Set up infrastructure mapping
3. **Enable Features Gradually**: Start with monitoring, then add scaling and upgrades
4. **Test Backup**: Configure and test backup functionality

### Best Practices

1. **Start Small**: Begin with monitoring only, then add other features
2. **Test in Staging**: Test configurations in a staging environment first
3. **Monitor Status**: Regularly check the status of adopted cluster operations
4. **Backup Before Changes**: Always backup before major changes
5. **Use Maintenance Windows**: Schedule upgrades during maintenance windows

## API Reference

### AdoptedClusterConfig

```yaml
adoptedCluster:
  enabled: boolean                    # Enable adopted cluster management
  discoveryMode: string              # "auto" or "manual"
  infrastructureMapping:              # Infrastructure mapping configuration
    provider: string                 # Cloud provider
    region: string                   # Cloud region
    credentials:                     # Credentials configuration
      secretName: string
      namespace: string
    nodeMapping:                     # Manual node mapping
      - nodeName: string
        machineName: string
        infrastructureRef:           # Infrastructure reference
          apiVersion: string
          kind: string
          name: string
  lifecycleManagement:               # Lifecycle management configuration
    enabled: boolean
    scaling:                         # Scaling configuration
      enabled: boolean
      autoScaling: boolean
      minNodes: integer
      maxNodes: integer
      nodeGroups:                    # Node group configuration
        - name: string
          instanceType: string
          minNodes: integer
          maxNodes: integer
          labels: map[string]string
          taints:                    # Node taints
            - key: string
              value: string
              effect: string
    upgrades:                        # Upgrade configuration
      enabled: boolean
      autoUpgrade: boolean
      maintenanceWindow: string
      upgradeStrategy: string        # "rolling" or "in-place"
      maxUnavailable: integer
    monitoring:                      # Monitoring configuration
      enabled: boolean
      metrics: []string
      healthChecks:                  # Health check configuration
        enabled: boolean
        interval: string
        timeout: string
        failureThreshold: integer
    backup:                          # Backup configuration
      enabled: boolean
      schedule: string
      retention: string
      storage:                       # Storage configuration
        type: string                 # "s3", "gcs", "azure", "local"
        location: string
        credentials:                 # Storage credentials
          secretName: string
          namespace: string
``` 