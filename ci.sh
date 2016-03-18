#!/bin/bash

if [ $TEST_SUITE == "unit" ]; then
    make lint
    make ci-test
    make ci-build
    ./inagoctl
elif [ $TEST_SUITE == "integration" ]; then
    pip install --user fabric
    make ci-build
    eval "$(ssh-agent -s)"
    chmod 400 ./inago-integration-test.pem
    ssh-add ./inago-integration-test.pem 2>/dev/null
    fab run_int_test
else
    echo "Unknown test suite"
    exit 1
fi
