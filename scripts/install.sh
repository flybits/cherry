#!/usr/bin/env sh

#
# This script will install the latest version of Cherry.
#
# USAGE:
#   ./install.sh
#

# -o pipefail will cause the grep and head pipes exit with 141 code.
# https://stackoverflow.com/questions/19120263/why-exit-code-141-with-grep-q
set -eu


get_latest_release() {
  os="$1"
  arch="$2"
  release_url="https://github.com/flybits/cherry/releases"
  bin_pattern="/flybits/cherry/releases/download/v0.4.4/cherry-$os-$arch"

  if hash curl 2>/dev/null; then
    content=$(curl -sL $release_url)
  elif hash wget 2>/dev/null; then
    content=$(wget -qO- $release_url)
  else
    echo "No command available to get $release_url"
    exit 1
  fi

  bin_path=$(echo "$content" | grep -o "$bin_pattern" | head -n 1)
  download_url="https://github.com$bin_path"
  latest_version=$(echo "$bin_path" | cut -d '/' -f6 | cut -d 'v' -f 2)
}

install_cherry() {
  download_url=https://github.com/flybits/cherry/releases/download/v0.4.4/cherry-linux-amd64
  bin_path="/usr/local/bin/cherry"

  if hash curl 2>/dev/null; then
    curl -fsSL -o "$bin_path" "$download_url"
  elif hash wget 2>/dev/null; then
    wget -qO "$bin_path" "$download_url"
  else
    echo "No command available to download $download_url"
    exit 1
  fi

  chmod 755 $bin_path
}

main() {
  echo "Installing the latest release of Cherry ..."

  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  arch=$(uname -m)

  if [ "$arch" = "i386" ]; then
    arch="386"
  elif [ "$arch" = "x86_64" ]; then
    arch="amd64"
  fi

  get_latest_release "$os" "$arch"
  install_cherry "$download_url"

  echo "Cherry $latest_version installed successfully."
}


main
