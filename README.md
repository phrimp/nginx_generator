# Nginx Server Manager

A simple, focused Go tool for adding new server blocks to existing nginx configuration files. Supports both static file servers and reverse proxy configurations with automatic nginx detection, interactive input, and preview features.

## Features

- 🎯 **Single Purpose**: Only adds server blocks to existing nginx configs
- 🔧 **Two Server Types**: Static file servers and reverse proxy servers
- 🔍 **Auto-Detection**: Automatically finds nginx config files on Linux systems
- 📝 **Dual Input Modes**: File-based configuration OR interactive terminal input
- 👀 **Preview Mode**: Review changes before applying them
- 🛡️ **Automatic Backup**: Creates timestamped backups before modifying
- ⚡ **Simple & Fast**: Minimal, focused functionality

## Installation

```bash
git clone <repository-url>
cd nginx-server-manager
go build -o tool-name ./
./tool-name -help
```

## Auto-Detection Feature (Linux)

The tool automatically detects nginx configuration files on Linux systems through multiple methods:

### Detection Methods
1. **Common Paths**: Checks standard nginx config locations
2. **Nginx Binary**: Extracts config path from nginx binary
3. **Running Process**: Finds config from active nginx process

### Search Locations
- `/etc/nginx/nginx.conf` (most common)
- `/usr/local/etc/nginx/nginx.conf`
- `/usr/local/nginx/conf/nginx.conf`
- `/opt/nginx/conf/nginx.conf`
- `/etc/nginx.conf`

### Smart Validation
- Validates detected files contain nginx directives
- Falls back through multiple detection methods
- Shows detected path before proceeding

## Usage

### Auto-Detection Mode (Simplest)

```bash
# Interactive with auto-detection
nginx-server-manager -interactive -type static
```

Output:
```
🔍 Auto-detecting nginx configuration...
🔍 Auto-detected nginx config: /etc/nginx/nginx.conf

🔧 Interactive Configuration Mode
========================================
Enter server name (e.g., example.com): mysite.com
Enter listen port [80]: 80
Enter document root (e.g., /var/www/html): /var/www/mysite
Enter index file [index.html]: index.html
```

```bash
# File-based with auto-detection
nginx-server-manager -config static-config.json -type static
```

### Manual Path Mode

```bash
# Interactive with manual nginx path
nginx-server-manager -interactive -nginx /custom/path/nginx.conf -type proxy
```

### File-Based Configuration

```bash
# Auto-detect nginx config
nginx-server-manager -config static-config.json -type static

# Manual nginx path
nginx-server-manager -config static-config.json -nginx /etc/nginx/nginx.conf -type static
```

## Configuration Files

### Static File Server (JSON)
```json
{
  "listen": "80",
  "server_name": "portfolio.phrimp.io.vn",
  "root": "/usr/share/nginx/html",
  "index": "portfolio.html"
}
```

### Proxy Server (JSON)
```json
{
  "listen": "80",
  "server_name": "evolvia.phrimp.io.vn",
  "proxy_port": "8084"
}
```

### Proxy Server with Full URL
```json
{
  "listen": "80",
  "server_name": "api.phrimp.io.vn",
  "proxy_pass": "http://127.0.0.1:3000"
}
```

### YAML Configuration
```yaml
listen: "80"
server_name: "blog.phrimp.io.vn"
root: "/var/www/blog"
index: "index.html"
```

## Generated Server Blocks

### Static File Server
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

### Reverse Proxy Server
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

## Command Line Options

- `-config`: Path to server configuration file (.json/.yaml)
- `-nginx`: Path to existing nginx.conf file (auto-detected if not specified)
- `-type`: Server type (`static` or `proxy`) **required**
- `-interactive`: Enable manual input mode via terminal
- `-auto-detect`: Auto-detect nginx configuration file (default: true)
- `-preview`: Show preview before applying changes (default: true)
- `-backup`: Create backup before modifying (default: true)
- `-help`: Show help message

## Examples

### Auto-Detection Examples (Recommended)

