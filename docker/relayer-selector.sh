#!/bin/bash
# Relayer version selector script

RELAYER_TYPE=$1
VERSION=${RELAYER_VERSION:-current}

case "$RELAYER_TYPE" in
  hermes)
    if [ "$VERSION" = "legacy" ]; then
      echo "Using Hermes legacy version"
      exec /usr/local/bin/hermes-legacy "${@:2}"
    else
      echo "Using Hermes current version"
      exec /usr/local/bin/hermes "${@:2}"
    fi
    ;;
  rly)
    if [ "$VERSION" = "legacy" ]; then
      echo "Using Go relayer legacy version"
      exec /usr/local/bin/rly-legacy "${@:2}"
    else
      echo "Using Go relayer current version"
      exec /usr/local/bin/rly "${@:2}"
    fi
    ;;
  *)
    echo "Unknown relayer type: $RELAYER_TYPE"
    echo "Usage: relayer-selector [hermes|rly] [args...]"
    exit 1
    ;;
esac