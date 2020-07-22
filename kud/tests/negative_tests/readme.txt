These negative tests that validate EMCO open APIs with various invalid
inputs can be run individually or all together.

Step 1:
    cd k8s/src/orchestrator/scripts
    start-dev.sh
    ctrl z
    bg

Step 2:
    cd k8s/kud/tests/negative_tests
    ./test_all_final.sh

OR

Step 2:
    cd k8s/kud/tests/negative_tests
    ./test_<name>.sh
    example: ./test_project.sh
