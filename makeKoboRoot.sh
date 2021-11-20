#!/bin/sh

#GOOS=linux GOARCH=arm go build
cp kobomail usr/local/kobomail/
tar -cvzf KoboRoot.tgz -C . etc usr