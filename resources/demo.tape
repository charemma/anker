# ikno demo - VHS tape file
# Live demo showing real ikno commands
#
# Prerequisites:
#   brew install vhs glow
#   ./demo-setup.sh (to create test data)
#
# Run with: IKNO_HOME=/tmp/ikno-demo/.config/ikno vhs demo.tape

Output resources/demo.gif

Set FontSize 18
Set Width 1600
Set Height 900
Set Theme "Catppuccin Mocha"
Set Padding 30

# Intro
Type "# ikno - reconstruct your workday"
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
Type "ikno source add git ~/code/ikno"
Sleep 500ms
Enter
Sleep 2s

Type "ikno source add markdown /tmp/ikno-demo/notes"
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
Type "ikno source list"
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
Type "ikno recap thisweek | claude -p 'Summarize my work' | glow"
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
Type "# ikno - a fixpoint for your work"
Sleep 2s
Enter
Enter
Type "# Install: curl -sSL https://charemma.de/ikno/install.sh | sh"
Sleep 2s
Enter
Enter
Type "# github.com/charemma/ikno"
Sleep 6s
