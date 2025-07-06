# KCM Examples

This directory contains examples demonstrating how to use KCM (Kubernetes Cluster Manager) to create and manage Kubernetes clusters across different cloud providers.

## AKS (Azure Kubernetes Service) Examples

### Overview

KCM provides a declarative way to create and manage AKS clusters using Kubernetes-native resources. The workflow involves:

1. **Credentials Management**: Azure service principal credentials stored as Kubernetes secrets
2. **Template-Based Deployment**: Pre-defined Helm charts that generate the necessary CAPI (Cluster API) resources
3. **Infrastructure Provisioning**: Azure Service Operator (ASO) creates the actual AKS resources
4. **Lifecycle Management**: KCM monitors and manages the cluster lifecycle

### Examples

#### 1. Minimal AKS Cluster (`aks-cluster-minimal.yaml`)
Shows the essential components needed to create a basic AKS cluster:
- Azure credentials (Secret + Credential)
- Basic ClusterDeployment with minimal configuration
- System and user node pools

#### 2. Comprehensive AKS Cluster (`aks-cluster-example.yaml`)
Demonstrates advanced features:
- Complete AKS configuration with all options
- Service deployment after cluster creation
- IPAM (IP Address Management) configuration
- Detailed workflow explanation and usage instructions

#### 3. Adopted AKS Cluster (`aks-adopted-cluster-minimal.yaml`)
Shows how to adopt an existing AKS cluster into KCM:
- Kubeconfig secret for existing cluster access
- Credential referencing the kubeconfig
- ClusterDeployment with adopted cluster configuration
- Lifecycle management for existing infrastructure

#### 4. Comprehensive Adopted AKS Cluster (`aks-adopted-cluster-example.yaml`)
Demonstrates full adoption with advanced features:
- Complete adoption process with detailed explanations
- Infrastructure discovery and mapping
- Advanced lifecycle management (scaling, upgrades, monitoring, backup)
- Service deployment to adopted clusters

### Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   ClusterDeploy │    │   KCM Controller │    │  Azure Service  │
│   ment          │───▶│                  │───▶│   Operator      │
│                 │    │                  │    │   (ASO)         │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌──────────────────┐
                       │   CAPI Resources │
                       │                  │
                       │ • Cluster        │
                       │ • AzureASOManaged│
                       │   Cluster        │
                       │ • MachinePool   │
                       └──────────────────┘
