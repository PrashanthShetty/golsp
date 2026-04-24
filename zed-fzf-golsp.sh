#!/bin/bash
set -e
# Use this bash script in Zed editor's task.json to find all references and Implementations.
# copy golsp-preview.sh and zed-fzf-golsp.sh script files to ~/.config/zed
# make the files executable
# chmod +x zed-fzf-golsp.sh
# chmod +x golsp-preview.sh
# Example
#  {
#     "label": "Find References",
#     "command": "zed \"$(~/.config/zed/zed-fzf-golsp.sh references $ZED_FILE $ZED_ROW $ZED_COLUMN $ZED_WORKTREE_ROOT)\"",
#     "tags": ["go"],
#     "allow_concurrent_runs": false,
#     "hide": "always",
#     "use_new_terminal": false,
#     "cwd": "$ZED_WORKTREE_ROOT",
# },
# {
#     "label": "Find Implementations",
#     "command": "zed \"$(~/.config/zed/zed-fzf-golsp.sh implementations $ZED_FILE $ZED_ROW $ZED_COLUMN $ZED_WORKTREE_ROOT)\"",
#     "tags": ["go"],
#     "allow_concurrent_runs": false,
#     "hide": "always",
#     "use_new_terminal": false,
#     "cwd": "$ZED_WORKTREE_ROOT",
# },
# Add Keybaord shortcut in keymap.json
# "cmd-shift-r": [
#     "task::Spawn",
#     {
#         "task_name": "Find References",
#         "reveal_target": "center",
#     },
# ],

COMMAND="${1}"
FILE="${2}"
LINE="${3}"
COL="${4}"
ROOT="${5}"

if [[ -z "$FILE" || -z "$LINE" || -z "$COL" || -z "$ROOT" ]]; then
    echo "Usage: zed-fzf-golsp.sh  <file> <line> <col> <root>"
    exit 1
fi

cd "$ROOT"

fzf \
    --delimiter=$'\t' \
    --with-nth=1 \
    --nth=1 \
    --layout=reverse \
    --border \
    --pointer='' \
    --no-scrollbar \
    --info=inline-right \
    --height=100% \
    --border=rounded \
    --prompt '⏳ ' \
    --header-first \
    --ansi \
    --disabled \
    --preview "$HOME/.config/zed/golsp-preview.sh {2}" \
    --preview-window "border-left,right,70%,+{2}+3/3" \
    --bind "change:ignore" \
    --bind "start:reload(golsp-cli \"$COMMAND\" \"$FILE\" $LINE $COL \"$ROOT\" \
      | sed 's/-[0-9]*$//' \
      | awk -F: -v root=\"$ROOT\" '{rel=\$1; sub(root \"/\", \"\", rel); n=split(rel, a, \"/\"); dir=\"\"; for(i=1;i<n;i++) dir=dir a[i] \"/\"; file=a[n]; highlight=\"\033[1;34m\" file \"\033[0m\"; print dir highlight\":\"\$2\":\"\$3\"\t\"\$0}')" \
    --bind "load:enable-search+unbind(change)+change-prompt(🔍 )" |
    cut -f2
