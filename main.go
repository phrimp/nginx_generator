package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"nginx_tool/internal/config"
	"nginx_tool/internal/generator"
	"os"
	"strings"
)

func main() {
	var (
		configPath  = flag.String("config", "", "Path to server configuration JSON/YAML file")
		nginxPath   = flag.String("nginx", "", "Path to existing nginx.conf file")
		serverType  = flag.String("type", "static", "Server type: 'static' or 'proxy'")
		interactive = flag.Bool("interactive", false, "Manual input mode via terminal")
		preview     = flag.Bool("preview", true, "Show preview before applying changes")
		backup      = flag.Bool("backup", true, "Create backup of nginx.conf before modification")
		help        = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		showUsage()
		return
	}

	if *nginxPath == "" {
		log.Fatal("Error: nginx path is required")
	}

	var cfg *config.ServerConfig
	var err error

	// Get configuration either from file or interactive input
	if *interactive {
		cfg, err = getInteractiveConfig(*serverType)
		if err != nil {
			log.Fatalf("Error getting interactive config: %v", err)
		}
	} else {
		if *configPath == "" {
			log.Fatal("Error: config path is required when not using interactive mode")
		}
		cfg, err = config.Load(*configPath)
		if err != nil {
			log.Fatalf("Error loading configuration: %v", err)
		}
	}

	// Validate server type
	if *serverType != "static" && *serverType != "proxy" {
		log.Fatal("Error: type must be either 'static' or 'proxy'")
	}

	// Create generator
	gen := generator.New()

	// Show preview if requested
	if *preview {
		shouldProceed, err := showPreview(gen, cfg, *nginxPath, *serverType)
		if err != nil {
			log.Fatalf("Error generating preview: %v", err)
		}
		if !shouldProceed {
			fmt.Println("Operation cancelled.")
			return
		}
	}

	// Add server block to nginx
	if err := gen.AddServerToNginx(cfg, *nginxPath, *serverType, *backup); err != nil {
		log.Fatalf("Error adding server to nginx config: %v", err)
	}

	fmt.Printf("✅ Server block added successfully to: %s\n", *nginxPath)
	fmt.Printf("📋 Server type: %s\n", *serverType)
	fmt.Printf("🌐 Server name: %s\n", cfg.ServerName)
}

func getInteractiveConfig(serverType string) (*config.ServerConfig, error) {
	reader := bufio.NewReader(os.Stdin)
	cfg := &config.ServerConfig{}

	fmt.Println("🔧 Interactive Configuration Mode")
	fmt.Println("=" + strings.Repeat("=", 40))

	// Get server name
	fmt.Print("Enter server name (e.g., example.com): ")
	serverName, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	cfg.ServerName = strings.TrimSpace(serverName)

	// Get listen port
	fmt.Print("Enter listen port [80]: ")
	listen, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	listen = strings.TrimSpace(listen)
	if listen == "" {
		listen = "80"
	}
	cfg.Listen = listen

	// Get type-specific configuration
	switch serverType {
	case "static":
		fmt.Print("Enter document root (e.g., /var/www/html): ")
		root, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		cfg.Root = strings.TrimSpace(root)

		fmt.Print("Enter index file [index.html]: ")
		index, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		index = strings.TrimSpace(index)
		if index == "" {
			index = "index.html"
		}
		cfg.Index = index

	case "proxy":
		fmt.Print("Enter proxy target (e.g., 8084 or http://127.0.0.1:8084): ")
		proxy, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		proxy = strings.TrimSpace(proxy)

		// Check if it's just a port number or full URL
		if strings.HasPrefix(proxy, "http://") || strings.HasPrefix(proxy, "https://") {
			cfg.ProxyPass = proxy
		} else {
			cfg.ProxyPort = proxy
		}
	}

	fmt.Println()
	return cfg, nil
}

func showPreview(gen *generator.Generator, cfg *config.ServerConfig, nginxPath, serverType string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("📋 Configuration Preview")
	fmt.Println("=" + strings.Repeat("=", 50))

	// Show current configuration
	fmt.Printf("Server Name: %s\n", cfg.ServerName)
	fmt.Printf("Listen Port: %s\n", cfg.Listen)
	fmt.Printf("Server Type: %s\n", serverType)

	if serverType == "static" {
		fmt.Printf("Document Root: %s\n", cfg.Root)
		fmt.Printf("Index File: %s\n", cfg.Index)
	} else {
		if cfg.ProxyPass != "" {
			fmt.Printf("Proxy Target: %s\n", cfg.ProxyPass)
		} else {
			fmt.Printf("Proxy Port: %s\n", cfg.ProxyPort)
		}
	}

	fmt.Println()

	// Generate and show the server block that will be added
	var serverBlock string
	switch serverType {
	case "static":
		serverBlock = gen.GenerateStaticServerBlock(cfg)
	case "proxy":
		serverBlock = gen.GenerateProxyServerBlock(cfg)
	}

	// Show preview with context
	preview, err := gen.GeneratePreview(nginxPath, serverBlock)
	if err != nil {
		return false, fmt.Errorf("failed to generate preview: %w", err)
	}

	fmt.Println("🔍 Nginx Configuration Preview")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println(preview)
	fmt.Println("=" + strings.Repeat("=", 50))

	// Ask for confirmation
	fmt.Print("Do you want to proceed with these changes? (y/N): ")
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes", nil
}

func showUsage() {
	fmt.Println("Nginx Server Manager")
	fmt.Println("Add new server blocks to existing nginx configuration")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  # File-based configuration")
	fmt.Println("  nginx-server-manager -config <config_file> -nginx <nginx_conf> -type <server_type>")
	fmt.Println()
	fmt.Println("  # Interactive configuration")
	fmt.Println("  nginx-server-manager -interactive -nginx <nginx_conf> -type <server_type>")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -config        Path to server configuration file (.json/.yaml)")
	fmt.Println("  -nginx         Path to existing nginx.conf file")
	fmt.Println("  -type          Server type:")
	fmt.Println("                   static - Static file server")
	fmt.Println("                   proxy  - Reverse proxy server")
	fmt.Println("  -interactive   Manual input mode via terminal")
	fmt.Println("  -preview       Show preview before applying changes (default: true)")
	fmt.Println("  -backup        Create backup before modifying (default: true)")
	fmt.Println("  -help          Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # File-based static server")
	fmt.Println("  nginx-server-manager -config static.json -nginx /etc/nginx/nginx.conf -type static")
	fmt.Println()
	fmt.Println("  # Interactive proxy server")
	fmt.Println("  nginx-server-manager -interactive -nginx /etc/nginx/nginx.conf -type proxy")
	fmt.Println()
	fmt.Println("  # Skip preview")
	fmt.Println("  nginx-server-manager -interactive -nginx nginx.conf -type static -preview=false")
}
