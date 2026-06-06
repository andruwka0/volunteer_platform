.PHONY: run seed test make-admin make-leader

run:
	go run ./cmd/server

seed:
	go run ./cmd/seed

test:
	go test ./...

make-admin:
	go run ./cmd/make-admin -username "$${USERNAME}" -password "$${PASSWORD}" -full-name "$${FULL_NAME}"

make-leader:
	go run ./cmd/make-leader -username "$${USERNAME}"
