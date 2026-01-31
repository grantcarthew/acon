#!/usr/bin/env bash
#
# Terminal Logging Library
# Provides colored output and formatting functions for shell scripts
#
# Usage:
#   source "$(dirname "$0")/lib.sh"
#

# Color codes
declare -gr COLOR_RESET='\033[0m'
declare -gr COLOR_BOLD='\033[1m'
declare -gr COLOR_RED='\033[0;31m'
declare -gr COLOR_GREEN='\033[0;32m'
declare -gr COLOR_YELLOW='\033[0;33m'
declare -gr COLOR_BLUE='\033[0;34m'
declare -gr COLOR_CYAN='\033[0;36m'
declare -gr COLOR_WHITE='\033[0;37m'
declare -gr COLOR_BOLD_RED='\033[1;31m'
declare -gr COLOR_BOLD_GREEN='\033[1;32m'
declare -gr COLOR_BOLD_YELLOW='\033[1;33m'
declare -gr COLOR_BOLD_BLUE='\033[1;34m'
declare -gr COLOR_BOLD_CYAN='\033[1;36m'

# Log a title (large, bold, cyan)
log_title() {
    echo -e "${COLOR_BOLD_CYAN}$*${COLOR_RESET}"
}

# Log a heading (bold, blue)
log_heading() {
    echo -e "${COLOR_BOLD_BLUE}$*${COLOR_RESET}"
}

# Log a subheading (bold, white)
log_subheading() {
    echo -e "${COLOR_BOLD}$*${COLOR_RESET}"
}

# Log a standard message
log_message() {
    echo -e "$*"
}

# Log a newline
log_newline() {
    echo
}

# Log a success message (green checkmark)
log_success() {
    echo -e "${COLOR_GREEN}✓${COLOR_RESET} $*"
}

# Log a failure message (red X)
log_failure() {
    echo -e "${COLOR_RED}✗${COLOR_RESET} $*"
}

# Log a warning message (yellow warning sign)
log_warning() {
    echo -e "${COLOR_YELLOW}⚠${COLOR_RESET} $*"
}

# Log an error message (bold red)
log_error() {
    echo -e "${COLOR_BOLD_RED}ERROR:${COLOR_RESET} $*" >&2
}

# Log a done message (green, bold)
log_done() {
    echo -e "${COLOR_BOLD_GREEN}$*${COLOR_RESET}"
}

# Log a line separator
# Usage: log_line "=" 80
log_line() {
    local char="${1:-=}"
    local length="${2:-80}"
    printf '%*s\n' "${length}" '' | tr ' ' "${char}"
}
