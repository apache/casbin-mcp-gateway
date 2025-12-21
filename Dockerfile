FROM golang:1.21-alpine AS backend-builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o mcp-gateway .

# Frontend builder stage
FROM node:20-alpine AS frontend-builder

WORKDIR /app

# Copy package files
COPY web/package*.json ./
RUN npm ci

# Copy source and build
COPY web/ ./
RUN npm run build

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy backend binary
COPY --from=backend-builder /app/mcp-gateway .

# Copy frontend build
COPY --from=frontend-builder /app/dist ./web/dist

# Copy configuration
COPY conf ./conf

EXPOSE 9000

CMD ["./mcp-gateway"]
