#!env sh
set -ex

MODULE=pydaemon

ROOT=$(git rev-parse --show-toplevel)
CLIENT=${ROOT}/daemon/client/${MODULE}

mkdir -p ${CLIENT}
rm -rf ${CLIENT}/schema

python -m grpc_tools.protoc -I${ROOT}/daemon/api/grpc/ --python_out=${CLIENT} --grpc_python_out=${CLIENT} schema/daemon.proto
touch ${CLIENT}/schema/__init__.py