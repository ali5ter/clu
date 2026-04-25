#!/usr/bin/env bash
#
# @file run_vhs.sh
# @brief Generate clu demo animations using vhs
# @author Alister Lewis-Bowen <alister@lewis-bowen.org>
# @version 1.0.0
# @date 2026-04-25
# @usage ./run_vhs.sh [tape-file]
#   With no argument, renders all *.tape files in the current directory.
#   Pass a specific tape file to render just that one.
# @dependencies vhs — https://github.com/charmbracelet/vhs
# @exit_codes 0 success, 1 vhs not installed

tape=${1:-"all"}

type vhs &>/dev/null || {
  echo "vhs is not installed. Refer to https://github.com/charmbracelet/vhs for installation instructions."
  exit 1
}

# https://github.com/charmbracelet/vhs/issues/419
unset PROMPT_COMMAND

if [[ "$tape" != "all" ]]; then
  vhs "$tape"
else
  for tape in *.tape; do
    # Skipe sourced vhs configuration file
    [[ "$tape" == "config.tape" ]] && continue
    vhs "$tape"
  done
fi