package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jelius-sama/logger"
	"os"
	"path/filepath"
	"strings"
)

const (
	Version = "1.1.0"
	Banner  = `
    ███╗   ███╗ █████╗ ██╗██╗     ███████╗██████╗
    ████╗ ████║██╔══██╗██║██║     ██╔════╝██╔══██╗
    ██╔████╔██║███████║██║██║     ███████╗██████╔╝
    ██║╚██╔╝██║██╔══██║██║██║     ██╔════╝██╔══██╗
    ██║ ╚═╝ ██║██║  ██║██║███████╗███████║██║  ██║
    ╚═╝     ╚═╝╚═╝  ╚═╝╚═╝╚══════╝╚══════╝╚═╝  ╚═╝

                    📧 Email Sending Made Simple
                         Version %s
`
)

type Config struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
}

// LoadConfig attempts to load configuration from ~/.config/mailer/config.json
func LoadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".config", "mailer", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("config file not found at %s: %w", configPath, err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("invalid config file: %w", err)
	}

	return &config, nil
}

// LoadConfigFromPath loads configuration from a specific path
func LoadConfigFromPath(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("config file not found at %s: %w", configPath, err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("invalid config file: %w", err)
	}

	return &config, nil
}

func showHelp() {
	fmt.Printf(Banner, Version)
	fmt.Println("\n📚 USAGE:")
	fmt.Println("  mailer [OPTIONS]")
	fmt.Println("\n🔧 SMTP OPTIONS:")
	fmt.Println("  --host      SMTP server hostname (e.g., smtp.gmail.com)")
	fmt.Println("  --port      SMTP server port (default: 587)")
	fmt.Println("  --user      SMTP authentication username")
	fmt.Println("  --pass      SMTP authentication password")
	fmt.Println("  --from      Sender email address (supports 'Name <email>' format)")
	fmt.Println("  -c          Path to custom config file (takes priority over default)")
	fmt.Println("\n📧 EMAIL OPTIONS:")
	fmt.Println("  --to        Recipient email address (required)")
	fmt.Println("  --cc        CC recipients (comma-separated)")
	fmt.Println("  --bcc       BCC recipients (comma-separated)")
	fmt.Println("  --subject   Email subject line")
	fmt.Println("  --body      Email body content")
	fmt.Println("  --attach    Attachment file paths (comma-separated)")
	fmt.Println("\n📄 RAW EML MODE:")
	fmt.Println("  --eml       Path to .eml file to send directly")
	fmt.Println("\n⚙️  SYSTEM:")
	fmt.Println("  --help      Show this help message")
	fmt.Println("  --version   Show version information")
	fmt.Println("\n💡 CONFIGURATION:")
	fmt.Println("  Store default SMTP settings in: ~/.config/mailer/config.json")
	fmt.Println("  Or specify custom config with: -c /path/to/config.json")
	fmt.Println("  Example config:")
	fmt.Println(`  {
    "host": "smtp.gmail.com",
    "port": 587,
    "username": "user@example.com",
    "password": "your-password",
    "from": "Your Name <user@example.com>"
  }`)
	fmt.Println("\n📖 EXAMPLES:")
	fmt.Println("  # Send simple email:")
	fmt.Println(`  mailer --to "recipient@example.com" --subject "Hello" --body "Test message"`)
	fmt.Println("\n  # Send with custom config:")
	fmt.Println(`  mailer -c /etc/mailer/work.json --to "user@example.com" --subject "Report"`)
	fmt.Println("\n  # Send with attachments:")
	fmt.Println(`  mailer --to "user@example.com" --subject "Report" --body "See attached" --attach report.pdf,data.csv`)
	fmt.Println("\n  # Send raw EML file:")
	fmt.Println(`  mailer --eml message.eml`)
	fmt.Println("\n  # Send with CC and BCC:")
	fmt.Println(`  mailer --to "user@example.com" --cc "boss@example.com" --bcc "archive@example.com" --subject "Update"`)
	fmt.Println()
}

