# nginxproxymanager-cli

A command-line interface (CLI) written in Go for managing Nginx Proxy Manager via its REST API.

## Features

- **Authentication**: Automatic token-based authentication with Nginx Proxy Manager API
- **Proxy Host Management**: List, create, and delete proxy hosts
- **Flexible Configuration**: Support for environment variables and command-line flags
- **Easy to Use**: Simple commands with helpful output

## Installation

### Build from Source

```bash
git clone https://github.com/arvesv/nginxproxymanager-cli.git
cd nginxproxymanager-cli
go build -o nginxproxymanager-cli
```

## Configuration

The CLI can be configured using command-line flags or environment variables:

### Environment Variables

- `NPM_API_URL`: API URL (default: `http://dockernuc:81/api`)
- `NPM_USERNAME`: Username for authentication
- `NPM_PASSWORD`: Password for authentication

### Command-line Flags

- `-a, --api-url`: Nginx Proxy Manager API URL
- `-u, --username`: Username for authentication  
- `-p, --password`: Password for authentication

## Usage

### Basic Usage

```bash
# Set credentials via environment variables
export NPM_USERNAME="admin@example.com"
export NPM_PASSWORD="password"

# Or use command-line flags
./nginxproxymanager-cli -u admin@example.com -p password [command]
```

### Commands

#### List Proxy Hosts

List all existing proxy hosts:

```bash
./nginxproxymanager-cli list
```

Example output:
```
Found 2 proxy hosts:

ID: 1
Domain Names: [example.com www.example.com]
Forward: http://192.168.1.100:8080
Enabled: true
SSL Forced: false
---
ID: 2
Domain Names: [api.example.com]
Forward: https://192.168.1.101:8443
Enabled: true
SSL Forced: true
---
```

#### Create Proxy Host

Create a new proxy host:

```bash
./nginxproxymanager-cli create \
  --domain "example.com" \
  --forward-host "192.168.1.100" \
  --forward-port 8080 \
  --forward-scheme "http"
```

Options:
- `--domain`: Domain name for the proxy host (required)
- `--forward-host`: Target host to forward requests to (required)
- `--forward-port`: Target port (required)
- `--forward-scheme`: Protocol scheme - `http` or `https` (default: `http`)

#### Delete Proxy Host

Delete a proxy host by its ID:

```bash
./nginxproxymanager-cli delete --id 1
```

### Help

Get help for any command:

```bash
./nginxproxymanager-cli --help
./nginxproxymanager-cli [command] --help
```

## API Endpoints

The CLI interacts with the following Nginx Proxy Manager API endpoints:

- `POST /api/tokens` - Authentication
- `GET /api/nginx/proxy-hosts` - List proxy hosts
- `POST /api/nginx/proxy-hosts` - Create proxy host
- `DELETE /api/nginx/proxy-hosts/{id}` - Delete proxy host

## Error Handling

The CLI provides clear error messages for common scenarios:

- Authentication failures
- Network connectivity issues
- Invalid parameters
- API errors

## Examples

### Complete Workflow Example

```bash
# Set environment variables
export NPM_API_URL="http://dockernuc:81/api"
export NPM_USERNAME="admin@example.com" 
export NPM_PASSWORD="changeme"

# List current proxy hosts
./nginxproxymanager-cli list

# Create a new proxy host for a web application
./nginxproxymanager-cli create \
  --domain "myapp.example.com" \
  --forward-host "192.168.1.50" \
  --forward-port 3000 \
  --forward-scheme "http"

# Create a proxy host for an HTTPS backend
./nginxproxymanager-cli create \
  --domain "secure.example.com" \
  --forward-host "192.168.1.51" \
  --forward-port 8443 \
  --forward-scheme "https"

# List hosts again to see the new entries
./nginxproxymanager-cli list

# Delete a proxy host (replace 3 with actual ID)
./nginxproxymanager-cli delete --id 3
```

## License

This project is licensed under the GNU General Public License v2.0 - see the [LICENSE](LICENSE) file for details.