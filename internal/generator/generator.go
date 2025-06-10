package generator

import (
	"fmt"
	"nginx_tool/internal/config"
	"os"
	"regexp"
	"strings"
	"time"
)

type Generator struct{}

func New() *Generator {
	return &Generator{}
}

func (g *Generator) AddServerToNginx(cfg *config.ServerConfig, nginxPath, serverType string, backup bool) error {
	if backup {
		backupPath := fmt.Sprintf("%s.backup.%d", nginxPath, time.Now().Unix())
		if err := g.copyFile(nginxPath, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		fmt.Printf("📋 Backup created: %s\n", backupPath)
	}

	nginxContent, err := os.ReadFile(nginxPath)
	if err != nil {
		return fmt.Errorf("failed to read nginx config: %w", err)
	}

	var serverBlock string
	switch serverType {
	case "static":
		serverBlock = g.GenerateStaticServerBlock(cfg)
	case "proxy":
		serverBlock = g.GenerateProxyServerBlock(cfg)
	default:
		return fmt.Errorf("unsupported server type: %s", serverType)
	}

	modifiedContent, err := g.addServerBlock(string(nginxContent), serverBlock)
	if err != nil {
		return fmt.Errorf("failed to add server block: %w", err)
	}

	if err := os.WriteFile(nginxPath, []byte(modifiedContent), 0644); err != nil {
		return fmt.Errorf("failed to write nginx config: %w", err)
	}

	return nil
}

func (g *Generator) GenerateStaticServerBlock(cfg *config.ServerConfig) string {
	return fmt.Sprintf(`    server {
        listen %s;
        server_name %s;
        root %s;
        index %s;
        location / {
            try_files $uri $uri/ =404;
        }
    }`, cfg.Listen, cfg.ServerName, cfg.Root, cfg.Index)
}

func (g *Generator) GenerateProxyServerBlock(cfg *config.ServerConfig) string {
	proxyTarget := cfg.ProxyPass
	if proxyTarget == "" && cfg.ProxyPort != "" {
		proxyTarget = fmt.Sprintf("http://127.0.0.1:%s", cfg.ProxyPort)
	}

	return fmt.Sprintf(`    server {
        listen %s;
        server_name %s;
        # Proxy all requests to %s
        location / {
            proxy_pass %s;
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
    }`, cfg.Listen, cfg.ServerName, proxyTarget, proxyTarget)
}

// GeneratePreview creates a preview of how the nginx config will look after modification
func (g *Generator) GeneratePreview(nginxPath, serverBlock string) (string, error) {
	// Read existing nginx configuration
	nginxContent, err := os.ReadFile(nginxPath)
	if err != nil {
		return "", fmt.Errorf("failed to read nginx config: %w", err)
	}

	content := string(nginxContent)

	httpRegex := regexp.MustCompile(`(?s)(http\s*\{)(.*?)(\})`)
	matches := httpRegex.FindStringSubmatch(content)

	if len(matches) != 4 {
		return "", fmt.Errorf("could not find http section in nginx configuration")
	}

	httpStart := matches[1]
	httpContent := matches[2]
	httpEnd := matches[3]

	serverRegex := regexp.MustCompile(`(?s)server\s*\{`)
	existingServers := serverRegex.FindAllString(httpContent, -1)
	serverCount := len(existingServers)

	var preview strings.Builder

	beforeHttp := strings.Split(content, httpStart)[0]
	beforeLines := strings.Split(strings.TrimSpace(beforeHttp), "\n")
	if len(beforeLines) > 3 {
		preview.WriteString("...\n")
		preview.WriteString(strings.Join(beforeLines[len(beforeLines)-2:], "\n"))
		preview.WriteString("\n")
	} else {
		preview.WriteString(beforeHttp)
	}

	preview.WriteString(httpStart)
	preview.WriteString("\n")

	if serverCount > 0 {
		preview.WriteString(fmt.Sprintf("    # ... (%d existing server block(s)) ...\n", serverCount))
		preview.WriteString("\n")
	} else {
		httpLines := strings.Split(strings.TrimSpace(httpContent), "\n")
		if len(httpLines) > 0 && strings.TrimSpace(httpLines[0]) != "" {
			preview.WriteString("    # ... (existing http directives) ...\n")
			preview.WriteString("\n")
		}
	}

	preview.WriteString("    # === NEW SERVER BLOCK ===\n")
	preview.WriteString(serverBlock)
	preview.WriteString("\n")
	preview.WriteString("    # === END NEW BLOCK ===\n")

	preview.WriteString(httpEnd)

	afterHttp := strings.Split(content, httpEnd)[1]
	if strings.TrimSpace(afterHttp) != "" {
		afterLines := strings.Split(strings.TrimSpace(afterHttp), "\n")
		if len(afterLines) > 2 {
			preview.WriteString("\n")
			preview.WriteString(strings.Join(afterLines[:2], "\n"))
			preview.WriteString("\n...")
		} else {
			preview.WriteString(afterHttp)
		}
	}

	return preview.String(), nil
}

func (g *Generator) addServerBlock(nginxContent, serverBlock string) (string, error) {
	// Find the http section
	httpRegex := regexp.MustCompile(`(?s)(http\s*\{)(.*?)(\})`)
	matches := httpRegex.FindStringSubmatch(nginxContent)

	if len(matches) != 4 {
		return "", fmt.Errorf("could not find http section in nginx configuration")
	}

	httpStart := matches[1]
	httpContent := matches[2]
	httpEnd := matches[3]

	httpContent = strings.TrimRight(httpContent, " \t\n")

	newHttpContent := httpContent + "\n\n" + serverBlock + "\n"

	result := httpRegex.ReplaceAllString(nginxContent, httpStart+newHttpContent+httpEnd)

	return result, nil
}

// copyFile creates a backup copy of a file
func (g *Generator) copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
