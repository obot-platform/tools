
build:
	./scripts/build.sh

test:
	cd tests && GPTSCRIPT_TOOL_REMAP="github.com/obot-platform/tools=.." go test -v tool_test.go
	cd github-auth-provider && go test ./... && cd ..
	cd google-auth-provider && go test ./... && cd ..

package-tools:
	./scripts/package-tools.sh

package-providers:
	./scripts/package-providers.sh

docker-build:
	docker build -t obot-platform/tools:latest --target tools .