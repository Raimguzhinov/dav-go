compose-up: ### Run docker-compose
	docker compose --project-directory deployments up --build -d postgres
.PHONY: compose-up

compose-down: ### Down docker-compose
	docker compose --project-directory deployments down --remove-orphans
.PHONY: compose-down

migrate-create:  ### create new migration
	migrate create -ext sql -dir migrations -seq 'create_caldav_tables'
.PHONY: migrate-create

migrate-up: ### migration up
	migrate -path migrations -database '$(PG_URL)?sslmode=disable' up
.PHONY: migrate-up

migrate-down: ### migration down
	migrate -path migrations -database '$(PG_URL)?sslmode=disable' down
.PHONY: migrate-up

bin-deps:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/golang/mock/mockgen@latest