```bash
# Simplest usage - auto-detect nginx config
nginx-server-manager -interactive -type static

# File-based with auto-detection
nginx-server-manager -config static-config.json -type proxy

# Auto-detect but skip preview
nginx-server-manager -interactive -type proxy -preview=false
```

### Manual Path Examples

```bash
# Interactive with specific nginx path
nginx-server-manager -interactive -nginx /etc/nginx/nginx.conf -type static

# File-based with specific nginx path
nginx-server-manager -config proxy-config.json -nginx /custom/nginx.conf -type proxy

# Disable auto-detection (force manual path)
nginx-server-manager -config config.json -nginx nginx.conf -type static -auto-detect=false
```

## Preview Feature

The preview feature shows exactly how your nginx configuration will look after adding the new server block:

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

## Safety Features

- **Automatic nginx Detection**: Finds nginx config automatically on Linux systems
- **Multiple Detection Methods**: Common paths → nginx binary → running process
- **Interactive Input**: Manual configuration via terminal prompts
- **Smart Preview**: Shows exactly where the new server block will be inserted
- **Context-Aware Display**: Preview shows existing blocks as summaries
- **Automatic Backups**: Creates timestamped backups before modification
- **Validation**: Checks for valid http section and validates detected configs
- **Error Handling**: Comprehensive error reporting with fallback detection
- **Non-destructive**: Only adds content, doesn't modify existing blocks
- **Confirmation Required**: Preview mode asks for confirmation before proceeding

## Auto-Detection Process

### Successful Detection
```
🔍 Auto-detecting nginx configuration...

1. Checking common paths:
   ✓ /etc/nginx/nginx.conf (found)
   
2. Validating nginx config file...
   ✓ Contains nginx directives (http, server, location)
   
🔍 Auto-detected nginx config: /etc/nginx/nginx.conf
```

### Binary Detection Fallback
```
🔍 Auto-detecting nginx configuration...

1. Checking common paths: (none found)
2. Looking for nginx binary: /usr/sbin/nginx
3. Extracting config from binary: /etc/nginx/nginx.conf
4. Validating config file...

🔍 Auto-detected nginx config: /etc/nginx/nginx.conf
```

### Process Detection Fallback
```
🔍 Auto-detecting nginx configuration...

1. Checking common paths: (none found)
2. Looking for nginx binary: (not found)
3. Checking running nginx process...
4. Found config in process: /custom/path/nginx.conf

🔍 Auto-detected nginx config: /custom/path/nginx.conf
```

## Project Structure

```
nginx-server-manager/
├── main.go                         # CLI entry point with auto-detection
├── internal/
│   ├── config/
│   │   └── config.go              # Configuration loading
│   └── generator/
│       └── generator.go           # Server block generation
├── examples/                      # Example configurations
│   ├── static-config.json
│   ├── proxy-config.json
│   ├── static-config.yaml
│   ├── proxy-config.yaml
│   └── nginx.conf
├── go.mod                         # Go module definition
└── README.md                      # This file
```

## Example Workflow

1. **Run in auto-detection mode:**
   ```bash
   nginx-server-manager -interactive -type proxy
   ```

2. **Auto-detection finds nginx config:**
   ```
   🔍 Auto-detected nginx config: /etc/nginx/nginx.conf
   ```

3. **Enter configuration interactively:**
   ```
   Enter server name: api.myapp.com
   Enter listen port [80]: 443
   Enter proxy target: 3000
   ```

4. **Review preview:**
   ```
   🔍 Nginx Configuration Preview
   ==================================================
   # === NEW SERVER BLOCK ===
   server {
       listen 443;
       server_name api.myapp.com;
       # Proxy all requests to http://127.0.0.1:3000
       location / {
           proxy_pass http://127.0.0.1:3000;
           # ... (complete proxy configuration)
       }
   }
   # === END NEW BLOCK ===
   ==================================================
   ```

5. **Confirm and apply:**
   ```
   Do you want to proceed with these changes? (y/N): y
   ✅ Server block added successfully
   📋 Backup created: nginx.conf.backup.1623456789
   ```

## License

MIT License
