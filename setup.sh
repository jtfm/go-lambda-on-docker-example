# This script must be run before any other deployment scripts
# It ensures that build directories and dockerfiles are created

mkdir -p bin
cp Dockerfile bin/Dockerfile