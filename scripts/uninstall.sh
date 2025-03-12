#!/bin/bash

# This script removes the startup configuration for easy-check

# Remove systemd service (for Linux)
if [ -f /etc/systemd/system/easy-check.service ]; then
  sudo systemctl stop easy-check.service
  sudo systemctl disable easy-check.service
  sudo rm /etc/systemd/system/easy-check.service
  echo "Removed easy-check systemd service."
fi

# Remove autostart entry (for Linux desktop environments)
if [ -f ~/.config/autostart/easy-check.desktop ]; then
  rm ~/.config/autostart/easy-check.desktop
  echo "Removed easy-check autostart entry."
fi

# Remove Windows Task Scheduler entry (for Windows)
if schtasks /query | findstr /C:"easy-check" >nul; then
  schtasks /delete /tn "easy-check" /f
  echo "Removed easy-check from Windows Task Scheduler."
fi

echo "Uninstallation complete."
