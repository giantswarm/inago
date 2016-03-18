#!/bin/bash

if [ $TEST_SUITE == "unit" ]; then
    make lint
    make ci-test
    make ci-build
    ./inagoctl
elif [ $TEST_SUITE == "integration" ]; then
    # pip install --user fabric
    
    # make ci-build
    
    echo $SSH_KEY > ./ssh-key
    chmod 400 ./ssh-key
    
    cat ./ssh-key
    
    eval "$(ssh-agent -s)"
    ssh-add ./ssh-key 2>/dev/null
    
    ssh -vvvv -i ./ssh-key -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no core@ec2-52-58-14-174.eu-central-1.compute.amazonaws.com
    
    # fab run_int_test -i ./ssh-key --show=debug
else
    echo "Unknown test suite"
    exit 1
fi