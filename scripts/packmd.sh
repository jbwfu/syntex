#!/bin/bash

if [ $# -ne 1 ]; then
    echo "Usage: $0 <directory_or_file>"
    exit 1
fi

TARGET="$1"

get_language() {
    case "$1" in
    *.go) echo "go" ;;
    *.el) echo "elisp" ;;
    *.org) echo "org" ;;
    *.py) echo "python" ;;
    *.js) echo "javascript" ;;
    *.java) echo "java" ;;
    *.c) echo "c" ;;
    *.cpp) echo "cpp" ;;
    *.sh) echo "bash" ;;
    *.html) echo "html" ;;
    *.css) echo "css" ;;
    *.json) echo "json" ;;
    *.xml) echo "xml" ;;
    *.md) echo "markdown" ;;
    *) echo "text" ;;
    esac
}

process_file() {
    local FILE="$1"
    local LANGUAGE

    LANGUAGE="$(get_language "$FILE")"
    echo "- ${FILE}"
    echo '```'"$LANGUAGE"
    cat "$FILE"
    echo '```'
    echo
}

if [ -d "$TARGET" ]; then
    for FILE in "$TARGET"/*; do
        if [ -f "$FILE" ]; then
            process_file "$FILE"
        fi
    done
elif [ -f "$TARGET" ]; then
    process_file "$TARGET"
else
    echo "Error: '$TARGET' is not a valid file or directory."
    exit 1
fi
