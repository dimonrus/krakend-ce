include env/.vars
include ../Makefile

BIN_DIR = ./bin
PLUGIN_DIR = ./plugins
PLUGINS = $(shell find $(PLUGIN_DIR) -mindepth 1 -type d)
PWD = $(shell pwd)

# build extra kraken
extra: ../cmd/krakend-extra/schema/schema.json
	@echo "Building the extra binary..."
	@go get .
	@go build -ldflags="-X ${MODULE}/pkg.Version=${VERSION} -X github.com/luraproject/lura/v2/core.KrakendVersion=${VERSION} \
	-X github.com/luraproject/lura/v2/core.GlibcVersion=${GLIBC_VERSION} ${EXTRA_LDFLAGS}" \
	-o ${BIN_NAME} ../cmd/krakend-extra
	@echo "You can now use ./${BIN_NAME}"

# get kraken config schema
../cmd/krakend-extra/schema/schema.json:
	@echo "Fetching v${SCHEMA_VERSION} schema"
	@wget -qO $@ https://raw.githubusercontent.com/krakend/krakend-schema/refs/heads/main/v${SCHEMA_VERSION}/krakend.json || wget -qO $@ https://krakend.io/schema/krakend.json

# Build kraken bin in docker
kraken:
	make extra; \
	mv krakend bin/

# Build config from template
proxy:
	$(shell cat env/.vars | xargs) ./bin/krakend check -t -d -c "${PWD}/config/krakend.tmpl"

# Build extra plugins
plugin:
	for i in $(PLUGINS); do \
		cd $${i} && go build -buildmode=plugin -o ../../$(BIN_DIR)/$${i}.so; \
		cd $(PWD); \
	done

# Build kraken and plugins in docker
builder:
	docker build --no-cache --debug --pull --build-arg GOLANG_VERSION=${GOLANG_VERSION} -t krakend/builder:${VERSION}-linux-generic -f Dockerfile-build ../.

# Copy bin files to local
cpbin:
	docker cp $(shell docker create krakend/builder:${VERSION}-linux-generic):/app/extra/bin ./

# build docker for deploy on server
deploy:
	docker build --no-cache --pull --build-arg ALPINE_VERSION=${ALPINE_VERSION} -t krakend/api:${VERSION} -f Dockerfile-deploy ../.