
build:
	./scripts/build.sh

test:
	cd tests && GPTSCRIPT_TOOL_REMAP="github.com/obot-platform/tools=.." go test -v tool_test.go
