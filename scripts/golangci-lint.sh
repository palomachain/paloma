#!/bin/bash
set -euo pipefail

# installs the golangci-lint of the specific version to the specific directory
function install_golangci_lint() {
  local version="$1"
  local project_dir="$2"

  # check presence and version of golangci-lint
  if [[ -f "$project_dir/third_party/golangci-lint" ]]; then
    local installed_version
    installed_version=$("$project_dir/third_party/golangci-lint" --version | grep -E -o '[0-9]+\.[0-9]+\.[0-9]+')
    if [[ $installed_version == "$version" ]]; then
      exit 0
    fi
    echo "golangci version is not as expected - got '$installed_version', want '$version' - will update"
  fi

  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$project_dir/third_party" "v$version"
}
