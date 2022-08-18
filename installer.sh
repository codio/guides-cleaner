#!/bin/sh

set -e

EXTRACT_PATH=cleaner.tar.gz
CLEANER_NAME="./guides-cleaner"

curl --fail -s https://api.github.com/repos/codio/guides-cleaner/releases/latest | grep linux | grep browser | cut -d : -f 2,3 | tr -d \" | xargs -n 1 curl -s --fail -o "${EXTRACT_PATH}" -L

tar zxf "${EXTRACT_PATH}"

rm "${EXTRACT_PATH}"

echo "${CLEANER_NAME}"

eval $CLEANER_NAME