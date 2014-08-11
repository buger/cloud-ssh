#!/usr/bin/env bash

shopt -s extglob
set -o errtrace
set -o errexit

readonly PROGNAME=$(basename $0)
readonly ARGS="$@"
readonly VERSION="0.4"

log()  { printf "%b\n" "$*"; }
fail() { log "\nERROR: $*\n" ; exit 1 ; }

install_initialization(){
  if [ "$(uname)" == "Darwin" ]; then
     URL="https://github.com/buger/cloud-ssh/releases/download/$VERSION/cloud_ssh_macosx.tar.gz"
  elif [ "$(expr substr $(uname -s) 1 5)" == "Linux" ]; then
     URL="https://github.com/buger/cloud-ssh/releases/download/$VERSION/cloud_ssh_x86.tar.gz"
  elif [ "$(expr substr $(uname -s) 1 10)" == "MINGW32_NT" ]; then
    fail 'Installation script for Windows platform not yet supported. But you still can manually download binaries.'
  fi
}

download_binary(){
  TEMPFILE=$(mktemp /tmp/cloud-ssh.XXXXXX) 
  
  echo "Downloading binary from Github: $1"

  if curl --fail -L "$1" --progress-bar > $TEMPFILE 
  then
    echo "Unpacking"
  else
    fail "Failed to download binary from Github. Try later, or install manually."
  fi
}

unpack_binary(){
  tar -xzf $TEMPFILE -C $1
  rm "$TEMPFILE"
}

install_example_configuration(){
  EXAMPLE_CONFIG="https://raw.githubusercontent.com/buger/cloud-ssh/master/cloud-ssh.yaml.example"
  CONFIG_PATH=~/.ssh/cloud-ssh.yaml

  if [ ! -f $CONFIG_PATH ]; then
    curl --fail -L -# "$EXAMPLE_CONFIG" > $CONFIG_PATH
    echo "Configuration file located at $CONFIG_PATH"
  fi
}

install(){
  DEFAULT_PATH="/usr/local/bin"
  read -p "Choose path to install cloud-ssh binary? [Default: $DEFAULT_PATH ] " path
  path=${path:-$DEFAULT_PATH}

  download_binary $URL
  unpack_binary $path
  install_example_configuration

  echo "cloud-ssh succesfully installed"
}

main() {
  install_initialization
  install $ARGS
}
main
