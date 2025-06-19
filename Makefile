include .env

generate:
	sqlc generate

generate-web:
	templ generate

migrate-up:
	goose -dir ./migrations postgres $(DATABASE_URL) up

create-migration:
ifdef name
	@goose -dir ./migrations create $(name) sql
else
	@echo "name is required"
	@exit 1
endif

seed:
	psql $(DATABASE_URL) -f seed.sql
.PHONY: generate migrate-up create-migration generate-web
