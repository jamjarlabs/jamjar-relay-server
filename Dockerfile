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

# Build stage
FROM golang:1.16.2
# Set up build dir
WORKDIR /build
# Copy in source files
COPY ./ ./
# Build the binary
RUN make linux_amd64

# Container stage
FROM gcr.io/distroless/static
WORKDIR /app/
COPY --from=0 /build/dist/linux_amd64 .
USER nonroot:nonroot

ENV PORT="8000"
ENV ADDRESS="0.0.0.0"

CMD [ "/app/jamjar-relay-server" ]
