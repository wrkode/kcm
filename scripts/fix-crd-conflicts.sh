#!/bin/bash

# Script to fix CRD conflicts when installing KCM
# This script helps resolve conflicts with existing cert-manager CRDs

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if kubectl is available
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed or not in PATH"
        exit 1
    fi
    
    if ! command -v helm &> /dev/null; then
        print_error "helm is not installed or not in PATH"
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

# Check for existing cert-manager installation
check_existing_cert_manager() {
    print_status "Checking for existing cert-manager installation..."
    
    if kubectl get namespace cert-manager > /dev/null 2>&1; then
        print_warning "cert-manager namespace exists"
        
        if kubectl get crd certificaterequests.cert-manager.io > /dev/null 2>&1; then
            print_warning "cert-manager CRDs exist in the cluster"
            return 0
        else
            print_status "cert-manager namespace exists but CRDs not found"
            return 1
        fi
    else
        print_status "No existing cert-manager installation found"
        return 1
    fi
}

# Uninstall existing cert-manager
uninstall_cert_manager() {
    print_status "Uninstalling existing cert-manager..."
    
    # Check if cert-manager is installed via Helm
    if helm list -n cert-manager | grep -q cert-manager; then
        print_status "Found cert-manager Helm release, uninstalling..."
        helm uninstall cert-manager -n cert-manager
    fi
    
    # Delete cert-manager namespace
    if kubectl get namespace cert-manager > /dev/null 2>&1; then
        print_status "Deleting cert-manager namespace..."
        kubectl delete namespace cert-manager
    fi
    
    # Wait for namespace deletion
    print_status "Waiting for cert-manager namespace to be deleted..."
    kubectl wait --for=delete namespace/cert-manager --timeout=60s 2>/dev/null || true
    
    print_success "cert-manager uninstalled successfully"
}

# Create installation script with skip-crds flag
create_skip_crds_script() {
    print_status "Creating installation script with skip-crds flag..."
    
    cat > install-kcm-skip-crds.sh << 'EOF'
#!/bin/bash

# KCM Installation Script with Skip CRDs
# This script installs KCM while skipping CRD installation to avoid conflicts

set -e

echo "Installing KCM with skip-crds flag to avoid CRD conflicts"
echo "Registry: ghcr.io/wrkode/kcm"
echo "Chart Version: 1.1.2"

# Create namespace
kubectl create namespace kcm-system --dry-run=client -o yaml | kubectl apply -f -

# Install KCM using Helm with skip-crds flag
helm install kcm oci://ghcr.io/wrkode/kcm/charts/kcm --version 1.1.2 -n kcm-system --create-namespace --skip-crds

echo "KCM installation completed successfully!"
echo "You can check the status with: kubectl get pods -n kcm-system"
echo ""
echo "Note: CRDs were skipped during installation. If you need cert-manager CRDs,"
echo "you may need to install cert-manager separately or ensure they exist in the cluster."
EOF
    
    chmod +x install-kcm-skip-crds.sh
    print_success "Installation script created: install-kcm-skip-crds.sh"
}

# Show options to user
show_options() {
    echo
    echo "=========================================="
    echo "CRD Conflict Resolution Options"
    echo "=========================================="
    echo
    echo "1. Uninstall existing cert-manager (recommended if you don't need it)"
    echo "2. Install KCM with --skip-crds flag (recommended if you want to keep cert-manager)"
    echo "3. Manual resolution (you handle it yourself)"
    echo
    read -p "Choose an option (1-3): " choice
    
    case $choice in
        1)
            uninstall_cert_manager
            print_success "Now you can run: ./install-kcm-custom.sh"
            ;;
        2)
            create_skip_crds_script
            print_success "Now you can run: ./install-kcm-skip-crds.sh"
            ;;
        3)
            print_status "Manual resolution selected"
            echo
            echo "You can:"
            echo "1. Uninstall cert-manager manually: helm uninstall cert-manager -n cert-manager"
            echo "2. Install KCM with skip-crds: helm install kcm oci://ghcr.io/wrkode/kcm/charts/kcm --version 1.1.2 -n kcm-system --create-namespace --skip-crds"
            echo "3. Install cert-manager separately before installing KCM"
            ;;
        *)
            print_error "Invalid option selected"
            exit 1
            ;;
    esac
}

# Main execution
main() {
    echo "=========================================="
    echo "KCM CRD Conflict Resolution Script"
    echo "=========================================="
    echo
    echo "This script helps resolve conflicts with existing cert-manager CRDs"
    echo "when installing KCM."
    echo
    echo "The error you encountered is because cert-manager CRDs already exist"
    echo "in your cluster, and Helm cannot import them into the KCM release."
    echo
    echo "=========================================="
    echo
    
    check_prerequisites
    
    if check_existing_cert_manager; then
        print_warning "Existing cert-manager installation detected"
        show_options
    else
        print_status "No existing cert-manager found"
        print_success "You can proceed with normal installation: ./install-kcm-custom.sh"
    fi
}

# Run main function
main "$@" 