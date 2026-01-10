all: mig_up

mig_up:
	migrate -path migrations/ -database postgres://root:123@localhost:5432/auth_service?sslmode=disable up

mig_down:
	migrate -path migrations/ -database postgres://root:123@localhost:5432/auth_service?sslmode=disable down
