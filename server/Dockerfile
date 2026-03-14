FROM node:20-alpine AS frontend
WORKDIR /app
COPY client/package*.json ./
RUN npm ci
COPY client/ .
RUN npm run build

FROM golang:1.26-alpine AS backend
WORKDIR /app
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=backend /server .
COPY --from=frontend /app/dist ./static
COPY server/internal/db/migrations ./internal/db/migrations
EXPOSE 8080
CMD ["./server"]
