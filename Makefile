
build:
	./scripts/build.sh

test:
	cd tests && GPTSCRIPT_TOOL_REMAP="github.com/obot-platform/tools=.." go test -v tool_test.go

package-tools:
	./scripts/package-tools.sh

package-providers:
	./scripts/package-providers.sh

docker-build:
	docker build -t obot-platform/tools:latest --target tools .