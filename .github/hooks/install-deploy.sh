#!/usr/bin/env bash
# deploy in linux, ci/cd

cd /tmp/dist/ || exit
TARGET_OS=$(uname | tr '[:upper:]' '[:lower:]')
TARGET_ARCH=$(uname -m)
[ "${TARGET_ARCH}" == 'x86_64' ] && TARGET_ARCH=amd64
ls -t easy-check-*-${TARGET_OS}-${TARGET_ARCH}.tar.gz | tail -n +2 | xargs rm -f
tar -zxvf easy-check-*-${TARGET_OS}-${TARGET_ARCH}.tar.gz
cd easy-check-*-${TARGET_OS}-${TARGET_ARCH}
\mv easy-check /usr/local/bin/easy-check
[ ! -e /usr/bin/easy-check ] && ln -s /usr/local/bin/easy-check /usr/bin/easy-check
if [ ! -f /etc/systemd/system/easy-check.service ]; then
  mv scripts/easy-check.service /etc/systemd/system/easy-check.service
  systemctl daemon-reload
  systemctl enable easy-check
fi
rm -rf /tmp/dist
systemctl restart easy-check
