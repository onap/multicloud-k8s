#=======================================================================
# Copyright (c) 2017-2020 Aarna Networks, Inc.
# All rights reserved.
# ======================================================================
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#           http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ========================================================================

FROM golang:1.14.1

# Set the Current Working Directory inside the container
WORKDIR /src
COPY ./ ./
RUN make all 

# Build the Go app
FROM ubuntu:16.04
WORKDIR /opt/emco
RUN groupadd -r emco && useradd -r -g emco emco
RUN chown emco:emco /opt/emco -R
RUN mkdir ./config
COPY --chown=emco --from=0 /src/middleend ./

# Command to run the executable
CMD ["./middleend"]
