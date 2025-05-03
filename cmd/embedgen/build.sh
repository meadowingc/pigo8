#!/bin/bash
# Build the embedgen tool
cd "$(dirname "$0")"
go build -o embedgen
