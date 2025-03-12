#!/bin/bash

# This script sets up the easy-check project to run on system startup.

# Define the service name
SERVICE_NAME="easy-check"

# Define the path to the executable
EXECUTABLE_PATH="$(pwd)/bin/easy-check"

# Create a systemd service file for Linux
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
  SERVICE_FILE="/etc/systemd/system/$SERVICE_NAME.service"

  cat <<EOF >$SERVICE_FILE
[Unit]
Description=Easy Check Network Monitor
After=network.target

[Service]
ExecStart=$EXECUTABLE_PATH
Restart=always
User=$(whoami)

[Install]
WantedBy=multi-user.target
EOF

  # Reload systemd to recognize the new service
  systemctl daemon-reload
  # Enable the service to start on boot
  systemctl enable $SERVICE_NAME
  echo "Service installed and enabled to start on boot."
  systemctl status $SERVICE_NAME

# Create a task for Windows
elif [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
  # Create a scheduled task to run the executable at startup
  schtasks /create /tn "$SERVICE_NAME" /tr "$EXECUTABLE_PATH" /sc onlogon /rl highest
  echo "Scheduled task created to run on Windows startup."

else
  echo "Unsupported OS. Please set up the startup manually."
fi
