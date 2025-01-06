POSTGRES_URL=postgres://postgres:postgres@localhost:5432/ndn?sslmode=disable

.PHONY: migrate-create
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

.PHONY: migrate-up
migrate-up:
	migrate -path migrations -database "$(POSTGRES_URL)" up

.PHONY: migrate-down
migrate-down:
	migrate -path migrations -database "$(POSTGRES_URL)" down

.PHONY: migrate-force
migrate-force:
	@read -p "Enter version: " version; \
	migrate -path migrations -database "$(POSTGRES_URL)" force $$version

.PHONY: migrate-goto
migrate-goto:
	@read -p "Enter version: " version; \
	migrate -path migrations -database "$(POSTGRES_URL)" goto $$version

.PHONY: migrate-drop
migrate-drop:
	migrate -path migrations -database "$(POSTGRES_URL)" drop 