func main() {
	// Define flags
	showHelpFlag := flag.Bool("help", false, "Show help")
	showVersion := flag.Bool("version", false, "Show version")

	configPath := flag.String("c", "", "Path to config file")

	smtpHost := flag.String("host", "", "SMTP server host")
	smtpPort := flag.Int("port", 0, "SMTP server port")
	username := flag.String("user", "", "SMTP username")
	password := flag.String("pass", "", "SMTP password")
	from := flag.String("from", "", "From email address")

	to := flag.String("to", "", "Recipient email address")
	ccAddrs := flag.String("cc", "", "CC recipients (comma-separated)")
	bccAddrs := flag.String("bcc", "", "BCC recipients (comma-separated)")
	subject := flag.String("subject", "", "Email subject")
	body := flag.String("body", "", "Email body")
	attachStr := flag.String("attach", "", "Attachments (comma-separated)")

	emlFile := flag.String("eml", "", "Path to raw EML file")

	flag.Parse()

	// Show help
	if *showHelpFlag {
		showHelp()
		return
	}

	// Show version
	if *showVersion {
		fmt.Printf("mailer version %s\n", Version)
		return
	}

	// Load config if SMTP details not provided
	var config *Config
	if *smtpHost == "" || *username == "" {
		var err error

		// Prioritize custom config path if provided
		if *configPath != "" {
			config, err = LoadConfigFromPath(*configPath)
			if err != nil {
				logger.Fatal(fmt.Sprintf("Failed to load config from %s: %v", *configPath, err))
			}
		} else {
			// Fall back to default config location
			config, err = LoadConfig()
			if err != nil {
				logger.Fatal("SMTP credentials not provided and config file not found.\n" +
					"Please provide --host, --user, --pass flags or create config at ~/.config/mailer/config.json\n" +
					"Run 'mailer --help' for more information.")
			}
		}
	}

	// Merge config with flags (flags take precedence)
	if config != nil {
		if *smtpHost == "" {
			*smtpHost = config.Host
		}
		if *smtpPort == 0 {
			*smtpPort = config.Port
		}
		if *username == "" {
			*username = config.Username
		}
		if *password == "" {
			*password = config.Password
		}
		if *from == "" {
			*from = config.From
		}
	}

	// Set default port if still not set
	if *smtpPort == 0 {
		*smtpPort = 587
	}

	// Validate SMTP configuration
	if *smtpHost == "" || *username == "" || *password == "" {
		logger.Fatal("Missing SMTP configuration. Use --help for usage information.")
	}

	// EML mode
	if *emlFile != "" {
		logger.Info("Sending EML file:", *emlFile)
		if err := SendRawEML(*smtpHost, *smtpPort, *username, *password, *emlFile); err != nil {
			logger.Fatal("EML send failed:", err)
		}
		logger.Okay("EML sent successfully!")
		return
	}

	// Normal mode validation
	if *to == "" {
		logger.Fatal("Recipient email (--to) is required. Use --help for usage information.")
	}

	if *subject == "" || *body == "" {
		logger.Fatal("Both --subject and --body are required. Use --help for usage information.")
	}

	// Validate and parse email addresses
	if _, err := ParseEmailAddress(*from); err != nil {
		logger.Fatal("Invalid --from email:", err)
	}

	if _, err := ParseEmailAddress(*to); err != nil {
		logger.Fatal("Invalid --to email:", err)
	}

	// Parse CC addresses
	var cc []string
	if *ccAddrs != "" {
		for addr := range strings.SplitSeq(*ccAddrs, ",") {
			addr = strings.TrimSpace(addr)
			if _, err := ParseEmailAddress(addr); err != nil {
				// Assume logger and fmt are imported
				logger.Fatal(fmt.Sprintf("Invalid CC email '%s': %v", addr, err))
			}
			cc = append(cc, addr)
		}
	}

	// Parse BCC addresses
	var bcc []string
	if *bccAddrs != "" {
		for addr := range strings.SplitSeq(*bccAddrs, ",") {
			addr = strings.TrimSpace(addr)
			if _, err := ParseEmailAddress(addr); err != nil {
				logger.Fatal(fmt.Sprintf("Invalid BCC email '%s': %v", addr, err))
			}
			bcc = append(bcc, addr)
		}
	}

	// Parse attachments
	var attachments []string
	if *attachStr != "" {
		attachments = strings.Split(*attachStr, ",")
		for i := range attachments {
			attachments[i] = strings.TrimSpace(attachments[i])
		}
	}

	// Send email
	logger.Info("Sending email to", *to+"...")
	err := SendMail(*smtpHost, *smtpPort, *username, *password, *from, *to, *subject, *body, cc, bcc, attachments)
	if err != nil {
		logger.Fatal("Send failed:", err)
	}

	logger.Okay("Mail sent successfully!")
}
