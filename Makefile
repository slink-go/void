ifndef VERBOSE
.SILENT:
endif

# --- directories ------------------------------------
ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
void_src_dir = $(ROOT_DIR)/api-gateway/cmd/gin
back_src_dir = $(ROOT_DIR)/backend/gin
bin_dst_dir = $(ROOT_DIR)/app/run
build_dir = $(ROOT_DIR)/app/build
# ----------------------------------------------------

# --- images -----------------------------------------
published_image_void = slinkgo/void
local_image_void = mvkvl/api-gateway
published_image_back = slinkgo/test-backend
local_image_back = mvkvl/api-backend
# ----------------------------------------------------

# --- defaults ---------------------------------------
OS=darwin
ARCH=arm64
GOLANG_VERSION=1.22.3
UPX_VERSION=3.96-2
PLATFORMS=linux/arm/v7,linux/arm64/v8,linux/amd64
VERSION_SHORT:=$(shell cat $(build_dir)/VERSION)
FILE:=
IMAGE:=
TYPE:=
VERSIONS:=
# ----------------------------------------------------

# --- calculated -------------------------------------
OSN:=$(OS)
ifeq ($(OSN), darwin)
  OSN=macos
endif
EXT:=
ifeq ($(OSN), windows)
  EXT=.exe
endif
SRCDIR:=$(void_src_dir)
ifeq ($@, back)
  SRCDIR=$(back_src_dir)
endif
# ----------------------------------------------------

# --- macro ------------------------------------------
define create-bin
  	@echo building $@ $(OSN)-$(ARCH) at $(SRCDIR) && \
	mkdir -p $(ROOT_DIR)/app/distr/$@/$(OSN)/$(ARCH) && \
	cd $(SRCDIR) && \
	GO111MODULE=on go get -d -v ./... && \
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build -ldflags="-s -w" -o $(ROOT_DIR)/app/distr/$@/$(OSN)/$(ARCH)/$@$(EXT) . && \
	cd $(ROOT_DIR)/app/distr/$@/$(OSN)/$(ARCH) && \
	tar cfz $@_$(VERSION_SHORT)_$(OSN)_$(ARCH).tgz $@$(EXT) > /dev/null
endef
# ----------------------------------------------------

.PHONY: 		\
	clean 		\
	all 		\
	binaries    \
	images      \
	flush

all: clean binaries images
clean:
	@make binaries-clean
	@make docker-clean
	echo "> cleaned"
binaries:
	@make bin OS=darwin ARCH=amd64
	@make bin OS=darwin ARCH=arm64
	@make bin OS=linux ARCH=amd64
	@make bin OS=linux ARCH=arm64
	@make bin OS=windows ARCH=amd64
	echo "> built programs"
images: docker-login
	@make docker-setup
	@make docker-void DOCKER_FILE="gin" IMAGE="void" TYPE="" VERSIONS="$(VERSION_SHORT) latest"
	@make docker-back DOCKER_FILE=back IMAGE=test-backend TYPE=gin VERSIONS="0.0.1-gin latest"
	@make docker-teardown
	@make flush
	echo "> built docker images"
flush:
	docker container ls -aq | xargs docker stop
	docker container ls -aq | xargs docker container rm

bin: void back
void:
	$(create-bin)
back:
	$(create-bin)
binaries-clean:
	(rm $(bin_dst_dir)/backend_gin 			 2> /dev/null || true)					&& \
	(rm $(bin_dst_dir)/gateway_gin 			 2> /dev/null || true)					&& \
	(rm $(bin_dst_dir)/back 				 2> /dev/null || true)					&& \
	(rm $(bin_dst_dir)/void 				 2> /dev/null || true)					&& \
	(cd $(void_src_dir) && go clean ./...) 											&& \
	(cd $(back_src_dir) && go clean ./... && rm -rf protobuf/*.pb.go 2>/dev/null) 	&& \
	(rm -r $(ROOT_DIR)/app/distr 2> /dev/null || true)

docker-login:
	source $(build_dir)/.env && \
	echo $${DOCKER_PASSWORD:?} | docker login --password-stdin -u $${DOCKER_LOGIN:?}
docker-setup:
	docker buildx create --use
docker-teardown:
	docker buildx rm
docker-void:
	make create-image DOCKER_FILE=gin IMAGE=void TYPE= VERSIONS="$(VERSION_SHORT) latest"
docker-back:
	make create-image DOCKER_FILE=back IMAGE=test-backend TYPE=gin VERSIONS="0.0.1-gin latest"
create-image:
	$(foreach ver, $(VERSIONS), $(eval TAGS=$(TAGS) -t slinkgo/$(IMAGE):$(ver) ))
	cd $(build_dir)                                      && \
	docker buildx build                                 	\
		-f $(build_dir)/Dockerfile-$(DOCKER_FILE)           \
		--push                                           	\
		--platform $(PLATFORMS)                          	\
		--build-arg "GOLANG_VERSION=$(GOLANG_VERSION)"   	\
		--build-arg "UPX_VERSION=$(UPX_VERSION)"         	\
		--build-arg "TYPE=$(TYPE)"                			\
		$(TAGS) .
docker-clean:
	(docker image rm $(published_image_void) 2> /dev/null || true)					&& \
	(docker image rm $(local_image_void) 	 2> /dev/null || true)					&& \
	(docker image rm $(published_image_back) 2> /dev/null || true)					&& \
	(docker image rm $(local_image_back) 	 2> /dev/null || true)
