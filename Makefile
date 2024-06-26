.SILENT:

compose-up: ### Run docker-compose
	docker compose --project-directory deployments up --build -d postgres
.PHONY: compose-up

compose-down: ### Down docker-compose
	docker compose --project-directory deployments down --remove-orphans
.PHONY: compose-down

migrate-create:  ### create new migration
	@if [ -z "$(BACKEND)" ] || [ -z "$(NAME)" ]; then \
		echo "Usage: make migrate-create BACKEND=<caldav|carddav> NAME=<name>"; \
		exit 1; \
	fi

	migrate create -ext sql -dir migrations/$(BACKEND) -seq $(NAME)
.PHONY: migrate-create

migrate-up: ### migration up
	@if [ -z "$(BACKEND)" ] || { [ "$(BACKEND)" != "caldav" ] && [ "$(BACKEND)" != "carddav" ]; }; then \
		echo "Usage: make migrate-up BACKEND=<caldav|carddav>"; \
		exit 1; \
	fi

	migrate -path migrations/$(BACKEND) -database \
	'$(PG_URL)?sslmode=disable&x-migrations-table=schema_migrations_$(BACKEND)' up
.PHONY: migrate-up

migrate-down: ### migration down
	@if [ -z "$(BACKEND)" ] || { [ "$(BACKEND)" != "caldav" ] && [ "$(BACKEND)" != "carddav" ]; }; then \
		echo "Usage: make migrate-down BACKEND=<caldav|carddav>"; \
		exit 1; \
	fi

	migrate -path migrations/$(BACKEND) -database \
	'$(PG_URL)?sslmode=disable&x-migrations-table=schema_migrations_$(BACKEND)' down
.PHONY: migrate-down

test:
	CONFIG_PATH=$(shell pwd)/configs/config.yml \
	TESTCASES_DIR=$(shell pwd)/tests/cases \
	gotestsum --format pkgname --raw-command go test -json -cover -count=1 ./...
.PHONY: test

raw-event:
	@if [ -z "$(FILE_UID)" ] || [ -z "$(NAME)" ]; then \
		echo "Usage: make raw-event FILE_UID=<uuid> NAME=<testcase_name>"; \
		exit 1; \
	fi
	curl "http://$(HTTP_SERVER_USER):$(HTTP_SERVER_PASSWORD)@$(HTTP_SERVER_IP):$(HTTP_SERVER_PORT)/$(HTTP_SERVER_USER)/calendars/1/$(FILE_UID).ics" \
		| tee tests/cases/$(NAME).in.ics
	cp tests/cases/$(NAME).in.ics tests/cases/$(NAME).out.ics
.PHONY: raw-event

bin-deps:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/golang/mock/mockgen@latest

generate:
	protoc --proto_path=api --go_out=. \
		--go_opt=module=github.com/Raimguzhinov/dav-go \
		--go-grpc_out=. --go-grpc_opt=module=github.com/Raimguzhinov/dav-go \
		api/protobuf/caldav.proto
.PHONY: generate