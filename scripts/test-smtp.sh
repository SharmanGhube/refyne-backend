#!/bin/bash

echo "🔍 SMTP Connectivity Diagnostic"
echo "================================"
echo ""

# 1. Test DNS resolution
echo "1️⃣  Testing DNS Resolution for smtp.gmail.com"
nslookup smtp.gmail.com 2>/dev/null || dig smtp.gmail.com 2>/dev/null || echo "❌ DNS lookup failed"
echo ""

# 2. Test TCP connection
echo "2️⃣  Testing TCP Connection to smtp.gmail.com:587"
timeout 5 bash -c 'cat < /dev/null > /dev/tcp/smtp.gmail.com/587' 2>/dev/null && echo "✅ Connection successful" || echo "❌ Connection failed (timeout or refused)"
echo ""

# 3. Show network info
echo "3️⃣  Network Configuration"
echo "Hostname: $(hostname)"
echo "DNS Servers: $(cat /etc/resolv.conf 2>/dev/null | grep nameserver | head -3)"
echo ""

# 4. Show environment variables
echo "4️⃣  Current SMTP Configuration"
echo "SMTP_HOST: ${SMTP_HOST:-not set}"
echo "SMTP_PORT: ${SMTP_PORT:-not set}"
echo "SMTP_USE_TLS: ${SMTP_USE_TLS:-not set}"
echo "SMTP_USE_SSL: ${SMTP_USE_SSL:-not set}"
echo "SMTP_USERNAME: ${SMTP_USERNAME:+***}"
echo "SMTP_PASSWORD: ${SMTP_PASSWORD:+***}"
echo ""

echo "⚠️  If connections are failing, check:"
echo "   - Railway outbound firewall rules"
echo "   - Try using a different email provider (SendGrid, Resend, etc.)"
echo "   - Verify SMTP_HOST is correctly set"
