#!/bin/bash
INPUT=$1
FILE=$(echo "$INPUT" | cut -d: -f1)
LINE=$(echo "$INPUT" | cut -d: -f2)

bat --style=numbers,changes \
    --color=always \
    --highlight-line "$LINE" \
    --line-range "$((LINE - 10)):$((LINE + 10))" \
    "$FILE"
