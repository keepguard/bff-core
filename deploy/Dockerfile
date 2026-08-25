# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Instalar dependências necessárias
RUN apk add --no-cache git

# Copiar arquivos de dependências
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fonte
COPY . .

# Build da aplicação e do utilitário de healthcheck
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bff-core cmd/bff-core/main.go \
    && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o healthcheck deploy/healthcheck/main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copiar binário da aplicação
COPY --from=builder /app/bff-core ./bff-core

# Copiar utilitário de healthcheck
COPY --from=builder /app/healthcheck ./healthcheck

# Copiar arquivos de configuração
COPY --from=builder /app/application*.yml ./

# Copiar documentação Swagger gerada
COPY --from=builder /app/docs ./docs

# Porta da aplicação
EXPOSE 8382

# Configuração padrão do healthcheck
ENV HEALTHCHECK_URL=http://localhost:8382/health

# Executar aplicação
CMD ["./bff-core"]
