#!/bin/bash
# =============================================================================
# 🚀 MINI MANUAL DE USO — script-deploy-github-bff-core.sh (bff-core)
# =============================================================================
# Este script automatiza o ciclo completo de desenvolvimento, build e deploy
# do serviço bff-core.
#
# 🌿 FLUXO DE BRANCHES RECOMENDADO:
#   - Por padrão, o trabalho diário ocorre na branch 'develop'.
#   - O script detecta automaticamente em qual branch você está.
#   - Se houver alterações locais não comitadas, cria o commit e faz push na branch ativa.
#
# 📋 OPÇÕES E EXEMPLOS DE USO:
#
#   1. Apenas commit, push e build da imagem da branch atual no GHCR:
#      $ ./script-deploy-github-bff-core.sh
#      -> Envia para a branch ativa (ex: develop) e gera as tags:
#         ghcr.io/keepguard/bff-core:develop-<sha>
#         ghcr.io/keepguard/bff-core:develop-latest
#
#   2. Commit, push, build e deploy imediato no Docker Compose local:
#      $ ./script-deploy-github-bff-core.sh up
#      -> Atualiza o docker-compose.yml local para a tag gerada e recria o container.
#
#   3. Commit na branch atual + Merge automático na 'main' (versão Produção):
#      $ ./script-deploy-github-bff-core.sh merge main
#      -> Comita na branch atual, faz checkout da 'main', mescla sem necessidade
#         de PR manual, faz push na 'main', constrói as tags de produção
#         (:<sha> e :latest), e retorna com segurança para a branch de trabalho!
#
#   4. Merge na 'main' + deploy imediato no Docker Compose local:
#      $ ./script-deploy-github-bff-core.sh merge main up
#
#   5. Pipeline Completo em 1 comando (Merge 'main' + Docker Local + K8s Produção):
#      $ ./script-deploy-github-bff-core.sh merge main up prod
#      OU simplesmente:
#      $ ./script-deploy-github-bff-core.sh full
#      -> Faz commit na branch atual, merge na 'main', push GHCR, atualiza o Docker
#         local com 'up' e aplica imediatamente no cluster Kubernetes de Produção!
#
# 🚢 DEPLOY EM PRODUÇÃO (KUBERNETES HOSTINGER):
#   - Após o merge em 'main', aplique a versão no cluster executando:
#      $ ./script-deploy-k8s-prod.sh
#      (ou use o parâmetro 'prod' / 'full' acima para automação total)
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
SERVICE_NAME="bff-core"
DOCKER_COMPOSE_DIR="${PROJECT_ROOT}/docker"
DOCKER_COMPOSE_FILE="${DOCKER_COMPOSE_DIR}/docker-compose.yml"
REGISTRY="ghcr.io/keepguard"
# Cores
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "${BLUE}[STEP]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }

MERGE_TARGET=""
DEPLOY_DOCKER=false
DEPLOY_PROD=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        up)
            DEPLOY_DOCKER=true
            shift
            ;;
        prod)
            DEPLOY_PROD=true
            shift
            ;;
        full|all)
            MERGE_TARGET="main"
            DEPLOY_DOCKER=true
            DEPLOY_PROD=true
            shift
            ;;
        merge)
            MERGE_TARGET="${2:-main}"
            shift 2
            ;;
        *)
            shift
            ;;
    esac
done

cd "${SCRIPT_DIR}"

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    log_error "Diretório não é um repositório git válido."
    exit 1
fi

CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)

# 1. Commit e Push na branch ativa
log_step "1/5 Verificando repositório Git na branch '${CURRENT_BRANCH}'..."
git add -A
if ! git diff --cached --quiet; then
    log_info "Criando commit com alterações pendentes..."
    git commit -m "feat(${SERVICE_NAME}): update ${SERVICE_NAME} $(date +'%Y-%m-%d %H:%M')"
    log_info "Fazendo push para branch '${CURRENT_BRANCH}'..."
    git push origin "${CURRENT_BRANCH}"
else
    log_info "Nenhuma alteração pendente na branch '${CURRENT_BRANCH}'."
fi
LOCAL_SHA=$(git rev-parse --short HEAD)

