#!/bin/bash
# Copyright 2025
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

# Adopted Cluster Lifecycle Management Test Script
# This script demonstrates the full lifecycle management capabilities for adopted clusters

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EXAMPLES_DIR="${SCRIPT_DIR}/../examples"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if kubectl is available
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed or not in PATH"
        exit 1
    fi
    
    # Check if KCM is installed
    if ! kubectl get namespace kcm-system &> /dev/null; then
        log_error "KCM is not installed. Please install KCM first."
        exit 1
    fi
    
    # Check if adopted-cluster template exists
    if ! kubectl get clustertemplate adopted-cluster &> /dev/null; then
        log_warning "adopted-cluster template not found. Creating it..."
        create_adopted_cluster_template
    fi
    
    log_success "Prerequisites check passed"
}

# Create adopted cluster template if it doesn't exist
create_adopted_cluster_template() {
    log_info "Creating adopted-cluster template..."
    
    cat <<EOF | kubectl apply -f -
apiVersion: k0rdent.mirantis.com/v1beta1
kind: ClusterTemplate
metadata:
  name: adopted-cluster
  namespace: default
spec:
  description: "Template for adopting existing Kubernetes clusters"
  type: "adopted"
  provider: "infrastructure-internal"
EOF
    
    log_success "Adopted cluster template created"
}

# Test minimal adopted cluster configuration
test_minimal_configuration() {
    log_info "Testing minimal adopted cluster configuration..."
    
    # Apply minimal configuration
    kubectl apply -f "${EXAMPLES_DIR}/adopted-cluster-minimal.yaml"
    
    # Wait for the cluster deployment to be created
    log_info "Waiting for ClusterDeployment to be created..."
    kubectl wait --for=condition=Ready clusterdeployment/adopted-cluster-minimal --timeout=60s || {
        log_error "ClusterDeployment failed to become ready"
        kubectl describe clusterdeployment adopted-cluster-minimal
        return 1
    }
    
    # Check adopted cluster status
    log_info "Checking adopted cluster status..."
    kubectl get clusterdeployment adopted-cluster-minimal -o jsonpath='{.status.adoptedClusterStatus}' | jq .
    
    log_success "Minimal configuration test passed"
}

# Test auto discovery
test_auto_discovery() {
    log_info "Testing auto discovery functionality..."
    
    # Apply auto discovery configuration
    kubectl apply -f "${EXAMPLES_DIR}/adopted-cluster-example.yaml"
    
    # Wait for discovery to complete
    log_info "Waiting for infrastructure discovery to complete..."
    timeout=300
    while [ $timeout -gt 0 ]; do
        discovery_phase=$(kubectl get clusterdeployment adopted-cluster-example -o jsonpath='{.status.adoptedClusterStatus.discoveryStatus.phase}' 2>/dev/null || echo "NotStarted")
        
        if [ "$discovery_phase" = "Completed" ]; then
            log_success "Infrastructure discovery completed"
            break
        elif [ "$discovery_phase" = "Failed" ]; then
            log_error "Infrastructure discovery failed"
            kubectl describe clusterdeployment adopted-cluster-example
            return 1
        fi
        
        log_info "Discovery phase: $discovery_phase, waiting..."
        sleep 10
        timeout=$((timeout - 10))
    done
    
    if [ $timeout -le 0 ]; then
        log_error "Discovery timed out"
        return 1
    fi
    
    # Check discovered resources
    log_info "Checking discovered resources..."
    kubectl get clusterdeployment adopted-cluster-example -o jsonpath='{.status.adoptedClusterStatus.discoveryStatus.discoveredResources}' | jq .
    
    log_success "Auto discovery test passed"
}

# Test CAPI integration
test_capi_integration() {
    log_info "Testing CAPI integration..."
    
    # Wait for CAPI integration to complete
    log_info "Waiting for CAPI integration to complete..."
    timeout=300
    while [ $timeout -gt 0 ]; do
        capi_phase=$(kubectl get clusterdeployment adopted-cluster-example -o jsonpath='{.status.adoptedClusterStatus.capiStatus.phase}' 2>/dev/null || echo "NotStarted")
        
        if [ "$capi_phase" = "Completed" ]; then
            log_success "CAPI integration completed"
            break
        elif [ "$capi_phase" = "Failed" ]; then
            log_error "CAPI integration failed"
            kubectl describe clusterdeployment adopted-cluster-example
            return 1
        fi
        
        log_info "CAPI integration phase: $capi_phase, waiting..."
        sleep 10
        timeout=$((timeout - 10))
    done
    
    if [ $timeout -le 0 ]; then
        log_error "CAPI integration timed out"
        return 1
    fi
    
    # Check generated CAPI resources
    log_info "Checking generated CAPI resources..."
    kubectl get clusterdeployment adopted-cluster-example -o jsonpath='{.status.adoptedClusterStatus.capiStatus.generatedResources}' | jq .
    
    log_success "CAPI integration test passed"
}

