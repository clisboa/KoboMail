#!/bin/sh

#GOOS=linux GOARCH=arm go build -o kobomail
cp kobomail usr/local/kobomail/
tar -cvzf KoboRoot.tgz -C . etc usr