# Stage 1: Build frontend
FROM node:22-alpine AS frontend-build
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: Build Go binary with embedded frontend
FROM golang:1.25-alpine AS go-build
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Copy built frontend into the embed directory
COPY --from=frontend-build /app/web/dist ./internal/frontend/dist
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-X main.version=${VERSION}" -o youtube-release ./cmd/youtube-automation

# Stage 3: Minimal runtime
FROM gcr.io/distroless/static-debian12
COPY --from=go-build /app/youtube-release /youtube-release
EXPOSE 8080
ENTRYPOINT ["/youtube-release"]
CMD ["serve", "--host=0.0.0.0"]