# Test monitoring
test_monitoring() {
    log_info "Testing monitoring functionality..."
    
    # Wait for monitoring to complete
    log_info "Waiting for monitoring to complete..."
    timeout=300
    while [ $timeout -gt 0 ]; do
        monitoring_phase=$(kubectl get clusterdeployment adopted-cluster-example -o jsonpath='{.status.adoptedClusterStatus.monitoringStatus.phase}' 2>/dev/null || echo "NotStarted")
        
        if [ "$monitoring_phase" = "Completed" ]; then
            log_success "Monitoring completed"
            break
        elif [ "$monitoring_phase" = "Failed" ]; then
            log_error "Monitoring failed"
            kubectl describe clusterdeployment adopted-cluster-example
            return 1
        fi
        
        log_info "Monitoring phase: $monitoring_phase, waiting..."
        sleep 10
        timeout=$((timeout - 10))
    done
    
    if [ $timeout -le 0 ]; then
        log_error "Monitoring timed out"
        return 1
    fi
    
    # Check monitoring status
    log_info "Checking monitoring status..."
    kubectl get clusterdeployment adopted-cluster-example -o jsonpath='{.status.adoptedClusterStatus.monitoringStatus}' | jq .
    
    log_success "Monitoring test passed"
}

# Test scaling (simulated)
test_scaling() {
    log_info "Testing scaling functionality (simulated)..."
    
    # Check current scaling status
    log_info "Checking current scaling status..."
    kubectl get clusterdeployment adopted-cluster-example -o jsonpath='{.status.adoptedClusterStatus.scalingStatus}' | jq .
    
    # Simulate scaling operation by updating the configuration
    log_info "Simulating scaling operation..."
    kubectl patch clusterdeployment adopted-cluster-example --type='merge' -p='{"spec":{"adoptedCluster":{"lifecycleManagement":{"scaling":{"maxNodes":5}}}}}'
    
    # Wait for scaling to be processed
    sleep 30
    
    # Check updated scaling status
    log_info "Checking updated scaling status..."
    kubectl get clusterdeployment adopted-cluster-example -o jsonpath='{.status.adoptedClusterStatus.scalingStatus}' | jq .
    
    log_success "Scaling test passed"
}

# Test backup (simulated)
test_backup() {
    log_info "Testing backup functionality (simulated)..."
    
    # Check backup status
    log_info "Checking backup status..."
    kubectl get clusterdeployment adopted-cluster-example -o jsonpath='{.status.adoptedClusterStatus.backupStatus}' | jq .
    
    # Simulate backup operation
    log_info "Simulating backup operation..."
    kubectl patch clusterdeployment adopted-cluster-example --type='merge' -p='{"spec":{"adoptedCluster":{"lifecycleManagement":{"backup":{"enabled":true}}}}}'
    
    # Wait for backup to be processed
    sleep 30
    
    # Check updated backup status
    log_info "Checking updated backup status..."
    kubectl get clusterdeployment adopted-cluster-example -o jsonpath='{.status.adoptedClusterStatus.backupStatus}' | jq .
    
    log_success "Backup test passed"
}

# Test manual discovery
test_manual_discovery() {
    log_info "Testing manual discovery functionality..."
    
    # Apply manual discovery configuration
    kubectl apply -f "${EXAMPLES_DIR}/adopted-cluster-manual.yaml"
    
    # Wait for manual discovery to complete
    log_info "Waiting for manual discovery to complete..."
    timeout=300
    while [ $timeout -gt 0 ]; do
        discovery_phase=$(kubectl get clusterdeployment adopted-cluster-manual -o jsonpath='{.status.adoptedClusterStatus.discoveryStatus.phase}' 2>/dev/null || echo "NotStarted")
        
        if [ "$discovery_phase" = "Completed" ]; then
            log_success "Manual discovery completed"
            break
        elif [ "$discovery_phase" = "Failed" ]; then
            log_error "Manual discovery failed"
            kubectl describe clusterdeployment adopted-cluster-manual
            return 1
        fi
        
        log_info "Manual discovery phase: $discovery_phase, waiting..."
        sleep 10
        timeout=$((timeout - 10))
    done
    
    if [ $timeout -le 0 ]; then
        log_error "Manual discovery timed out"
        return 1
    fi
    
    log_success "Manual discovery test passed"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up test resources..."
    
    # Delete test cluster deployments
    kubectl delete clusterdeployment adopted-cluster-minimal --ignore-not-found=true
    kubectl delete clusterdeployment adopted-cluster-example --ignore-not-found=true
    kubectl delete clusterdeployment adopted-cluster-manual --ignore-not-found=true
    
    log_success "Cleanup completed"
}

# Main test function
main() {
    log_info "Starting adopted cluster lifecycle management tests..."
    
    # Set up trap for cleanup
    trap cleanup EXIT
    
    # Check prerequisites
    check_prerequisites
    
    # Run tests
    test_minimal_configuration
    test_auto_discovery
    test_capi_integration
    test_monitoring
    test_scaling
    test_backup
    test_manual_discovery
    
    log_success "All tests passed!"
}

# Run main function
main "$@" 