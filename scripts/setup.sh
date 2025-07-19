#!/bin/bash

# Go Radio v2 Setup Script
# This script handles dependency installation and building for the local self-hosted radio service

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

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check Go version
check_go_version() {
    if ! command_exists go; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi
    
    GO_VERSION=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | sed 's/go//')
    MAJOR=$(echo $GO_VERSION | cut -d. -f1)
    MINOR=$(echo $GO_VERSION | cut -d. -f2)
    
    if [ "$MAJOR" -lt 1 ] || ([ "$MAJOR" -eq 1 ] && [ "$MINOR" -lt 21 ]); then
        print_error "Go version $GO_VERSION is too old. Please install Go 1.21 or later."
        exit 1
    fi
    
    print_success "Go version $GO_VERSION detected"
}

# Function to check Node.js version
check_node_version() {
    if ! command_exists node; then
        print_error "Node.js is not installed. Please install Node.js 18 or later."
        exit 1
    fi
    
    NODE_VERSION=$(node --version | sed 's/v//')
    MAJOR=$(echo $NODE_VERSION | cut -d. -f1)
    
    if [ "$MAJOR" -lt 18 ]; then
        print_error "Node.js version $NODE_VERSION is too old. Please install Node.js 18 or later."
        exit 1
    fi
    
    print_success "Node.js version $NODE_VERSION detected"
}

# Function to check if yarn is available
check_yarn() {
    if ! command_exists yarn; then
        print_warning "Yarn is not installed. Installing yarn..."
        npm install -g yarn
    fi
    print_success "Yarn is available"
}

# Function to install Go dependencies
install_go_deps() {
    print_status "Installing Go dependencies..."
    go mod download
    go mod tidy
    print_success "Go dependencies installed"
}

# Function to install frontend dependencies
install_frontend_deps() {
    print_status "Installing frontend dependencies..."
    cd client
    yarn install
    cd ..
    print_success "Frontend dependencies installed"
}

# Function to build frontend
build_frontend() {
    print_status "Building frontend..."
    cd client
    yarn build
    cd ..
    print_success "Frontend built successfully"
}

# Function to build backend
build_backend() {
    print_status "Building backend..."
    go build -o bin/go-radio-server cmd/server/main.go
    go build -o bin/go-radio-setup cmd/setup/main.go
    print_success "Backend built successfully"
}

# Function to run setup wizard
run_setup_wizard() {
    print_status "Running setup wizard..."
    if [ -f "bin/go-radio-setup" ]; then
        ./bin/go-radio-setup
    else
        print_error "Setup binary not found. Please run build first."
        exit 1
    fi
}

# Function to create directories
create_directories() {
    print_status "Creating necessary directories..."
    mkdir -p bin
    mkdir -p data/audio
    print_success "Directories created"
}

# Function to check for CGO requirement
check_cgo() {
    print_status "Checking CGO availability for SQLite..."
    if ! command_exists gcc; then
        print_warning "GCC not found. SQLite support requires CGO."
        print_warning "On Ubuntu/Debian: sudo apt-get install build-essential"
        print_warning "On CentOS/RHEL: sudo yum groupinstall 'Development Tools'"
        print_warning "On macOS: xcode-select --install"
    else
        print_success "CGO support available"
    fi
}

# Main execution
main() {
    echo "=================================================="
    echo "         Go Radio v2 Setup Script"
    echo "=================================================="
    echo ""
    
    # Check prerequisites
    print_status "Checking prerequisites..."
    check_go_version
    check_node_version
    check_yarn
    check_cgo
    
    # Create directories
    create_directories
    
    # Install dependencies
    install_go_deps
    install_frontend_deps
    
    # Build applications
    build_frontend
    build_backend
    
    echo ""
    echo "=================================================="
    print_success "Setup completed successfully!"
    echo "=================================================="
    echo ""
    print_status "Next steps:"
    echo "  1. Run './bin/go-radio-setup' to configure your radio service"
    echo "  2. Run 'make run' or './bin/go-radio-server' to start the server"
    echo "  3. Open http://localhost:8080 in your browser"
    echo ""
    print_status "Optional: Run the configuration wizard now? (y/n)"
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        run_setup_wizard
    fi
}

# Handle command line arguments
case "${1:-setup}" in
    "setup")
        main
        ;;
    "deps")
        check_go_version
        check_node_version
        check_yarn
        install_go_deps
        install_frontend_deps
        ;;
    "build")
        create_directories
        build_frontend
        build_backend
        ;;
    "config")
        run_setup_wizard
        ;;
    "help"|"-h"|"--help")
        echo "Go Radio v2 Setup Script"
        echo ""
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  setup    - Full setup (default)"
        echo "  deps     - Install dependencies only"
        echo "  build    - Build applications only"
        echo "  config   - Run configuration wizard"
        echo "  help     - Show this help message"
        ;;
    *)
        print_error "Unknown command: $1"
        echo "Run '$0 help' for usage information"
        exit 1
        ;;
esac