```

### Key Components

#### 1. ClusterTemplate
- References a Helm chart that contains the AKS cluster templates
- Defines the structure and configuration options for AKS clusters
- Example: `azure-aks-1-0-1` template

#### 2. Credential
- Standardized way to reference cloud provider credentials
- Points to a Kubernetes Secret containing Azure service principal details
- Used by KCM and propagated to the cluster for CCM (Cloud Controller Manager)

#### 3. ClusterDeployment
- The main resource that triggers cluster creation
- References a ClusterTemplate and Credential
- Contains configuration that gets passed to the Helm chart
- KCM creates a HelmRelease that manages the cluster lifecycle

#### 4. Generated Resources
When a ClusterDeployment is applied, KCM generates:

- **Cluster** (`cluster.x-k8s.io/v1beta1`): The main CAPI cluster resource
- **AzureASOManagedCluster** (`infrastructure.cluster.x-k8s.io/v1alpha1`): Azure-specific cluster infrastructure
- **AzureASOManagedControlPlane** (`infrastructure.cluster.x-k8s.io/v1alpha1`): AKS control plane configuration
- **MachinePool** resources for each node pool (system and user)

### Examples

#### 1. Minimal AKS Cluster (`aks-cluster-minimal.yaml`)
Shows the essential components needed to create a basic AKS cluster:
- Azure credentials (Secret + Credential)
- Basic ClusterDeployment with minimal configuration
- System and user node pools

#### 2. Comprehensive AKS Cluster (`aks-cluster-example.yaml`)
Demonstrates advanced features:
- Complete AKS configuration with all options
- Service deployment after cluster creation
- IPAM (IP Address Management) configuration
- Detailed workflow explanation

### Workflow Steps

1. **Validation Phase**
   - KCM validates the ClusterTemplate exists and is valid
   - Validates the Credential is ready and accessible
   - Validates the configuration against the template schema

2. **Helm Release Creation**
   - Creates a HelmRelease that references the AKS chart
   - The chart contains templates for AzureASOManagedCluster, AzureASOManagedControlPlane, and MachinePool resources

3. **Infrastructure Provisioning**
   - Azure Resource Group is created
   - AKS Managed Cluster is provisioned
   - System and User node pools are created
   - Network configuration is applied

4. **CAPI Integration**
   - CAPI Cluster resource is created
   - CAPI MachinePool resources are created for each node pool
   - Infrastructure resources are linked to CAPI resources

5. **Monitoring and Status**
   - KCM monitors the cluster creation process
   - Status conditions are updated (Ready, InfrastructureReady, etc.)
   - Events are generated for important milestones

6. **Service Deployment** (if specified)
   - Services are deployed to the cluster after it's ready
   - Sveltos Profile manages the service lifecycle
   - Health checks ensure services are running correctly

### Configuration Options

#### Basic Configuration
```yaml
config:
  location: "eastus"
  kubernetes:
    version: "v1.31.1"
    networkPlugin: "azure"
    networkPolicy: "azure"
  machinePools:
    system:
      count: 3
      vmSize: "Standard_D2s_v3"
    user:
      count: 2
      vmSize: "Standard_D4s_v3"
```

#### Advanced Configuration
```yaml
config:
  # Network configuration
  clusterNetwork:
    pods:
      cidrBlocks: ["10.244.0.0/16"]
    services:
      cidrBlocks: ["10.96.0.0/12"]
  
  # API Server access
  apiServerAccessProfile:
    authorizedIPRanges: []
    enablePrivateCluster: false
  
  # Auto-upgrade settings
  autoUpgradeProfile:
    nodeOSUpgradeChannel: "None"
    upgradeChannel: "none"
  
  # Security features
  securityProfile:
    azureKeyVaultKms:
      enabled: false
    defender:
      securityMonitoring:
        enabled: false
    workloadIdentity:
      enabled: false
```

### Monitoring and Troubleshooting

#### Check Cluster Status
```bash
# Check ClusterDeployment status
kubectl get clusterdeployment my-aks-cluster -o yaml

# Check generated CAPI resources
kubectl get cluster,azureasomanagedcluster,machinepool

# Check HelmRelease status
kubectl get helmrelease my-aks-cluster
```

#### Common Status Conditions
- `TemplateReady`: ClusterTemplate validation status
- `CredentialReady`: Credential validation status
- `HelmReleaseReady`: HelmRelease deployment status
- `CAPIClusterSummary`: CAPI cluster health status
- `SveltosProfileReady`: Service deployment status

#### Troubleshooting
1. **Template Not Found**: Ensure the ClusterTemplate exists and is valid
2. **Credential Issues**: Verify Azure credentials are correct and have proper permissions
3. **Infrastructure Errors**: Check Azure subscription and resource group permissions
4. **Network Issues**: Verify network configuration and CIDR blocks don't conflict

### Best Practices

1. **Credentials Management**
   - Use dedicated service principals for each environment
   - Rotate credentials regularly
   - Use Azure Key Vault for credential storage in production

2. **Network Planning**
   - Plan CIDR blocks carefully to avoid conflicts
   - Use private clusters for production workloads
   - Configure authorized IP ranges for API server access

3. **Node Pool Design**
   - Use system node pool for system workloads only
   - Design user node pools based on workload requirements
   - Enable autoscaling for cost optimization

4. **Security**
   - Enable Azure Defender for production clusters
   - Use workload identity for secure service-to-service communication
   - Configure RBAC and network policies

5. **Monitoring**
   - Enable Azure Monitor for comprehensive monitoring
   - Set up alerts for cluster health and performance
   - Use KCM's service deployment for monitoring stacks

## Adopted Cluster Examples

### Overview

KCM can also adopt existing Kubernetes clusters (including AKS clusters) for lifecycle management. This allows you to:

1. **Adopt Existing Clusters**: Bring existing clusters under KCM management
2. **Infrastructure Discovery**: Automatically discover existing infrastructure
3. **Lifecycle Management**: Apply KCM's lifecycle management to adopted clusters
4. **Unified Management**: Manage both new and existing clusters through KCM

### Adoption Process

The adoption process involves:

1. **Kubeconfig Access**: Provide kubeconfig for the existing cluster
2. **Infrastructure Discovery**: KCM discovers existing infrastructure
3. **CAPI Integration**: Creates CAPI resources representing the adopted cluster
4. **Lifecycle Management**: Enables scaling, monitoring, backup, and upgrades

### Key Components

#### 1. Kubeconfig Secret
Contains the kubeconfig for accessing the existing cluster:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: existing-cluster-kubeconfig
type: Opaque
stringData:
  value: |
    # Your cluster kubeconfig content
```

