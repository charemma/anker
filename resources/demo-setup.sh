#!/bin/bash
set -e

echo "Setting up anker demo environment..."

# Clean previous demo
rm -rf /tmp/anker-demo
mkdir -p /tmp/anker-demo/notes

# Get today's date
TODAY=$(date +%Y-%m-%d)
TODAY_DISPLAY=$(date "+%B %d, %Y")

# Create realistic work notes for today
cat > /tmp/anker-demo/notes/${TODAY}.md << EOF
# HTB Machine: Boardlight

Target: 10.10.11.23

## Initial Scan

\`\`\`
nmap -sC -sV -oA boardlight 10.10.11.23
PORT   STATE SERVICE VERSION
22/tcp open  ssh     OpenSSH 8.2p1 Ubuntu
80/tcp open  http    Apache httpd 2.4.41
\`\`\`

Interesting: only SSH and HTTP. Web server running Apache.

## Web Enumeration

Visited http://10.10.11.23 - corporate website for some board game company. Static HTML, nothing obvious.

**Findings:**
- Contact form at /contact.php (potential SQLi?)
- Virtual host hint in footer: "board.htb"
- Added to /etc/hosts

### Subdomain fuzzing
\`\`\`
ffuf -w subdomains.txt -u http://FUZZ.board.htb

crm.board.htb - 200 OK
\`\`\`

Bingo! CRM subdomain found.

## CRM Investigation

http://crm.board.htb redirects to Dolibarr login page
- Dolibarr version 17.0.0 (footer)
- Default creds don't work (admin/admin, admin/dolibarr)

Quick searchsploit:
\`\`\`
Dolibarr 17.0.0 - CVE-2023-30253 - PHP Code Injection
\`\`\`

Needs auth though. Need to find creds first.

## TODOs
- [ ] Check for password reset functionality
- [ ] Try SQL injection on contact form
- [ ] Look for backup files (.bak, .sql, .old)
- [ ] Directory bruteforce on CRM
- [ ] Check dolibarr documentation for default paths

## Directory Scan on CRM

\`\`\`
gobuster dir -u http://crm.board.htb -w common.txt
/admin      - 302 (redirect to login)
/install    - 403 (forbidden, but exists!)
/conf       - 403
\`\`\`

/install endpoint is interesting. Might have leftover files?

## Random thought

Maybe check robots.txt? Always forget that one.
Nothing interesting there.

## Breakthrough: Found leaked credentials

Checked page source more carefully. Found HTML comment:
\`\`\`html
<!-- TODO: remove test account admin@board.htb / D0l1barr2023! -->
\`\`\`

DEVELOPERS. Never change. Trying now...

## Access Gained

Logged into Dolibarr CRM with leaked creds!
- Admin panel access ✓
- Can create websites/pages
- PHP code execution via CVE-2023-30253 should work now

Next: Get reverse shell via PHP injection

## Exploitation

Created malicious website in Dolibarr with PHP payload.
Got shell as www-data!

\`\`\`bash
www-data@boardlight:/var/www/html$
\`\`\`

Now looking for user flag and privilege escalation vectors.

## Progress Today
- Initial recon ✓
- Found subdomain ✓
- Discovered CVE ✓
- Found leaked creds ✓
- Got initial foothold ✓
- Next: privesc to user → root

Time to enumerate for user creds and find the flags.
EOF

# Set file timestamp to today
touch /tmp/anker-demo/notes/${TODAY}.md

echo ""
echo "✓ Demo environment ready!"
echo ""
echo "Demo data:"
echo "  - Notes: /tmp/anker-demo/notes/${TODAY}.md"
echo "  - Git repo: ~/code/anker (will be used as source)"
echo ""
echo "Next steps:"
echo "  1. Review/edit demo.tape if needed"
echo "  2. Run: just demo-gif"
