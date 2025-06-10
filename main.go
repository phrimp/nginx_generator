package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"nginx_tool/internal/config"
	"nginx_tool/internal/generator"
	"os"
	"os/exec"
	"strings"
)

func main() {
	var (
		configPath  = flag.String("config", "", "Path to server configuration JSON/YAML file")
		nginxPath   = flag.String("nginx", "", "Path to existing nginx.conf file (auto-detected if not specified)")
		serverType  = flag.String("type", "static", "Server type: 'static' or 'proxy'")
		interactive = flag.Bool("interactive", false, "Manual input mode via terminal")
		preview     = flag.Bool("preview", true, "Show preview before applying changes")
		backup      = flag.Bool("backup", true, "Create backup of nginx.conf before modification")
		autoDetect  = flag.Bool("auto-detect", true, "Auto-detect nginx configuration file")
		help        = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		showUsage()
		return
	}

	if *nginxPath == "" && *autoDetect {
		detectedPath, err := detectNginxConfig()
		if err != nil {
			log.Printf("Warning: Could not auto-detect nginx config: %v", err)
			log.Fatal("Error: nginx path is required. Use -nginx flag to specify manually.")
		}
		*nginxPath = detectedPath
		fmt.Printf("🔍 Auto-detected nginx config: %s\n", *nginxPath)
	} else if *nginxPath == "" {
		log.Fatal("Error: nginx path is required when auto-detection is disabled")
	}

	var cfg *config.ServerConfig
	var err error

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

	if *serverType != "static" && *serverType != "proxy" {
		log.Fatal("Error: type must be either 'static' or 'proxy'")
	}

	gen := generator.New()

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

	if err := gen.AddServerToNginx(cfg, *nginxPath, *serverType, *backup); err != nil {
		log.Fatalf("Error adding server to nginx config: %v", err)
	}

	fmt.Printf("✅ Server block added successfully to: %s\n", *nginxPath)
	fmt.Printf("📋 Server type: %s\n", *serverType)
	fmt.Printf("🌐 Server name: %s\n", cfg.ServerName)
}

func detectNginxConfig() (string, error) {
	fmt.Println("🔍 Auto-detecting nginx configuration...")

	commonPaths := []string{
		"/etc/nginx/nginx.conf",
		"/usr/local/etc/nginx/nginx.conf",
		"/usr/local/nginx/conf/nginx.conf",
		"/opt/nginx/conf/nginx.conf",
		"/etc/nginx.conf",
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			if isValidNginxConfig(path) {
				return path, nil
			}
		}
	}

	if nginxBinary, err := findNginxBinary(); err == nil {
		if configPath, err := getNginxConfigFromBinary(nginxBinary); err == nil {
			return configPath, nil
		}
	}

	if configPath, err := getNginxConfigFromProcess(); err == nil {
		return configPath, nil
	}

	return "", fmt.Errorf("no nginx configuration file found")
}

func findNginxBinary() (string, error) {
	commonBinPaths := []string{
		"/usr/sbin/nginx",
		"/usr/bin/nginx",
		"/usr/local/sbin/nginx",
		"/usr/local/bin/nginx",
		"/opt/nginx/sbin/nginx",
		"/sbin/nginx",
		"/bin/nginx",
	}

	for _, path := range commonBinPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	if path, err := exec.LookPath("nginx"); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("nginx binary not found")
}

