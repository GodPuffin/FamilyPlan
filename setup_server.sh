#!/bin/bash
# Setup script for FamilyPlan server
# Run this script on your DigitalOcean droplet to configure server access

set -e

echo "==== FamilyPlan Server Setup ===="
echo

# Fix permissions for application files
echo "Setting up permissions..."
DEPLOY_PATH="/var/www/familyplan"
chmod +x $DEPLOY_PATH/familyplan

# Configure firewall
echo "Configuring firewall..."
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 8090/tcp
ufw --force enable
ufw status

# Check if service file exists
if [ -f /etc/systemd/system/familyplan.service ]; then
  echo "Updating systemd service to bind explicitly..."
  # Add explicit binding to systemd service
  cat > /etc/systemd/system/familyplan.service << EOL
[Unit]
Description=Family Plan Application Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=${DEPLOY_PATH}
ExecStart=${DEPLOY_PATH}/familyplan serve --http=0.0.0.0:8090
Restart=on-failure
RestartSec=5
Environment="PORT=8090"

[Install]
WantedBy=multi-user.target
EOL

  # Reload and restart service
  systemctl daemon-reload
  systemctl restart familyplan
  systemctl status familyplan
else
  echo "Service file not found. Is the application deployed correctly?"
fi

# Check if the service is listening properly
echo "Checking network status..."
ss -tulpn | grep 8090

# Test local connection
echo "Testing local connection..."
curl -v localhost:8090 || echo "Local connection failed"

echo 
echo "Setup complete! Your application should now be accessible at http://$(curl -s ifconfig.me):8090"
echo "If you still can't connect, check DigitalOcean's firewall settings in the control panel." 