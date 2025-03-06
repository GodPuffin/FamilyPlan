#!/bin/bash

# Configure firewall
echo "Configuring firewall to allow port 8090..."
ufw allow 8090/tcp
ufw status

# Check if the service is running
echo "Checking service status..."
systemctl status familyplan

# Check network listening status
echo "Checking network listening status..."
ss -tulpn | grep 8090

# Restart the service to ensure changes take effect
echo "Restarting service..."
systemctl restart familyplan
systemctl status familyplan

echo "Configuration complete!" 