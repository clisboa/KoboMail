#!/bin/sh

KM_HOME=$(dirname $0)

ConfigFileTemplate=$KM_HOME/kobomail_cfg.toml.tmpl
UserConfig=/mnt/onboard/.add/kobomail/kobomail_cfg.toml

if [ -f $UserConfig ]; then
    logger -t "KoboMail" -p daemon.warning "Setup: config file in place"
    echo "$UserConfig in place"
else
    logger -t "KoboMail" -p daemon.warning "Setup: config file not found. Creating directories and config file template"
    mkdir /mnt/onboard/.add
    mkdir /mnt/onboard/.add/kobomail
    cp $ConfigFileTemplate $UserConfig
    mkdir /mnt/onboard/KoboMailLibrary
    touch /mnt/onboard/.add/kobomail/logs.txt
    logger -t "KoboMail" -p daemon.warning "Setup: Directories and config file template created"
fi

CERTS=/etc/ssl/certs/ca-certificates.crt
if [ -f "$CERTS" ]; then
    logger -t "KoboMail" -p daemon.warning "Setup: Certificates in place"
    echo "$CERTS in place"
else
    logger -t "KoboMail" -p daemon.warning "Setup: Certificates not found, placing them"
    echo "$CERTS missing, placing own certs"
    cd /usr/local/kobomail/
    cp --parents ssl/certs/ca-certificates.crt /etc
    logger -t "KoboMail" -p daemon.warning "Setup: Certificates installed"
fi

logger -t "KoboMail" -p daemon.warning "Setup: finished"