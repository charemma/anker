# anker demo - VHS tape file
# Live demo showing real anker commands
#
# Prerequisites:
#   brew install vhs glow
#   ./demo-setup.sh (to create test data)
#
# Run with: ANKER_HOME=/tmp/anker-demo/.anker vhs demo.tape

Output resources/demo.gif

Set FontSize 18
Set Width 1600
Set Height 900
Set Theme "Catppuccin Mocha"
Set Padding 30

# Intro
Type "# anker - reconstruct your workday"
Sleep 2s
Enter
Enter
Type "# Work first, summarize later. #AntiProductivity"
Enter
Sleep 3s

Hide
Enter
Type "clear"
Enter
Show

# Step 1: Add sources
Type "# Step 1: Add data sources"
Sleep 1s
Enter
Enter
Type "anker source add git ~/code/anker"
Sleep 500ms
Enter
Sleep 2s

Type "anker source add markdown /tmp/anker-demo/notes"
Sleep 500ms
Enter
Sleep 2s

Hide
Enter
Type "clear"
Enter
Show

# Step 2: List sources
Type "# Step 2: Check configured sources"
Sleep 1s
Enter
Enter
Type "anker source list"
Sleep 500ms
Enter
Sleep 3s

Hide
Enter
Type "clear"
Enter
Show


# Step 3: AI Integration example
Type "# Step 3: Generate summarized notes with AI"
Sleep 1s
Enter
Type "# (Example - requires Claude CLI)"
Enter
Enter
Type "anker recap thisweek | claude -p 'Summarize my work' | glow"
Enter
Sleep 20s
Enter
Ctrl+C

Hide
Enter
Type "clear"
Enter
Show

# Outro
Type "# anker - a fixpoint for your work"
Sleep 2s
Enter
Enter
Type "# Install: curl -sSL https://charemma.de/anker/install.sh | sh"
Sleep 2s
Enter
Enter
Type "# github.com/charemma/anker"
Sleep 6s
