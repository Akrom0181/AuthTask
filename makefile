GO=go

# Initialize swagger documentation
swag_init:
	@swag init -g api/router.go -o api/docs

# Run database migrations up
migration_up:
	@migrate -path ./migrations/postgres -database 'postgres://barber_2yy0_user:wmTA7ajeoNMrPrDO0opWxNaTCE8HTkWt@dpg-ct5fmgrqf0us7386ntv0-a.oregon-postgres.render.com:5432/barber_2yy0?sslmode=require' up

# Run database migrations down
migration_down:
	@migrate -path ./migrations/postgres -database 'postgres://barber_2yy0_user:wmTA7ajeoNMrPrDO0opWxNaTCE8HTkWt@dpg-ct5fmgrqf0us7386ntv0-a.oregon-postgres.render.com:5432/barber_2yy0?sslmode=require' down

# Run the application
gaa:
	@$(GO) run cmd/main.go
