#!/usr/bin/env bash
set -ex

VERSION=$(git describe --always --tags --long)
PLATFORM=""

if [[ ${RUNNER_OS} == 'Linux' ]]; then
  PLATFORM="linux"
elif [[ ${RUNNER_OS} == 'macOS' ]]; then
  PLATFORM="darwin"
else
  PLATFORM="windows"
  exit 1
fi



env GO111MODULE=on make ontology-${PLATFORM} tools-${PLATFORM}
mkdir tool-${PLATFORM}
cp ./tools/abi/* tool-${PLATFORM}
cp ./tools/sigsvr* tool-${PLATFORM}

zip -q -r tool-${PLATFORM}.zip tool-${PLATFORM};
rm -r tool-${PLATFORM};

set +x
echo "ontology-${PLATFORM}-amd64 |" $(md5sum ontology-${PLATFORM}-amd64|cut -d ' ' -f1)
echo "tool-${PLATFORM}.zip |" $(md5sum tool-${PLATFORM}.zip|cut -d ' ' -f1)
