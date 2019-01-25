B=\033[0;1m
G=\033[0;92m
R=\033[0m

NAME ?= qlrepl
DIR = ${CURDIR}

.PHONY: help attach auto up down
# Show this help prompt
help:
	@echo '  Usage:'
	@echo ''
	@echo '    make <target>'
	@echo ''
	@echo '  Targets:'
	@echo ''
	@awk '/^#/{ comment = substr($$0,3) } comment && /^[a-zA-Z][a-zA-Z0-9_-]+ ?:/{ print "   ", $$1, comment }' $(MAKEFILE_LIST) | column -t -s ':' | grep -v 'IGNORE' | sort | uniq

down:
	@echo "\n${B}${G}Stop container${R}\n"
	@docker stop query-demo || true
	@docker rm query-demo || true

up: down
	@echo "\n${B}${G}build container${R}\n"
	@time docker build -t qlrepl .
	@echo "\n${B}${G}enter inside container:${R}\n"
	@time docker run -v ${DIR}/temp:/pics --rm -it --name query-demo qlrepl:latest
