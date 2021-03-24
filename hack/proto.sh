#!/bin/bash

# Copyright 2021 The JamJar Relay Server Authors.

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

DIST_DIR="dist"
PROTOBUF_DIR="${DIST_DIR}/protobuf"

mkdir -p "${DIST_DIR}"
rm -rf "${PROTOBUF_DIR}"
mkdir -p "${PROTOBUF_DIR}"

cp "LICENSE" "${PROTOBUF_DIR}"

V1_DIR="${PROTOBUF_DIR}/v1"
mkdir -p "${V1_DIR}"
readarray -d '' proto_files_v1 < <(find specs/v1 -iname "*.proto" -print0)

for proto_file in "${proto_files_v1[@]}"
do
   cp "${proto_file}" "${V1_DIR}/"$(basename "${proto_file}")
done
