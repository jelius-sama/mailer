package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jelius-sama/logger"
	"io"
	"net/mail"
	"os"
	"path/filepath"
	"strings"

	gomail "gopkg.in/gomail.v2"
)

const (
	Version = "1.0.0"
	Banner  = `
    ███╗   ███╗ █████╗ ██╗██╗     ███╗   ███╗ █████╗ ███╗   ██╗
    ████╗ ████║██╔══██╗██║██║     ████╗ ████║██╔══██╗████╗  ██║
    ██╔████╔██║███████║██║██║     ██╔████╔██║███████║██╔██╗ ██║
    ██║╚██╔╝██║██╔══██║██║██║     ██║╚██╔╝██║██╔══██║██║╚██╗██║
    ██║ ╚═╝ ██║██║  ██║██║███████╗██║ ╚═╝ ██║██║  ██║██║ ╚████║
    ╚═╝     ╚═╝╚═╝  ╚═╝╚═╝╚══════╝╚═╝     ╚═╝╚═╝  ╚═╝╚═╝  ╚═══╝
    
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

// ParseEmailAddress handles email formats like "Name <email@domain.com>" or "email@domain.com"
func ParseEmailAddress(addr string) (string, error) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return "", fmt.Errorf("empty email address")
	}

	// Try parsing as RFC 5322 address
	parsed, err := mail.ParseAddress(addr)
	if err != nil {
		// If parsing fails, check if it's a simple email
		if strings.Contains(addr, "@") && !strings.Contains(addr, "<") {
			return addr, nil
		}
		return "", fmt.Errorf("invalid email address format: %w", err)
	}

	return parsed.Address, nil
}

// FormatEmailAddress creates proper "Name <email>" format
func FormatEmailAddress(addr string) string {
	parsed, err := mail.ParseAddress(addr)
	if err != nil {
		return addr
	}
	return parsed.String()
}

// SendMail sends an email using provided parameters
func SendMail(smtpHost string, smtpPort int, username, password, from, to, subject, body string, cc, bcc []string, attachments []string) error {
	m := gomail.NewMessage()

	// Set From with proper formatting
	m.SetHeader("From", FormatEmailAddress(from))

	// Set To with proper formatting
	m.SetHeader("To", FormatEmailAddress(to))

	// Set CC if provided
	if len(cc) > 0 {
		formattedCC := make([]string, len(cc))
		for i, addr := range cc {
			formattedCC[i] = FormatEmailAddress(addr)
		}
		m.SetHeader("Cc", formattedCC...)
	}

	// Set BCC if provided
	if len(bcc) > 0 {
		formattedBCC := make([]string, len(bcc))
		for i, addr := range bcc {
			formattedBCC[i] = FormatEmailAddress(addr)
		}
		m.SetHeader("Bcc", formattedBCC...)
	}

	m.SetHeader("Subject", subject)

	// Detect content type (simple check for HTML)
	if strings.Contains(body, "<html") || strings.Contains(body, "<HTML") {
		m.SetBody("text/html", body)
	} else {
		m.SetBody("text/plain", body)
	}

	// Add attachments
	for _, attachment := range attachments {
		if _, err := os.Stat(attachment); err != nil {
			return fmt.Errorf("attachment not found: %s", attachment)
		}
		m.Attach(attachment)
	}

	d := gomail.NewDialer(smtpHost, smtpPort, username, password)
	return d.DialAndSend(m)
}

// SendRawEML sends a raw .eml file
func SendRawEML(smtpHost string, smtpPort int, username, password string, emlPath string) error {
	file, err := os.Open(emlPath)
	if err != nil {
		return fmt.Errorf("cannot open EML file: %w", err)
	}
	defer file.Close()

	// Parse the EML file to extract headers and body
	msg, err := mail.ReadMessage(file)
	if err != nil {
		return fmt.Errorf("invalid EML file format: %w", err)
	}

	// Create new message
	m := gomail.NewMessage()

	// Copy headers
	for key, values := range msg.Header {
		if len(values) > 0 {
			m.SetHeader(key, values...)
		}
	}

	// Read body
	bodyBytes, err := io.ReadAll(msg.Body)
	if err != nil {
		return fmt.Errorf("cannot read EML body: %w", err)
	}

	// Detect content type from header or body
	contentType := msg.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") {
		m.SetBody("text/html", string(bodyBytes))
	} else {
		m.SetBody("text/plain", string(bodyBytes))
	}

	d := gomail.NewDialer(smtpHost, smtpPort, username, password)
	return d.DialAndSend(m)
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
		config, err = LoadConfig()
		if err != nil {
			logger.Fatal("SMTP credentials not provided and config file not found.\n" +
				"Please provide --host, --user, --pass flags or create config at ~/.config/mailer/config.json\n" +
				"Run 'mailer --help' for more information.")
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

	// // Parse CC addresses
	// var cc []string
	// if *ccAddrs != "" {
	// 	for _, addr := range strings.Split(*ccAddrs, ",") {
	// 		addr = strings.TrimSpace(addr)
	// 		if _, err := ParseEmailAddress(addr); err != nil {
	// 			logger.Fatal(fmt.Sprintf("Invalid CC email '%s': %v", addr, err))
	// 		}
	// 		cc = append(cc, addr)
	// 	}
	// }
	//
	// // Parse BCC addresses
	// var bcc []string
	// if *bccAddrs != "" {
	// 	for _, addr := range strings.Split(*bccAddrs, ",") {
	// 		addr = strings.TrimSpace(addr)
	// 		if _, err := ParseEmailAddress(addr); err != nil {
	// 			logger.Fatal(fmt.Sprintf("Invalid BCC email '%s': %v", addr, err))
	// 		}
	// 		bcc = append(bcc, addr)
	// 	}
	// }

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