# 2. Se solicitado merge para outra branch (ex: merge main)
if [ -n "$MERGE_TARGET" ] && [ "$MERGE_TARGET" != "$CURRENT_BRANCH" ]; then
    log_step "Promovendo branch '${CURRENT_BRANCH}' para '${MERGE_TARGET}'..."
    git checkout "${MERGE_TARGET}"
    git pull origin "${MERGE_TARGET}" --rebase || true
    git merge "${CURRENT_BRANCH}" -m "chore(merge): merge branch '${CURRENT_BRANCH}' into ${MERGE_TARGET}"
    git push origin "${MERGE_TARGET}"
    BUILD_SHA=$(git rev-parse --short HEAD)
    BUILD_BRANCH="${MERGE_TARGET}"
    log_info "Retornando checkout com segurança para a branch de trabalho '${CURRENT_BRANCH}'..."
    git checkout "${CURRENT_BRANCH}"
else
    BUILD_SHA="${LOCAL_SHA}"
    BUILD_BRANCH="${CURRENT_BRANCH}"
fi

# 3. Definição das tags do GHCR
if [ "$BUILD_BRANCH" = "main" ]; then
    PRIMARY_TAG="${REGISTRY}/${SERVICE_NAME}:${BUILD_SHA}"
    LATEST_TAG="${REGISTRY}/${SERVICE_NAME}:latest"
    BRANCH_TAG="${REGISTRY}/${SERVICE_NAME}:main-${BUILD_SHA}"
    BRANCH_LATEST="${REGISTRY}/${SERVICE_NAME}:main-latest"
else
    PRIMARY_TAG="${REGISTRY}/${SERVICE_NAME}:${BUILD_BRANCH}-${BUILD_SHA}"
    LATEST_TAG="${REGISTRY}/${SERVICE_NAME}:${BUILD_BRANCH}-latest"
    BRANCH_TAG="${PRIMARY_TAG}"
    BRANCH_LATEST="${LATEST_TAG}"
fi

log_info "============================================"
log_info "  Deploy ${SERVICE_NAME}"
log_info "============================================"
log_info "Branch de Trabalho : ${CURRENT_BRANCH}"
log_info "Branch de Build    : ${BUILD_BRANCH}"
log_info "Commit SHA         : ${BUILD_SHA}"
log_info "Deploy Docker Local: ${DEPLOY_DOCKER}"
log_info "Imagem Tag         : ${PRIMARY_TAG}"
log_info "Imagem Latest      : ${LATEST_TAG}"
log_info "============================================"
# Produção (GHCR/K8s) SEMPRE exige linux/amd64.
GO_ARCH=amd64
# 2. Compilação dos binários nativamente em linux/amd64
log_step "2/5 Compilando binários Go (linux/amd64)..."
mkdir -p .bin
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o .bin/bff-core cmd/bff-core/main.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o .bin/healthcheck deploy/healthcheck/main.go
log_success "Binários compilados com sucesso em .bin/ (linux/amd64)"

# 3. Build da Imagem Docker (linux/amd64)
# --platform linux/amd64 é obrigatório: sem ele o Docker do Mac (arm64) puxa base arm64
# e o nó amd64 de produção cai em ImagePullBackOff com "no match for platform in manifest".
log_step "3/5 Construindo imagem Docker (linux/amd64)..."
if [ "$BUILD_BRANCH" = "main" ]; then
    docker build --platform linux/amd64 -f deploy/Dockerfile.local -t "${PRIMARY_TAG}" -t "${LATEST_TAG}" -t "${BRANCH_TAG}" -t "${BRANCH_LATEST}" .
else
    docker build --platform linux/amd64 -f deploy/Dockerfile.local -t "${PRIMARY_TAG}" -t "${LATEST_TAG}" .
fi
log_success "Imagem Docker construída com sucesso"

# 4. Push para GitHub Container Registry
log_step "4/5 Fazendo push para o GitHub Container Registry..."
docker push "${PRIMARY_TAG}" || true
docker push "${LATEST_TAG}" || true
if [ "$BUILD_BRANCH" = "main" ]; then
    docker push "${BRANCH_TAG}" || true
    docker push "${BRANCH_LATEST}" || true
