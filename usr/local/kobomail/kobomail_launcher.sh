#!/bin/sh

logger -t "KoboMail" -p daemon.warning "Launcher: started"

if [ "$1" != "manual" ]; then
    #sleeping for 20s to wait for connection to be up and running
    logger -t "KoboMail" -p daemon.warning "Launcher: sleep for 20s"
    sleep 20
fi

UNINSTALL=/mnt/onboard/.adds/kobomail/UNINSTALL
if [ -f "$UNINSTALL" ]; then
    echo "$UNINSTALL exists, removing KoboMail..."
    logger -t "KoboMail" -p daemon.warning "Launcher: KoboMail UNINSTALL file located, removing KoboMail..."
    ./usr/local/kobomail/uninstall.sh
    logger -t "KoboMail" -p daemon.warning "Launcher: KoboMail removed"
else 
    echo "Running KoboMail..."
    logger -t "KoboMail" -p daemon.warning "Launcher: KoboMail binary execution started"
    /usr/local/kobomail/kobomail
    logger -t "KoboMail" -p daemon.warning "Launcher: KoboMail binary execution finished"
fi
logger -t "KoboMail" -p daemon.warning "Launcher: finished"