#### 2. Adopted Cluster Credential
References the kubeconfig secret:
```yaml
apiVersion: k0rdent.mirantis.com/v1beta1
kind: Credential
metadata:
  name: adopted-cluster-credential
spec:
  identityRef:
    apiVersion: v1
    kind: Secret
    name: existing-cluster-kubeconfig
```

#### 3. ClusterDeployment with Adopted Configuration
```yaml
apiVersion: k0rdent.mirantis.com/v1beta1
kind: ClusterDeployment
metadata:
  name: adopted-cluster
spec:
  template: adopted-cluster-1-0-1
  credential: adopted-cluster-credential
  adoptedCluster:
    enabled: true
    discoveryMode: "auto"  # or "manual"
    infrastructureMapping:
      provider: "azure"
      region: "eastus"
    lifecycleManagement:
      enabled: true
      scaling:
        enabled: true
      monitoring:
        enabled: true
      backup:
        enabled: true
```

### Discovery Modes

#### Auto Discovery
KCM automatically discovers infrastructure:
```yaml
adoptedCluster:
  discoveryMode: "auto"
  infrastructureMapping:
    provider: "azure"
    region: "eastus"
```

#### Manual Discovery
Manually map existing nodes to CAPI resources:
```yaml
adoptedCluster:
  discoveryMode: "manual"
  infrastructureMapping:
    provider: "azure"
    region: "eastus"
    nodeMapping:
      - nodeName: "aks-agentpool-12345678-vmss000000"
        machineName: "worker-machine-1"
```

### Lifecycle Management Features

#### Scaling
```yaml
lifecycleManagement:
  scaling:
    enabled: true
    autoScaling: true
    minNodes: 2
    maxNodes: 10
    nodeGroups:
      - name: "worker-pool"
        minNodes: 2
        maxNodes: 8
```

#### Monitoring
```yaml
lifecycleManagement:
  monitoring:
    enabled: true
    metrics:
      - "cpu_usage_percent"
      - "memory_usage_percent"
    healthChecks:
      enabled: true
      interval: "30s"
```

#### Backup
```yaml
lifecycleManagement:
  backup:
    enabled: true
    schedule: "0 1 * * *"
    retention: "30d"
    storage:
      type: "azure"
      location: "https://storage.blob.core.windows.net/backups/"
```

### Related Resources

- [KCM Documentation](../docs/)
- [Azure AKS Templates](../templates/cluster/azure-aks/)
- [Adopted Cluster Templates](../templates/cluster/adopted-cluster/)
- [Cluster API Documentation](https://cluster-api.sigs.k8s.io/)
- [Azure Service Operator](https://azure.github.io/azure-service-operator/) 