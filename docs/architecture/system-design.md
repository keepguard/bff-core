# bff-core - System Design

## 1. Propósito e Domínio
- **Responsabilidade Principal:** Backend for Frontend (BFF) do núcleo da plataforma KeepGuard. Orquestra registro de usuários (init/confirm/resend com SAGA), perfil (`/users/me`), consentimentos, billing (planos, assinaturas, gateways Asaas/Stripe), auditoria, AI Guardian, collector de dados, knowledge/LLM e gestão de OAuth clients — sem persistência própria de domínio.
- **Domínio/Subdomínio:** Core Platform / IAM & Onboarding, Billing & Entitlements, Ops & Observabilidade (audit, connections health), Data Collection & AI (guardian, knowledge, LLM gateway).

## 2. Tech Stack Local
- **Linguagem & Framework:** Go 1.24 / Echo v4 (`labstack/echo`), Resty (HTTP clients), Viper (config YAML por ambiente), Zap (logging), go-playground/validator, swaggo (Swagger em `/swagger/*`), Prometheus client.
- **Persistência e Cache:** Sem banco próprio. **Redis** para rate limiting, cache de respostas de `ms-user`/`ms-company`, snapshot de connections health, validação de sessão JWT (`tokenlogin:*`) e blacklist de tokens. Config: `redis.host/port/password/db`.
- **Mensageria:** **RabbitMQ** (publisher only):
  - Comunicação: exchange topic `ms-communication-exchange-{local|dev|prod}`, routing key `communication.message.send` (e-mail/SMS no fluxo de registro); decorators com logging, métricas e circuit breaker com fallback HTTP via `ms-communication`.
  - Auditoria: exchange `srv-audit-exchange-{local|dev|prod}`, routing key `audit.event` (middleware de auditoria; fail-open se broker indisponível).

## 3. Arquitetura Interna
- **Padrão Utilizado:** Arquitetura Hexagonal (Ports & Adapters) com elementos de DDD (entities, value objects, enums) e SAGA in-memory no registro. Camadas: `cmd` → `adapters/inbound` → `application` → `domain` ← `adapters/outbound` + `infrastructure`.
- **Módulos Principais:**
  - `cmd/bff-core` — composition root (wiring de clients, decorators, use cases, HTTP e métricas).
  - `internal/adapters/inbound/http` — server Echo, handlers, middleware (JWT, RBAC, rate limit, CORS, audit, company resolve, billing restricted).
  - `internal/adapters/outbound/http/client` + `decorator/*` — clientes Resty com metrics/retry/circuit-breaker/cache Redis/logging.
  - `internal/adapters/outbound/messaging` — publishers RabbitMQ (comunicação + audit).
  - `internal/application/{register,user,consent,consentdocument,billing,audit,guardian,oauth,collector,knowledge,llm,connections}` — ports/use cases de orquestração.
  - `internal/application/port` — interfaces de saída (Auth, User, Company, Communication, Consents, Audit, Guardian, Collector, Knowledge, LLM, Billing, OAuth, ServiceToken).
  - `internal/domain` — entities (`user`, `company`, `register_session`), value objects (`email`, `phone`, `cnpj`, `user_type`), saga engine, ports de messaging/audit.
  - `internal/infrastructure` — config, Redis cache, metrics, resilience (circuit breaker manager), logger, validation, clientip.

## 4. Superfície de Contato (I/O)
- **Endpoints Expostos Principais:** (prefixo `/api/v1`; health `/health`; métricas Prometheus em porta dedicada, ex. `9092/metrics`)
  - **Públicos:** `POST /register/{init,confirm,resend}`; `GET /consents/published`, `GET /consents/type/:type/latest`; `POST /core/billing/webhooks/{asaas,stripe}`.
  - **Autenticados (JWT):** `GET /users/me`; `POST /user-consents/accept-batch`.
  - **Ops:** `GET /core/connections/health` (`ops:read`).
  - **Audit:** `GET /core/audits`, `GET /core/audits/:eventId` (+ aliases legados `/audits*`) (`audit:read`).
  - **Guardian:** CRUD de incidents/actions e alert-recipients (`guardian:read|write`).
  - **OAuth clients:** CRUD + block/unblock + service-roles (`oauth:read|write`).
  - **Collector:** agents, data-sources, executions/payloads, incidents, bulk ops (`collector:read|write`).
  - **Knowledge:** `POST /core/knowledge/ask` (`knowledge:read`).
  - **LLM:** providers, complete, usage, alert-rules/firings (`llm:read|write` da pessoa logada; downstream com token de serviço do BFF).
  - **Billing:** entitlement, plans, gateway-accounts, subscriptions, invoices, entitlements, users/lookup (`billing:read|write` / org-read para gestores).
