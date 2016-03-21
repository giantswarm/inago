#!/bin/bash

set -e

FILES=(gitcookies.sh inago-integration-test.pem)
SECRET_FILE=secrets.tar

tar cvf $SECRET_FILE ${FILES[*]}
travis encrypt-file $SECRET_FILE