fi
log_success "Push concluído para as tags do GHCR"
# 5. Atualização e Deploy Docker Compose Local (se solicitado 'up')
if [ "$DEPLOY_DOCKER" = true ]; then
    log_step "Atualizando docker-compose.yml e subindo container local..."
    if [ -f "$DOCKER_COMPOSE_FILE" ]; then
        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' "s|image: ${REGISTRY}/${SERVICE_NAME}:.*|image: ${PRIMARY_TAG}|g" "${DOCKER_COMPOSE_FILE}"
        else
            sed -i "s|image: ${REGISTRY}/${SERVICE_NAME}:.*|image: ${PRIMARY_TAG}|g" "${DOCKER_COMPOSE_FILE}"
        fi
        log_info "docker-compose.yml atualizado para ${PRIMARY_TAG}"
    fi

    cd "${DOCKER_COMPOSE_DIR}"
    docker compose pull "${SERVICE_NAME}" || true
    docker compose up -d --no-deps --force-recreate "${SERVICE_NAME}"
    log_success "Container ${SERVICE_NAME} recriado com sucesso no Docker local!"
else
    log_step "Deploy Docker local ignorado (use './${SCRIPT_NAME:-script-deploy.sh} up' para subir local)"
fi

# 6. Deploy no Kubernetes de Produção (se solicitado 'prod' ou 'full')
if [ "$DEPLOY_PROD" = true ]; then
    log_step "Atualizando helm/values.yaml (tag: \"${BUILD_SHA}\") e aplicando via SSH (porta 22)..."
    sed -i '' "s|^[[:space:]]*tag:.*|  tag: \"${BUILD_SHA}\"|g" "${SCRIPT_DIR}/helm/values.yaml"
    git add "${SCRIPT_DIR}/helm/values.yaml"
    if ! git diff --cached --quiet; then
        git commit -q -m "chore(helm): bump ${SERVICE_NAME} image tag to ${BUILD_SHA}"
        git push origin "${CURRENT_BRANCH}" || true
    fi
    # Deployment por replace (o chart é a verdade): o apply sobre Deployment criado sem a anotação do kubectl
    # mantinha variáveis que saíram do chart e misturava value com valueFrom (deploy inválido na janela de vhost,
    # 05/11/2026). Demais recursos seguem por apply. --save-config deixa a anotação para os próximos.
    HELM_TMP="$(mktemp -d)"
    helm template "${SERVICE_NAME}" "${SCRIPT_DIR}/helm" -n keepguard > "${HELM_TMP}/todos.yaml"
    python3 - "${HELM_TMP}" <<'PY_SEPARAR'
import sys, yaml
d = sys.argv[1]
docs = [x for x in yaml.safe_load_all(open(d + "/todos.yaml")) if x]
yaml.safe_dump_all([x for x in docs if x.get("kind") == "Deployment"], open(d + "/deployment.yaml", "w"))
yaml.safe_dump_all([x for x in docs if x.get("kind") != "Deployment"], open(d + "/resto.yaml", "w"))
PY_SEPARAR
    ssh -o BatchMode=yes root@31.97.175.92 'T=$(mktemp) && cat > "$T" && { kubectl -n keepguard replace --save-config -f "$T" || kubectl -n keepguard create --save-config -f "$T"; }; R=$?; rm -f "$T"; exit $R' < "${HELM_TMP}/deployment.yaml"
    # Services e demais recursos só depois do rollout: um Service que passa a procurar a porta pelo nome
    # ("http") deixa o pod antigo, que não tinha portas com nome, sem tráfego até o novo subir (bff-auth, 05/11/2026).
    ssh -o BatchMode=yes root@31.97.175.92 "kubectl rollout status deployment/${SERVICE_NAME} -n keepguard --timeout=300s"
    if grep -q "kind:" "${HELM_TMP}/resto.yaml"; then
        ssh -o BatchMode=yes root@31.97.175.92 "kubectl apply -n keepguard -f -" < "${HELM_TMP}/resto.yaml"
    fi
    rm -rf "${HELM_TMP}"
    ssh -o BatchMode=yes root@31.97.175.92 "kubectl rollout status deployment/${SERVICE_NAME} -n keepguard --timeout=300s"
    if [ -x "${SCRIPT_DIR}/../../../scripts/verificar-infra.sh" ]; then
        "${SCRIPT_DIR}/../../../scripts/verificar-infra.sh"
    fi
fi

log_success "============================================"
log_success "  Deploy de ${SERVICE_NAME} finalizado com sucesso!"
log_success "============================================"
