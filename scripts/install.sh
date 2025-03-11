#!/bin/bash

# This script sets up the easy-check project to run on system startup.

# Define the service name
SERVICE_NAME="easy-check"

# Define the path to the executable
EXECUTABLE_PATH="$(pwd)/cmd/easy-check"

# Create a systemd service file for Linux
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    SERVICE_FILE="/etc/systemd/system/$SERVICE_NAME.service"

    echo "[Unit]" >$SERVICE_FILE
    echo "Description=Easy Check Network Monitor" >>$SERVICE_FILE
    echo "After=network.target" >>$SERVICE_FILE
    echo "" >>$SERVICE_FILE
    echo "[Service]" >>$SERVICE_FILE
    echo "ExecStart=$EXECUTABLE_PATH" >>$SERVICE_FILE
    echo "Restart=always" >>$SERVICE_FILE
    echo "User=$(whoami)" >>$SERVICE_FILE
    echo "" >>$SERVICE_FILE
    echo "[Install]" >>$SERVICE_FILE
    echo "WantedBy=multi-user.target" >>$SERVICE_FILE

    # Reload systemd to recognize the new service
    systemctl daemon-reload
    # Enable the service to start on boot
    systemctl enable $SERVICE_NAME
    echo "Service installed and enabled to start on boot."

# Create a task for Windows
elif [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
    # Create a scheduled task to run the executable at startup
    schtasks /create /tn "$SERVICE_NAME" /tr "$EXECUTABLE_PATH" /sc onlogon /rl highest
    echo "Scheduled task created to run on Windows startup."

else
    echo "Unsupported OS. Please set up the startup manually."
fi
