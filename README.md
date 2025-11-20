# 📧 Mailer - Production Ready Email CLI & Library

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

### Building the Shared Library (.so)

```bash
# Build shared library (.so for Linux/Mac)
go build -buildmode=c-shared -o libmailer.so libmailer.go

# Build static library (.a)
go build -buildmode=c-archive -o libmailer.a libmailer.go
```

**Note:** 
- `.so` = shared object (dynamic library)
- `.a` = archive (static library)
- On macOS, `.so` becomes `.dylib`
- On Windows, `.so` becomes `.dll`

## ⚙️ Configuration

Create a config file at `~/.config/mailer/config.json`:

```bash
mkdir -p ~/.config/mailer
cat > ~/.config/mailer/config.json << 'EOF'
{
  "host": "smtp.gmail.com",
  "port": 587,
  "username": "your-email@gmail.com",
  "password": "your-app-password",
  "from": "Your Name <your-email@gmail.com>"
}
EOF
```

### Gmail Setup

For Gmail, you need an **App Password**:

1. Go to Google Account → Security
2. Enable 2-Factor Authentication
3. Generate an App Password for "Mail"
4. Use that password in the config

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

## 🔧 Library Usage

### C/C++ Example

Create `test.c`:

```c
#include <stdio.h>
#include <stdlib.h>
#include "libmailer.h"

int main() {
    const char* config = "{\"host\":\"smtp.gmail.com\",\"port\":587,"
                        "\"username\":\"user@gmail.com\",\"password\":\"app-pass\","
                        "\"from\":\"Sender <user@gmail.com>\"}";
    
    const char* message = "{\"to\":\"recipient@example.com\","
                         "\"subject\":\"Test from C\","
                         "\"body\":\"This email was sent via C library\","
                         "\"isHTML\":false}";
    
    int result = SendEmail((char*)config, (char*)message);
    
    if (result != 0) {
        char* error = GetLastError();
        printf("Error: %s\n", error);
        FreeString(error);
        return 1;
    }
    
    printf("Email sent successfully!\n");
    return 0;
}
```

Compile and run:

```bash
# With shared library
gcc test.c -L. -lmailer -o test
LD_LIBRARY_PATH=. ./test

# With static library
gcc test.c libmailer.a -o test
./test
```

### Python Example (ctypes)

```python
import ctypes
import json

# Load library
lib = ctypes.CDLL('./libmailer.so')

# Define function signatures
lib.SendEmail.argtypes = [ctypes.c_char_p, ctypes.c_char_p]
lib.SendEmail.restype = ctypes.c_int
lib.GetLastError.restype = ctypes.c_char_p

# Configuration
config = {
    "host": "smtp.gmail.com",
    "port": 587,
    "username": "user@gmail.com",
    "password": "app-password",
    "from": "Sender <user@gmail.com>"
}

# Message
message = {
    "to": "recipient@example.com",
    "subject": "Test from Python",
    "body": "This email was sent via Python using the mailer library",
    "isHTML": False
}

# Send email
result = lib.SendEmail(
    json.dumps(config).encode('utf-8'),
    json.dumps(message).encode('utf-8')
)

if result != 0:
    error = lib.GetLastError().decode('utf-8')
    print(f"Error: {error}")
else:
    print("Email sent successfully!")
```

### Go Example

```go
package main

import (
    "encoding/json"
    "fmt"
)

func main() {
    config := map[string]interface{}{
        "host":     "smtp.gmail.com",
        "port":     587,
        "username": "user@gmail.com",
        "password": "app-password",
        "from":     "Sender <user@gmail.com>",
    }

    message := map[string]interface{}{
        "to":      "recipient@example.com",
        "subject": "Test from Go",
        "body":    "This email was sent via Go",
        "isHTML":  false,
    }

    configJSON, _ := json.Marshal(config)
    messageJSON, _ := json.Marshal(message)

    // Call library functions here
    fmt.Println("Config:", string(configJSON))
    fmt.Println("Message:", string(messageJSON))
}
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

## 📋 Library API Reference

### C Functions

**SendEmail(configJSON, messageJSON)** → int
- Returns 0 on success, negative on error
- Config: `{"host", "port", "username", "password", "from"}`
- Message: `{"to", "subject", "body", "cc", "bcc", "attachments", "isHTML"}`

**SendRawEML(configJSON, emlPath)** → int
- Sends a raw EML file

**SendSimpleEmail(host, port, user, pass, from, to, subject, body)** → int
- Simple interface for basic emails

**GetLastError()** → char*
- Returns last error message
- Must free with FreeString()

**FreeString(str)**
- Frees strings returned by library

## 📦 Distribution

### Creating a Release

```bash
# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o mailer-linux-amd64 main.go
GOOS=darwin GOARCH=amd64 go build -o mailer-macos-amd64 main.go
GOOS=darwin GOARCH=arm64 go build -o mailer-macos-arm64 main.go
GOOS=windows GOARCH=amd64 go build -o mailer-windows-amd64.exe main.go

# Build libraries
GOOS=linux go build -buildmode=c-shared -o libmailer-linux.so libmailer.go
GOOS=darwin go build -buildmode=c-shared -o libmailer-macos.dylib libmailer.go
```

## 📄 License & Credits

Built with Go and [gomail.v2](https://github.com/go-gomail/gomail)
