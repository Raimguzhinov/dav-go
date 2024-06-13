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

bin-deps:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/golang/mock/mockgen@latest