# Regenerate the README screenshots and GIF from the repo root:
#
#   vhs docs/demo.tape && ffmpeg -y -loglevel error -i docs/assets/board.png -vf crop=iw:400:0:0 docs/assets/board-crop.png && mv docs/assets/board-crop.png docs/assets/board.png \
#     && ffmpeg -y -loglevel error -i docs/assets/help.png -vf crop=500:615:0:0 docs/assets/help-crop.png && mv docs/assets/help-crop.png docs/assets/help.png
#
# The crops drop the empty space below the Cards and beside the help; the frame stays tall enough for the framed help.
#
# It builds a throwaway binary stamped with the latest tag, so the help shows a clean version.
# The Board runs standalone on the fictitious project in docs/demo. herdr is replaced by
# `false` and state/config go to fresh temp dirs, so nothing from a real herdr session shows.

Output docs/assets/demo.gif

Set Shell "bash"
Set FontSize 16
Set Width 1600
Set Height 700
Set Padding 20
Set Theme "Catppuccin Mocha"
Set Framerate 12
Set TypingSpeed 120ms

Hide
Type "export PS1='$ ' HERDR_BIN_PATH=/usr/bin/false EDITOR=true KANBAN_DEMO_TMP=$(mktemp -d)" Enter
Type "export HERDR_PLUGIN_STATE_DIR=$KANBAN_DEMO_TMP/state HERDR_PLUGIN_CONFIG_DIR=$KANBAN_DEMO_TMP/config" Enter
Type "mkdir -p $HERDR_PLUGIN_STATE_DIR $HERDR_PLUGIN_CONFIG_DIR" Enter
Type `go build -ldflags "-X main.version=$(git describe --tags --abbrev=0)" -o $KANBAN_DEMO_TMP/kanban ./cmd/kanban` Enter
Type "cd docs/demo && export KANBAN_PROJECT=$PWD" Enter
Type "clear && $KANBAN_DEMO_TMP/kanban board; rm -rf $KANBAN_DEMO_TMP" Enter
Wait+Screen@60s /Archived/
Sleep 1s
# select the first Card in In Progress
Type "3"
Sleep 300ms
Show

Sleep 1.1s
Screenshot docs/assets/board.png

# navigate
Type "k"
Sleep 400ms
Type "j"
Sleep 400ms
Type "h"
Sleep 400ms
Type "l"
Sleep 400ms
Type "l"
Sleep 400ms
Type "h"
Sleep 400ms
Type "G"
Sleep 600ms

# milestone filter
Type "/m2"
Sleep 1s
Enter
Sleep 1.2s

# detail of the selected In Progress Change, Tasks tab
Enter
Sleep 1s
Type "1"
Sleep 1.2s
Screenshot docs/assets/detail.png
Tab
Sleep 1.1s
Tab
Sleep 1.1s
Escape
Sleep 600ms
Escape
Sleep 600ms

# help
Type "?"
Sleep 1.5s
Screenshot docs/assets/help.png
Sleep 500ms

Hide
Type "q"
Sleep 300ms
Type "q"

