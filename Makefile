.PHONY: dep
dep:
	go mod tidy && go mod verify

.PHONY: test
test:
	go test ./...

.PHONY: cover
cover:
	mkdir -p tmp
	go test -coverprofile=./tmp/cover.out ./...
	go tool cover -html=./tmp/cover.out

.PHONY: lint
lint:
	golangci-lint run --fix

.PHONY: build
build:
	go build -o ./bin/timer .

.PHONY: run
run:
	./bin/timer

.PHONY: docker
docker:
	docker build -t ghcr.io/tetafro/timer .

.PHONY: deploy
deploy:
	ansible-playbook \
	--become \
	--private-key ~/.ssh/id_ed25519 \
	--inventory '${SSH_SERVER},' \
	--user ${SSH_USER} \
	./playbook.yml
