#!/usr/bin/env bash

# prints error if a specific command does not exist
function panic_if_command_not_exists() {
  if ! command -v "$1" >>/dev/null 2>/dev/null; then
    echo -e "\033[0;33m$1\033[0m command is not available - please install it first"
    exit 1
  fi
}

# installs the golangci-lint of the specific version to the specific directory
function install_golangci_lint() {
  local version="$1"

  # check presence and version of golangci-lint
  if [[ -f third_party/golangci-lint ]]; then
    local installed_version
    installed_version=$(third_party/golangci-lint --version | grep -E -o '[0-9]+\.[0-9]+\.[0-9]+')
    if [[ $installed_version == "$version" ]]; then
      exit 0
    fi
    echo "golangci version is not as expected - got '$installed_version', want '$version' - will update"
  fi

  panic_if_command_not_exists "curl"
  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "./third_party" "v$version"
}