func getNginxConfigFromBinary(nginxBinary string) (string, error) {
	cmd := exec.Command(nginxBinary, "-t")
	output, err := cmd.CombinedOutput()
	if err != nil {
		cmd = exec.Command(nginxBinary, "-T")
		output, err = cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("failed to get config from nginx binary: %v", err)
		}
	}

	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")

	for _, line := range lines {
		if strings.Contains(line, "configuration file") && strings.Contains(line, "nginx.conf") {
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasSuffix(part, "nginx.conf") {
					if _, err := os.Stat(part); err == nil {
						return part, nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("could not extract config path from nginx binary output")
}

func getNginxConfigFromProcess() (string, error) {
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run ps command: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "nginx: master process") {
			fields := strings.Fields(line)
			for i, field := range fields {
				if field == "-c" && i+1 < len(fields) {
					configPath := fields[i+1]
					if _, err := os.Stat(configPath); err == nil {
						return configPath, nil
					}
				}
				if strings.HasSuffix(field, "nginx.conf") {
					if _, err := os.Stat(field); err == nil {
						return field, nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("could not find nginx config from running process")
}

func isValidNginxConfig(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return false
	}

	content := string(data)
	nginxKeywords := []string{"http", "server", "location", "events"}

	keywordCount := 0
	for _, keyword := range nginxKeywords {
		if strings.Contains(content, keyword) {
			keywordCount++
		}
	}

	return keywordCount >= 2
}

func getInteractiveConfig(serverType string) (*config.ServerConfig, error) {
	reader := bufio.NewReader(os.Stdin)
	cfg := &config.ServerConfig{}

	fmt.Println("🔧 Interactive Configuration Mode")
	fmt.Println("=" + strings.Repeat("=", 40))

	fmt.Print("Enter server name (e.g., example.com): ")
	serverName, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	cfg.ServerName = strings.TrimSpace(serverName)

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

	var serverBlock string
	switch serverType {
	case "static":
		serverBlock = gen.GenerateStaticServerBlock(cfg)
	case "proxy":
		serverBlock = gen.GenerateProxyServerBlock(cfg)
	}

	preview, err := gen.GeneratePreview(nginxPath, serverBlock)
	if err != nil {
		return false, fmt.Errorf("failed to generate preview: %w", err)
	}

	fmt.Println("🔍 Nginx Configuration Preview")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println(preview)
	fmt.Println("=" + strings.Repeat("=", 50))

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
	fmt.Println("  # Auto-detect nginx config (Linux)")
	fmt.Println("  nginx-server-manager -interactive -type <server_type>")
	fmt.Println()
	fmt.Println("  # File-based configuration")
	fmt.Println("  nginx-server-manager -config <config_file> -nginx <nginx_conf> -type <server_type>")
	fmt.Println()
	fmt.Println("  # Interactive configuration")
	fmt.Println("  nginx-server-manager -interactive -nginx <nginx_conf> -type <server_type>")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -config        Path to server configuration file (.json/.yaml)")
	fmt.Println("  -nginx         Path to existing nginx.conf file (auto-detected if not specified)")
	fmt.Println("  -type          Server type:")
	fmt.Println("                   static - Static file server")
	fmt.Println("                   proxy  - Reverse proxy server")
	fmt.Println("  -interactive   Manual input mode via terminal")
	fmt.Println("  -auto-detect   Auto-detect nginx configuration file (default: true)")
	fmt.Println("  -preview       Show preview before applying changes (default: true)")
	fmt.Println("  -backup        Create backup before modifying (default: true)")
	fmt.Println("  -help          Show this help message")
	fmt.Println()
	fmt.Println("Auto-Detection (Linux):")
	fmt.Println("  The tool automatically searches for nginx.conf in common locations:")
	fmt.Println("    • /etc/nginx/nginx.conf")
	fmt.Println("    • /usr/local/etc/nginx/nginx.conf")
	fmt.Println("    • /usr/local/nginx/conf/nginx.conf")
	fmt.Println("    • /opt/nginx/conf/nginx.conf")
	fmt.Println("  Also attempts to:")
	fmt.Println("    • Find nginx binary and extract config path")
	fmt.Println("    • Detect config from running nginx process")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Auto-detect and interactive mode")
	fmt.Println("  nginx-server-manager -interactive -type static")
	fmt.Println()
	fmt.Println("  # Auto-detect with config file")
	fmt.Println("  nginx-server-manager -config static.json -type static")
	fmt.Println()
	fmt.Println("  # Manual nginx path")
	fmt.Println("  nginx-server-manager -interactive -nginx /custom/path/nginx.conf -type proxy")
	fmt.Println()
	fmt.Println("  # Disable auto-detection")
	fmt.Println("  nginx-server-manager -config config.json -nginx nginx.conf -type static -auto-detect=false")
}
