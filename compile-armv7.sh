#!/bin/sh
GOOS=linux GOARCH=arm GOARM=7 go build -ldflags "-w" -o bin/wol-mqtt main.go