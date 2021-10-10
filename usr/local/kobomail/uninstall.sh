#!/bin/sh

# Uninstall kobomail
rm -f /etc/udev/rules.d/97-kobomail.rules
rm -f /etc/ssl/certs/ca-certificates.crt
rm -rf /usr/local/kobomail/
rm -rf /mnt/onboard/.add/kobomail