# Stage 1: Build frontend
FROM node:22-slim AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build backend
FROM golang:1.23-alpine AS backend-builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/frontend/out /app/frontend/out
RUN go generate ./...
RUN CGO_ENABLED=0 GOOS=linux go build -o /static-admin

# Stage 3: Final stage
FROM alpine:3.21
WORKDIR /app
COPY --from=backend-builder /static-admin .
ENV GIN_MODE=release
CMD ["/app/static-admin"] 