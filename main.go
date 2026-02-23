package main

import (
    "flag"
    "fmt"
    libmailer "github.com/jelius-sama/libmailer/api"
    "github.com/jelius-sama/logger"
    "strings"
)

const (
    Version = "1.1.1"
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

func init() {
    logger.Configure(logger.Cnf{
        IsDev: logger.IsDev{
            DirectValue: logger.BoolPtr(false),
        },
        UseSyslog: false,
    })
}

func showHelp() {
    fmt.Printf(Banner, Version)
    fmt.Println("\n📚 USAGE:")
    fmt.Println("  mailer [OPTIONS]")
    fmt.Println("\n🔧 AWS SES OPTIONS:")
    fmt.Println("  --region    AWS region for SES (e.g., us-east-1)")
    fmt.Println("  --from      Sender email address (supports 'Name <email>' format)")
    fmt.Println("  --dualstack Use dual-stack (IPv4+IPv6) SES endpoint (default: false)")
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
    fmt.Println("  Store default AWS SES settings in: ~/.config/mailer/config.aws.json")
    fmt.Println("  Or specify a custom config path with: -c /path/to/config.aws.json")
    fmt.Println("  AWS credentials are NOT stored in the config file.")
    fmt.Println("  They are resolved automatically via the AWS credential chain:")
    fmt.Println("    - AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY environment variables")
    fmt.Println("    - ~/.aws/credentials")
    fmt.Println("    - EC2 / ECS instance or task IAM role")
    fmt.Println("  Example config:")
    fmt.Println(`  {
    "from": "Your Name <user@example.com>",
    "region": "us-east-1",
    "use_dual_stack": true
  }`)
    fmt.Println("\n📖 EXAMPLES:")
    fmt.Println("  # Send simple email:")
    fmt.Println(`  mailer --to "recipient@example.com" --subject "Hello" --body "Test message"`)
    fmt.Println("\n  # Send with custom config:")
    fmt.Println(`  mailer -c /etc/mailer/work.aws.json --to "user@example.com" --subject "Report"`)
    fmt.Println("\n  # Send with attachments:")
    fmt.Println(`  mailer --to "user@example.com" --subject "Report" --body "See attached" --attach report.pdf,data.csv`)
    fmt.Println("\n  # Send raw EML file:")
    fmt.Println(`  mailer --eml message.eml`)
    fmt.Println("\n  # Send with CC and BCC:")
    fmt.Println(`  mailer --to "user@example.com" --cc "boss@example.com" --bcc "archive@example.com" --subject "Update"`)
    fmt.Println("\n  # Force dual-stack (IPv6) endpoint via flag:")
    fmt.Println(`  mailer --dualstack --to "user@example.com" --subject "Hello" --body "Hi"`)
    fmt.Println()
}

func main() {
    // Define flags
    showHelpFlag := flag.Bool("help", false, "Show help")
    showVersion := flag.Bool("version", false, "Show version")

    configPath := flag.String("c", "", "Path to config file")

    // AWS SES flags — replace the old SMTP host/port/user/pass flags
    region := flag.String("region", "", "AWS region for SES (e.g., us-east-1)")
    from := flag.String("from", "", "Sender email address")
    dualStack := flag.Bool("dualstack", false, "Use dual-stack (IPv4+IPv6) SES endpoint")

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

    // Load config when region or from are not fully specified via flags.
    // Config file provides defaults; flags always take precedence over it.
    var config *libmailer.Config
    if *region == "" || *from == "" {
        var err error

        if *configPath != "" {
            config, err = libmailer.LoadConfigFromPath(*configPath)
            if err != nil {
                logger.Fatal(fmt.Sprintf("Failed to load config from %s: %v", *configPath, err))
            }
        } else {
            config, err = libmailer.LoadConfig()
            if err != nil {
                logger.Fatal("AWS SES configuration not provided and config file not found.\n" +
                    "Please provide --region and --from flags, or create a config at ~/.config/mailer/config.aws.json\n" +
                    "Run 'mailer --help' for more information.")
            }
        }
    }

    // Merge config with flags (flags take precedence)
    if config != nil {
        if *region == "" {
            *region = config.Region
        }
        if *from == "" {
            *from = config.From
        }
        // Only apply config's dual-stack if the flag was not explicitly set.
        // Since false is the zero value we can't distinguish "not set" from
        // "explicitly set to false" with the standard flag package, so we
        // only promote the config value when the flag is still false and the
        // config wants it enabled.
        if !*dualStack && config.UseDualStack {
            *dualStack = true
        }
    }

    // Validate required AWS SES configuration
    if *region == "" {
        logger.Fatal("AWS region is required. Provide --region or set \"region\" in your config file.")
    }
    if *from == "" {
        logger.Fatal("Sender address is required. Provide --from or set \"from\" in your config file.")
    }

    // Build the resolved config that will be passed to every api call
    resolvedConfig := &libmailer.Config{
        From:         *from,
        Region:       *region,
        UseDualStack: *dualStack,
    }

    // EML mode
    if *emlFile != "" {
        logger.Info("Sending EML file:", *emlFile)
        if err := libmailer.SendRawEML(resolvedConfig, *emlFile); err != nil {
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

    // Validate email addresses
    if _, err := libmailer.ParseEmailAddress(*from); err != nil {
        logger.Fatal("Invalid --from email:", err)
    }
    if _, err := libmailer.ParseEmailAddress(*to); err != nil {
        logger.Fatal("Invalid --to email:", err)
    }

    // Parse CC addresses
    var cc []string
    if *ccAddrs != "" {
        for addr := range strings.SplitSeq(*ccAddrs, ",") {
            addr = strings.TrimSpace(addr)
            if _, err := libmailer.ParseEmailAddress(addr); err != nil {
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
            if _, err := libmailer.ParseEmailAddress(addr); err != nil {
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
    if err := libmailer.SendMail(resolvedConfig, *from, *to, *subject, *body, cc, bcc, attachments); err != nil {
        logger.Fatal("Send failed:", err)
    }

    logger.Okay("Mail sent successfully!")
}