- **Dependências Externas:**
  - **HTTP downstream:** `ms-auth`, `ms-user`, `ms-company`, `ms-communication`, `ms-user-consents` (+ consent documents), `ms-billing`, `ms-ai-guardian`, `ms-knowledge`, `srv-audit`, `srv-data-collector`, `srv-llm-gateway`; config prevê `ms-user-profile` (8091).
  - **Token de serviço:** OAuth client_id `bff-core` via `ms-auth` (`client_credentials`, client cadastrado em System > Clients) — usado em knowledge/collector e em **todas** as chamadas ao `srv-llm-gateway`. As authorities do token vêm da service role `ROLE_SERVICE_BFF_CORE` (`knowledge:read`, `llm:read`, `llm:write`): é lá que se concede acesso ao LLM, não no gateway. O JWT de quem está logado não é repassado ao gateway — a autorização da pessoa acontece nas middlewares de rota do BFF.
  - **Infra:** Redis, RabbitMQ; probes de health agregam dezenas de serviços (front, BFFs, MSs, SRVs, MinIO, Prometheus, Grafana) apenas para a tela Conexões.
  - **Gateways de pagamento:** webhooks Asaas/Stripe recebidos no BFF e encaminhados ao `ms-billing`.

## 5. Invariantes Locais e Observações
- **Multi-tenancy:** `tenant_id` obrigatório no JWT (ou `X-Tenant-Id` em fluxos públicos). `CompanyResolveMiddleware` resolve tenant → `companyId` via `ms-company` e anexa ao contexto; webhooks de billing e health/swagger são isentos.
- **JWT local + Redis:** validação HMAC com secret compartilhado (não chama `ms-auth` para authN); checa sessão `tokenlogin:{codeUser}:{token}` e blacklist no Redis; `ms-auth` só para criação de usuário/login no registro.
- **RBAC por authorities:** rotas admin exigem `ADMIN`/`SYSTEM` **ou** authority específica (`billing:read`, `collector:write`, etc.). `RequireBillingOrgRead` restringe vistas org a ADMIN/SYSTEM ou MANAGER com billing.
- **Modo P (Billing Restricted):** mutações `*:write` (guardian, oauth, collector, llm) passam por `RequireBillingPlatformNotRestricted`. Status `restricted`/`canceled` (ou `allowsWrite=false`) → HTTP 403 `BILLING_RESTRICTED`. **ADMIN da empresa não bypassa; apenas role SYSTEM.** Falha de comunicação com billing é fail-open.
- **Registro com SAGA compensatória:** `register/confirm` orquestra user → auth → consents → login em memória; falhas disparam compensações (ex.: hard delete). E-mail de boas-vindas só após SAGA OK, via RabbitMQ (ou fallback HTTP).
- **Resiliência outbound:** circuit breakers (gobreaker) em auth/user/company e no publisher RabbitMQ; retry com jitter nos decorators; cache Redis em user/company.
- **Rate limit:** regras por família de rota no Redis, limites mais rígidos em prod (ex.: register_init 5/600s) do que em local.
- **Observabilidade:** correlation/request IDs, audit middleware publicando eventos, métricas HTTP/RabbitMQ, Swagger embutido. Porta HTTP: `8282` (local/base) ou `8382` (dev/prod K8s). Versão snapshot em config: `1.0.96-SNAPSHOT`.
- **Peculiaridade:** BFF sem datastore de domínio — estado vive nos microserviços; Redis é apenas cache/controle. Path canônico de audits é `/api/v1/core/audits`; `/api/v1/audits` é legado (Traefik/Ingress).
