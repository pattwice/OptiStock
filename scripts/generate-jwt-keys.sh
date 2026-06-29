#!/usr/bin/env bash
set -euo pipefail

KEY_DIR="${1:-keys}"
mkdir -p "$KEY_DIR"

openssl genrsa -out "$KEY_DIR/private.pem" 2048
openssl rsa -in "$KEY_DIR/private.pem" -pubout -out "$KEY_DIR/public.pem"

echo "JWT keys written to $KEY_DIR/"
