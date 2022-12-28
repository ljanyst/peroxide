#!/bin/bash

set -e

(
    cd cmd/peroxide
    go build
)

(
    cd cmd/peroxide-cfg
    go build
)

sudo cp cmd/peroxide/peroxide /usr/sbin
sudo cp cmd/peroxide-cfg/peroxide-cfg /usr/sbin

set +e

getent group peroxide >/dev/null 2>&1
if [ $? != 0 ]; then
    sudo groupadd -r peroxide
fi

GRP=""
getent group www-data >/dev/null 2>&1
if [ $? == 0 ]; then
    GRP="-G www-data"
fi

getent passwd peroxide >/dev/null 2>&1
if [ "$?" != "0" ]; then
    sudo useradd --system --no-create-home -g peroxide $GRP peroxide
fi

set -e

if [ ! -d /var/cache/peroxide ]; then
    sudo mkdir /var/cache/peroxide
    sudo chown peroxide:peroxide /var/cache/peroxide
    sudo chmod 700 /var/cache/peroxide
fi

if [ ! -d /var/lib/peroxide ]; then
    sudo mkdir /var/lib/peroxide
    sudo chown peroxide:peroxide /var/lib/peroxide
    sudo chmod 700 /var/lib/peroxide
fi

if [ ! -f /etc/peroxide.conf ]; then
    sudo cp config.example.yaml /etc/peroxide.conf
fi

if [ ! -d /etc/peroxide ]; then
    sudo mkdir /etc/peroxide
    sudo chown peroxide:peroxide /etc/peroxide
    sudo chmod 700 /etc/peroxide
fi

if [ ! -f /etc/systemd/system/peroxide.service ]; then
    sudo cp peroxide.service /etc/systemd/system/peroxide.service
    sudo systemctl daemon-reload
fi

if [ ! -d /var/log/peroxide ]; then
    sudo mkdir /var/log/peroxide
    sudo chown peroxide:peroxide /var/log/peroxide
    sudo chmod 750 /var/log/peroxide
fi

if [ -d /etc/logrotate.d ] && [ ! -f /etc/logrotate.d/peroxide ]; then
    sudo cp peroxide.logrotate /etc/logrotate.d/peroxide
    sudo systemctl restart logrotate
fi
