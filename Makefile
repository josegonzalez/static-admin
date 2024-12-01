build: embedded_box/frontend/box.go
	go build

.PHONY: frontend/out
frontend/out:
	cd frontend && npm run build

embedded_box/frontend/box.go: frontend/out
	go generate ./...
