FROM mcr.microsoft.com/devcontainers/go:1.25 AS dev

USER vscode

# Install air
RUN go install github.com/air-verse/air@latest

WORKDIR /app

FROM golang:1.25-alpine AS prod_build

RUN apk update && apk add --no-cache make curl jq git

# Install swagger
RUN download_url=$(curl -s https://api.github.com/repos/go-swagger/go-swagger/releases/latest | \
      jq -r '.assets[] | select(.name | contains("'"$(uname | tr '[:upper:]' '[:lower:]')"'_amd64")) | .browser_download_url') && \
    curl -o /usr/local/bin/swagger -L'#' "$download_url" && \
    chmod +x /usr/local/bin/swagger

RUN go install github.com/golang/mock/mockgen@v1.6.0

COPY ../.. /app
WORKDIR /app
RUN make build

FROM golang:1.25-alpine AS prod

COPY --from=prod_build /app/cmd/app-server/main /app/main

WORKDIR /app
CMD ["./main", "--scheme=http", "--port", "80", "--host", "0.0.0.0"]
