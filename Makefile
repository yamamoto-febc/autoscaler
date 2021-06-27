#
# Copyright 2017-2021 The Usacloud Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
GO_FILES        ?=$(shell find . -name '*.go')
AUTHOR          ?="The sacloud Authors"
COPYRIGHT_YEAR  ?="2021"
COPYRIGHT_FILES ?=$$(find . \( -name "*.dockerfile" -or -name "*.go" -or -name "*.sh" -or -name "*.pl" -or -name "*.bats" -or -name "*.bash" \) -print | grep -v "/vendor/")
BUILD_LDFLAGS   ?= "-s -w -X github.com/sacloud/autoscaler/version.Revision=`git rev-parse --short HEAD`"

export GOPROXY=https://proxy.golang.org

.PHONY: default
default: set-license fmt goimports lint test build

.PHONY: run
run:
	go run $(CURDIR)/main.go $(ARGS)

.PHONY: clean
clean:
	rm -Rf bin/*

.PHONY: tools
tools:
	(cd tools; go install golang.org/x/tools/cmd/goimports)
	(cd tools; go install golang.org/x/tools/cmd/stringer)
	(cd tools; go install github.com/sacloud/addlicense)
	(cd tools; go install google.golang.org/grpc/cmd/protoc-gen-go-grpc)
	(cd tools; go install google.golang.org/protobuf/cmd/protoc-gen-go)
	(cd tools; go install github.com/google/go-licenses)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/v1.40.0/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.40.0

.PHONY: gen
gen: gen-request gen-handler
gen-request:
	(cd protos; protoc --go_out=../request --go_opt=paths=source_relative --go-grpc_out=../request --go-grpc_opt=paths=source_relative request.proto)

gen-handler:
	(cd protos; protoc --go_out=../handler --go_opt=paths=source_relative --go-grpc_out=../handler --go-grpc_opt=paths=source_relative handler.proto)

.PHONY: build build-handlers-fake

build: bin/autoscaler

build-autoscaler: bin/autoscaler
bin/autoscaler: $(GO_FILES)
	GOOS=$${OS:-"`go env GOOS`"} GOARCH=$${ARCH:-"`go env GOARCH`"} CGO_ENABLED=0 go build -ldflags=$(BUILD_LDFLAGS) -o bin/autoscaler main.go

build-handlers-fake: bin/autoscaler-handlers-fake
bin/autoscaler-handlers-fake: $(GO_FILES)
	GOOS=$${OS:-"`go env GOOS`"} GOARCH=$${ARCH:-"`go env GOARCH`"} CGO_ENABLED=0 go build -ldflags=$(BUILD_LDFLAGS) -o bin/autoscaler-handlers-fake cmd/autoscaler-handlers-fake/main.go

.PHONY: shasum
shasum:
	(cd bin/; shasum -a 256 * > autoscaler_SHA256SUMS)

.PHONY: test
test: 
	go test $(TESTARGS) -v ./...

.PHONY: e2e-test

e2e-test:
	docker run -it --rm \
	    -v $$(PWD):/work \
	    -w /work/e2e \
	    -e SAKURACLOUD_ACCESS_TOKEN \
	    -e SAKURACLOUD_ACCESS_TOKEN_SECRET \
	    -e SKIP_CLEANUP \
	    ghcr.io/sacloud/autoscaler:e2e sh -c "cd e2e;./run.sh"

.PHONY: lint
lint:
	golangci-lint run ./... --modules-download-mode=readonly

.PHONY: goimports
goimports:
	goimports -l -w .

.PHONY: fmt
fmt:
	find . -name '*.go' | grep -v vendor | xargs gofmt -s -w

.PHONY: build-textlint
build-textlint:
	@echo "building sacloud/textlint:local"
	@docker build -t sacloud/textlint:local .github/actions/textlint

.PHONY: textlint
textlint:
	@docker run -it --rm -v $$PWD:/work -w /work sacloud/textlint:local .

.PHONY: go-licenses-check
go-licenses-check:
	go-licenses check .

.PNONY: generate-test-cert
generate-test-cert:
	# valid certs
	openssl req -x509 -text -newkey rsa:4096 -days 7300 -set_serial 1 -nodes -keyout test/ca-key.pem -out test/ca-cert.pem -subj "/C=JP/O=Usacloud/OU=Usacloud Certificate Authority/CN=Usacloud TLS CA";
	openssl req -text -newkey rsa:4096 -nodes -keyout test/server-key.pem -out test/server-csr.pem -subj "/C=JP/O=Usacloud/CN=usacloud.example.com"
	openssl x509 -text -req -in test/server-csr.pem -days 7300 -set_serial 2 -CA test/ca-cert.pem -CAkey test/ca-key.pem -CAcreateserial -out test/server-cert.pem -extfile test/openssl.ext
	openssl req -text -newkey rsa:4096 -nodes -keyout test/client-key.pem -out test/client-csr.pem -subj "/C=JP/O=Usacloud/CN=client01.usacloud.example.com"
	openssl x509 -text -req -in test/client-csr.pem -days 7300 -set_serial 3 -CA test/ca-cert.pem -CAkey test/ca-key.pem -CAcreateserial -out test/client-cert.pem  -extfile test/openssl.ext
	# invalid certs
	openssl req -x509 -text -newkey rsa:4096 -days 7300 -set_serial 1 -nodes -keyout test/invalid-ca-key.pem -out test/invalid-ca-cert.pem -subj "/C=JP/O=Usacloud/OU=Usacloud Certificate Authority/CN=Usacloud TLS CA";
	openssl req -text -newkey rsa:4096 -nodes -keyout test/invalid-client-key.pem -out test/invalid-client-csr.pem -subj "/C=JP/O=Usacloud/CN=client01.usacloud.example.com"
	openssl x509 -text -req -in test/invalid-client-csr.pem -days 7300 -set_serial 3 -CA test/invalid-ca-cert.pem -CAkey test/invalid-ca-key.pem -CAcreateserial -out test/invalid-client-cert.pem  -extfile test/openssl.ext
	rm -f test/*-csr.pem

set-license:
	@addlicense -c $(AUTHOR) -y $(COPYRIGHT_YEAR) $(COPYRIGHT_FILES)

.SUFFIXES:
.SUFFIXES: .go
