from os.path import join
from time import sleep

from fabric.api import env, local, put, run
from fabric.context_managers import cd

BINARY = 'inagoctl'
INT_TESTS_DIR = 'int-tests'
VAGRANT_DIR = 'vagrant'

env.hosts = ['core@ec2-52-58-89-242.eu-central-1.compute.amazonaws.com']
env.disable_known_hosts = True
env.colorize_errors = True
env.command_timeout = 60 * 5

def create_build_directory():
    """ Create a temporary directory for us to run the test in. """
    
    return run('mktemp -d')

def upload_binary_and_tests(build_directory):
    """ Upload the binary and the integration tests. """
    
    put(BINARY, build_directory)
    run('chmod +x %s' % join(build_directory, BINARY))
    put(INT_TESTS_DIR, build_directory)

def run_cram_container(build_directory):
    """ Run the cram container. """
    
    int_tests_path = join(build_directory, INT_TESTS_DIR)
    
    run(
        """docker run --rm -ti \
-e FLEET_ENDPOINT=unix:///var/run/fleet.sock \
-v /var/run/fleet.sock:/var/run/fleet.sock \
-v {build_directory}/inagoctl:/usr/local/bin/inagoctl \
-v {int_tests_path}:{int_tests_path} \
zeisss/cram-docker -v {int_tests_path}""".format(**{
            'build_directory': build_directory,
            'int_tests_path': int_tests_path,
        })
    )

def run_int_test():
    """ Run the integration test. """
    
    build_directory = create_build_directory()
    
    upload_binary_and_tests(build_directory)
    run_cram_container(build_directory)
