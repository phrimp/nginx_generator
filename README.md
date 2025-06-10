# Nginx Server Manager

A simple, focused Go tool for adding new server blocks to existing nginx configuration files. Supports both static file servers and reverse proxy configurations with interactive input and preview features.

## Features

- 🎯 **Single Purpose**: Only adds server blocks to existing nginx configs
- 🔧 **Two Server Types**: Static file servers and reverse proxy servers
- 📝 **Dual Input Modes**: File-based configuration OR interactive terminal input
- 👀 **Preview Mode**: Review changes before applying them
- 🛡️ **Automatic Backup**: Creates backups before modifying nginx.conf
- ⚡ **Simple & Fast**: Minimal, focused functionality

## Installation

```bash
git clone <repository-url>
cd nginx-server-manager
go build -o tool-name ./
./tool-name
```

## Usage

### Interactive Mode (Manual Input)

```bash
nginx-server-manager -interactive -nginx /etc/nginx/nginx.conf -type static
```

**Interactive Session Example:**
```
🔧 Interactive Configuration Mode
========================================
Enter server name (e.g., example.com): portfolio.phrimp.io.vn
Enter listen port [80]: 80
Enter document root (e.g., /var/www/html): /usr/share/nginx/html
Enter index file [index.html]: portfolio.html

📋 Configuration Preview
==================================================
Server Name: portfolio.phrimp.io.vn
Listen Port: 80
Server Type: static
Document Root: /usr/share/nginx/html
Index File: portfolio.html

🔍 Nginx Configuration Preview
==================================================
events {
    worker_connections 1024;
}

http {
    # ... (existing http directives) ...
    
    # ... (2 existing server block(s)) ...

    # === NEW SERVER BLOCK ===
    server {
        listen 80;
        server_name portfolio.phrimp.io.vn;
        root /usr/share/nginx/html;
        index portfolio.html;
        location / {
            try_files $uri $uri/ =404;
        }
    }
    # === END NEW BLOCK ===
}
==================================================
Do you want to proceed with these changes? (y/N): y
```

### File-Based Configuration

```bash
nginx-server-manager -config static-config.json -nginx /etc/nginx/nginx.conf -type static
```

**static-config.json:**
```json
{
  "listen": "80",
  "server_name": "portfolio.phrimp.io.vn",
  "root": "/usr/share/nginx/html",
  "index": "portfolio.html"
}
```

**Generated server block:**
```nginx
server {
    listen 80;
    server_name portfolio.phrimp.io.vn;
    root /usr/share/nginx/html;
    index portfolio.html;
    location / {
        try_files $uri $uri/ =404;
    }
}
```

### Add Proxy Server

```bash
nginx-server-manager -config proxy-config.json -nginx /etc/nginx/nginx.conf -type proxy
```

**proxy-config.json:**
```json
{
  "listen": "80",
  "server_name": "evolvia.phrimp.io.vn",
  "proxy_port": "8084"
}
```

**Generated server block:**
```nginx
server {
    listen 80;
    server_name evolvia.phrimp.io.vn;
    # Proxy all requests to http://127.0.0.1:8084
    location / {
        proxy_pass http://127.0.0.1:8084;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Forwarded-Port $server_port;
        proxy_cache_bypass $http_upgrade;
        proxy_redirect off;
    }
}
```

## Configuration Options

### Common Fields
- `listen`: Port to listen on (default: "80")
- `server_name`: Domain name for the server

### Static Server Fields
- `root`: Document root directory
- `index`: Index file name (default: "index.html")

### Proxy Server Fields
- `proxy_port`: Port number for proxy target (creates http://127.0.0.1:PORT)
- `proxy_pass`: Full proxy URL (alternative to proxy_port)

## Command Line Options

- `-config`: Path to server configuration file (.json/.yaml)
- `-nginx`: Path to existing nginx.conf file **(required)**
- `-type`: Server type (`static` or `proxy`) **(required)**
- `-interactive`: Enable manual input mode via terminal
- `-preview`: Show preview before applying changes (default: true)
- `-backup`: Create backup before modifying (default: true)
- `-help`: Show help message

## Examples

### Interactive Mode Examples

```bash
# Interactive static file server
nginx-server-manager -interactive -nginx /etc/nginx/nginx.conf -type static

# Interactive proxy server
nginx-server-manager -interactive -nginx /etc/nginx/nginx.conf -type proxy

# Skip preview in interactive mode
nginx-server-manager -interactive -nginx nginx.conf -type proxy -preview=false
```

### File-Based Examples

```bash
# Add static file server from config file
nginx-server-manager -config examples/static-config.json -nginx /etc/nginx/nginx.conf -type static

# Add proxy server from config file
nginx-server-manager -config examples/proxy-config.json -nginx /etc/nginx/nginx.conf -type proxy

# Skip preview and backup
nginx-server-manager -config config.json -nginx nginx.conf -type static -preview=false -backup=false
```

## Project Structure

```
nginx-server-manager/
├── main.go                 # CLI entry point
├── internal/
│   ├── config/            # Configuration loading
│   └── generator/         # Server block generation
├── examples/              # Example configurations
├── Makefile              # Build automation
└── go.mod                # Dependencies
```

## Safety Features

- **Interactive Input**: Manual configuration via terminal prompts
- **Smart Preview**: Shows exactly where the new server block will be inserted
- **Context-Aware Display**: Preview shows existing blocks as summaries ("... (2 existing server blocks) ...")
- **Automatic Backups**: Creates timestamped backups before modification
- **Validation**: Checks for valid http section in nginx.conf
- **Error Handling**: Comprehensive error reporting
- **Non-destructive**: Only adds content, doesn't modify existing blocks
- **Confirmation Required**: Preview mode asks for confirmation before proceeding

## Preview Feature

The preview feature shows you exactly how your nginx configuration will look after adding the new server block:

```
🔍 Nginx Configuration Preview
==================================================
events {
    worker_connections 1024;
}

http {
    # ... (existing http directives) ...
    
    # ... (1 existing server block(s)) ...

    # === NEW SERVER BLOCK ===
    server {
        listen 80;
        server_name evolvia.phrimp.io.vn;
        # Proxy all requests to http://127.0.0.1:8084
        location / {
            proxy_pass http://127.0.0.1:8084;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection 'upgrade';
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_set_header X-Forwarded-Host $host;
            proxy_set_header X-Forwarded-Port $server_port;
            proxy_cache_bypass $http_upgrade;
            proxy_redirect off;
        }
    }
    # === END NEW BLOCK ===
}
==================================================
Do you want to proceed with these changes? (y/N):
```

## Development

```bash
# Build the project
make build

# Run tests
make test

# Clean build artifacts
make clean

# Run examples
make examples
```

This tool is designed to be simple, safe, and focused on the single task of adding server blocks to nginx configurations.
