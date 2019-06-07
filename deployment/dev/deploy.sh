#!/bin/bash

# considering that you have git clone from repo and stored into /tmp folder

# add user to run service 'locationtracker'
sudo useradd locationtracker -s /sbin/nologin -M

# move service file
cd /tmp/Fleet-Location/deployment/dev
sudo cp locationtracker-udp.service /etc/systemd/system/
sudo chmod 755 /etc/systemd/system/locationtracker-udp.service

sudo cp locationtracker-http.service /etc/systemd/system/
sudo chmod 755 /etc/systemd/system/locationtracker-http.service

sudo cp locationtracker-socket.service /etc/systemd/system/
sudo chmod 755 /etc/systemd/system/locationtracker-socket.service

echo "service moved"

# copy source code to go folder
mkdir -p /home/rsa-key-20190103/go/src/github.com/iknowhtml/locationtracker
cd /tmp/Fleet-Location
rm -rf /home/rsa-key-20190103/go/src/github.com/iknowhtml/locationtracker/*
cp -r * /home/rsa-key-20190103/go/src/github.com/iknowhtml/locationtracker/
cd /home/rsa-key-20190103/go/src/github.com/iknowhtml/locationtracker
/usr/local/go/bin/go build

echo "building source..."

sleep 10s

echo "source built"

# enable and start the service
sudo systemctl daemon-reload

sudo systemctl enable locationtracker-udp
sudo systemctl restart locationtracker-udp

sudo systemctl enable locationtracker-http
sudo systemctl restart locationtracker-http

sudo systemctl enable locationtracker-socket
sudo systemctl restart locationtracker-socket

echo "deployment complete"