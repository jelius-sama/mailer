# 📧 Mailer - Email CLI

## 🚀 Quick Start

### Building the CLI

```bash
# Build the CLI executable
go build -o mailer main.go

# Install to your PATH
sudo mv mailer /usr/local/bin/

# Or for local user
mkdir -p ~/bin
mv mailer ~/bin/
# Add ~/bin to PATH in ~/.bashrc or ~/.zshrc
```
## ⚙️ Configuration

Create a config file at `~/.config/mailer/config.json`:

```bash
mkdir -p ~/.config/mailer
cat > ~/.config/mailer/config.json << 'EOF'
{
  "host": "inbound-smtp.ap-northeast-1.amazonaws.com",
  "port": 587,
  "username": "your-email@example.com",
  "password": "your-password",
  "from": "Your Name <your-email@example.com>"
}
EOF
```
## 📖 CLI Usage Examples

### Basic Email

```bash
mailer --to "recipient@example.com" \
       --subject "Hello World" \
       --body "This is a test message"
```

### With Name Format

```bash
mailer --to "John Doe <john@example.com>" \
       --subject "Meeting Tomorrow" \
       --body "Don't forget our 2pm meeting!"
```

### With CC and BCC

```bash
mailer --to "primary@example.com" \
       --cc "boss@example.com,colleague@example.com" \
       --bcc "archive@example.com" \
       --subject "Project Update" \
       --body "Please see the attached report"
```

### With Attachments

```bash
mailer --to "client@example.com" \
       --subject "Monthly Report" \
       --body "Please find attached the report" \
       --attach "report.pdf,data.xlsx,chart.png"
```

### HTML Email

```bash
mailer --to "user@example.com" \
       --subject "Welcome!" \
       --body "<html><body><h1>Welcome!</h1><p>Thanks for joining.</p></body></html>"
```

### Send Raw EML File

```bash
mailer --eml message.eml
```

### Override Config SMTP Settings

```bash
mailer --host smtp.office365.com \
       --port 587 \
       --user work@company.com \
       --pass work-password \
       --from "Work Name <work@company.com>" \
       --to "client@example.com" \
       --subject "Business Email" \
       --body "Professional correspondence"
```

## 🎯 Advanced Features

### Creating EML Files

```bash
cat > test.eml << 'EOF'
From: Sender <sender@example.com>
To: Recipient <recipient@example.com>
Subject: Test EML
Content-Type: text/plain; charset=UTF-8

This is a test email from an EML file.
EOF

mailer --eml test.eml
```

### Multiline Body (Bash)

```bash
BODY="Line 1
Line 2
Line 3"

mailer --to "user@example.com" --subject "Multiline" --body "$BODY"
```

### Reading Body from File

```bash
mailer --to "user@example.com" \
       --subject "File Content" \
       --body "$(cat message.txt)"
```

## 🔒 Security Best Practices

1. **Never commit credentials** to version control
2. **Use app-specific passwords** for Gmail/Outlook
3. **Set restrictive permissions** on config file:
   ```bash
   chmod 600 ~/.config/mailer/config.json
   ```
4. **Use environment variables** for CI/CD:
   ```bash
   mailer --host "$SMTP_HOST" \
          --user "$SMTP_USER" \
          --pass "$SMTP_PASS" \
          --to "user@example.com" \
          --subject "Automated Email" \
          --body "Build completed"
   ```

## 🧪 Testing

### Test Configuration

```bash
mailer --to "your-email@example.com" \
       --subject "Test Email" \
       --body "If you receive this, the configuration works!"
```

### Test with Different Providers

**Gmail:**
```bash
--host smtp.gmail.com --port 587
```

**Outlook/Office365:**
```bash
--host smtp.office365.com --port 587
```

**Yahoo:**
```bash
--host smtp.mail.yahoo.com --port 587
```

**Custom SMTP:**
```bash
--host smtp.yourdomain.com --port 587
```

## 🐛 Troubleshooting

### "Authentication failed"
- Check username/password
- Use app-specific password for Gmail
- Enable "Less secure app access" (not recommended)

### "Connection refused"
- Check host and port
- Verify firewall settings
- Try port 465 (SSL) or 25

### "Invalid email address"
- Ensure proper format: `user@domain.com` or `Name <user@domain.com>`
- Check for typos

### "Attachment not found"
- Use absolute paths or ensure files exist
- Check file permissions

## 📦 Distribution

### Creating a Release

```bash
# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o mailer-linux-amd64 main.go
GOOS=darwin GOARCH=amd64 go build -o mailer-macos-amd64 main.go
GOOS=darwin GOARCH=arm64 go build -o mailer-macos-arm64 main.go
```

## 📄 License & Credits

Built with Go and [gomail.v2](https://github.com/go-gomail/gomail)
