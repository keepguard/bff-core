#!/bin/bash

# Script automatizado de deploy para BFF Core
# Fluxo: SNAPSHOT → RELEASE → Próxima versão
# Uso: ./script-deploy-github-bff-core.sh [version] [up]
# Exemplos:
#   ./script-deploy-github-bff-core.sh 1.0.0 up    # Versão + Docker
#   ./script-deploy-github-bff-core.sh 1.0.0       # Versão + GitHub Only
#   ./script-deploy-github-bff-core.sh up          # Auto + Docker
#   ./script-deploy-github-bff-core.sh             # Auto + GitHub Only

# set -e removido para permitir que o script continue mesmo com erros no deploy

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Funções de log colorido
log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "${BLUE}[STEP]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_debug() { echo -e "${PURPLE}[DEBUG]${NC} $1"; }

# Configurações
SERVICE_NAME="bff-core"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="$PROJECT_ROOT/application.yml"
DOCKER_COMPOSE_DIR="$(cd "$PROJECT_ROOT/../../../docker" && pwd)"
DOCKER_COMPOSE_FILE="$DOCKER_COMPOSE_DIR/docker-compose.yml"
DOCKERFILE="$PROJECT_ROOT/deploy/Dockerfile"
BACKUP_DIR=".deploy-backup"

# Configurações do GitHub
GITHUB_REGISTRY="ghcr.io"
GITHUB_USERNAME="keepguard"
NAMESPACE="keepguard"

# Parâmetros (serão definidos pela função detect_deploy_mode)
RELEASE_VERSION=""
DEPLOY_MODE=""
NON_INTERACTIVE=""

# Função para detectar modo de operação
detect_deploy_mode() {
    if [ "$1" = "--auto" ] || [ "$2" = "--auto" ] || [ "$3" = "--auto" ]; then
        NON_INTERACTIVE="true"
    fi
    if [ "$1" = "up" ] || [ "$2" = "up" ]; then
        if [ "$1" != "up" ] && [ "$1" != "--auto" ]; then
            RELEASE_VERSION="$1"
            DEPLOY_MODE="up"
            log_info "Modo detectado: Versão $1 + Docker"
        else
            RELEASE_VERSION=""
            DEPLOY_MODE="up"
            log_info "Modo detectado: Auto + Docker (versão do config)"
        fi
    elif [ -n "$1" ] && [ "$1" != "up" ] && [ "$1" != "--auto" ]; then
        # Modo 2: ./script 1.0.0
        RELEASE_VERSION="$1"
        DEPLOY_MODE="GitHub-only"
        log_info "Modo detectado: Versão $1 + GitHub Only"
    else
        # Modo 4: ./script
        RELEASE_VERSION=""
        DEPLOY_MODE="GitHub-only"
        log_info "Modo detectado: Auto + GitHub Only (versão do config)"
    fi
}

# Função para extrair versão atual do config
get_current_version() {
    grep 'version:' "$CONFIG_FILE" | sed 's/.*version: *"\(.*\)".*/\1/' | tr -d ' 	'
}

# Função para extrair versão base (sem -SNAPSHOT)
get_base_version() {
    local version=$(get_current_version)
    echo "${version%-SNAPSHOT}"
}

# Função para verificar se versão é SNAPSHOT
is_snapshot_version() {
    local version=$(get_current_version)
    [[ "$version" == *"-SNAPSHOT" ]]
}

# Função para verificar se versão já existe no GitHub
check_version_exists() {
    local version=$1
    local is_snapshot=$2
    
    log_info "Verificando se versão $version já existe no GitHub..."
    
    if [ "$is_snapshot" = "true" ]; then
        local url="$GITHUB_REGISTRY/$NAMESPACE/$SERVICE_NAME:$version"
    else
        local url="$GITHUB_REGISTRY/$NAMESPACE/$SERVICE_NAME:$version"
    fi
    
    # Verificar se imagem existe no GitHub Packages
    if docker manifest inspect "$url" > /dev/null 2>&1; then
        log_warn "⚠️  Versão $version já existe no GitHub!"
        return 0  # Existe
    else
        log_info "✅ Versão $version não existe no GitHub"
        return 1  # Não existe
    fi
}

# Função para incrementar versão
increment_version() {
    local version=$1
    local major=$(echo $version | cut -d. -f1)
    local minor=$(echo $version | cut -d. -f2)
    local patch=$(echo $version | cut -d. -f3)
    
    patch=$((patch + 1))
    echo "${major}.${minor}.${patch}"
}

# Função para atualizar versão no config
update_config_version() {
    local new_version=$1
    log_info "Atualizando config para versão: $new_version"
    
    # Atualizar a versão no arquivo de configuração
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s|version: \".*\"|version: \"$new_version\"|g" "$CONFIG_FILE"
    else
        sed -i "s|version: \".*\"|version: \"$new_version\"|g" "$CONFIG_FILE"
    fi
    
    log_success "Config atualizado para: $new_version"
}

# Função para fazer backup
create_backup() {
    log_step "Criando backup do estado atual..."
    
    rm -rf "$BACKUP_DIR"
    mkdir -p "$BACKUP_DIR"
    
    cp "$CONFIG_FILE" "$BACKUP_DIR/"
    cp "$DOCKER_COMPOSE_FILE" "$BACKUP_DIR/" 2>/dev/null || true
    
    log_success "Backup criado em: $BACKUP_DIR"
}

# Função para restaurar backup
restore_backup() {
    log_warn "Restaurando backup..."
    
    if [ -d "$BACKUP_DIR" ]; then
        cp "$BACKUP_DIR/application.yml" "$CONFIG_FILE"
        cp "$BACKUP_DIR/docker-compose.yml" "$DOCKER_COMPOSE_FILE" 2>/dev/null || true
        log_success "Backup restaurado"
    fi
}

# Função para limpar backup
cleanup_backup() {
    log_info "Limpando arquivos de backup..."
    rm -rf "$BACKUP_DIR"
    log_success "Backup removido"
}

# Função para verificar se serviço existe no docker-compose
check_service_exists() {
    if [ ! -f "$DOCKER_COMPOSE_FILE" ]; then
        return 1
    fi
    
    # Verificar se o serviço existe no docker-compose.yml
    if grep -q "^\s*$SERVICE_NAME:" "$DOCKER_COMPOSE_FILE"; then
        return 0  # Existe
    else
        return 1  # Não existe
    fi
}

# Função para verificar pré-requisitos
check_prerequisites() {
    log_step "Verificando pré-requisitos..."
    
    # Verificar Docker
    if ! docker info > /dev/null 2>&1; then
        log_error "Docker não está rodando!"
        exit 1
    fi
    log_success "Docker está rodando"
    
    # Verificar Docker Compose
    if ! command -v docker-compose > /dev/null 2>&1; then
        log_error "Docker Compose não está instalado!"
        exit 1
    fi
    log_success "Docker Compose está disponível"
    
    # Verificar arquivo de configuração
    if [ ! -f "$CONFIG_FILE" ]; then
        log_error "Arquivo de configuração não encontrado: $CONFIG_FILE"
        exit 1
    fi
    log_success "Arquivo de configuração encontrado"
    
    # Verificar Dockerfile
    if [ ! -f "$DOCKERFILE" ]; then
        log_error "Dockerfile não encontrado: $DOCKERFILE"
        exit 1
    fi
    log_success "Dockerfile encontrado"
    
    # Verificar arquivo docker-compose (apenas se modo "up")
    if [ "$DEPLOY_MODE" = "up" ]; then
        if [ ! -f "$DOCKER_COMPOSE_FILE" ]; then
            log_error "Arquivo docker-compose.yml não encontrado: $DOCKER_COMPOSE_FILE"
            exit 1
        fi
        log_success "Arquivo docker-compose.yml encontrado"
        
        # Verificar se o serviço existe no docker-compose
        if ! check_service_exists; then
            log_warn "⚠️  Serviço $SERVICE_NAME não encontrado no docker-compose.yml"
            log_info "Adicione o serviço ao docker-compose.yml primeiro:"
            log_info "  - Adicione a configuração do serviço $SERVICE_NAME"
            log_info "  - Configure a imagem: ghcr.io/keepguard/$SERVICE_NAME:VERSION"
            log_info "  - Configure portas, dependências, etc."
            log_info ""
            log_info "Continuando apenas com deploy no GitHub..."
            log_info "Execute novamente com modo 'up' após adicionar o serviço ao docker-compose.yml"
            return 1
        fi
        log_success "Serviço $SERVICE_NAME encontrado no docker-compose.yml"
    fi
}

# Função para fazer build da imagem
build_image() {
    local version=$1
    log_step "Fazendo build da imagem $SERVICE_NAME:$version..."
    
    # Build a partir do diretório raiz do projeto usando o Dockerfile em deploy/
    docker build -f "$DOCKERFILE" -t "$SERVICE_NAME:$version" .
    
    if [ $? -eq 0 ]; then
        log_success "Build da imagem $SERVICE_NAME:$version concluído"
    else
        log_error "Erro no build da imagem $SERVICE_NAME:$version"
        exit 1
    fi
}

# Função para fazer push para GitHub
push_to_github() {
    local version=$1
    log_step "Fazendo push para GitHub Packages..."
    
    # Login no GitHub
    log_info "Fazendo login no GitHub Packages..."
    
    # Tag da imagem para o GitHub
    local github_image="$GITHUB_REGISTRY/$NAMESPACE/$SERVICE_NAME:$version"
    log_info "Fazendo tag da imagem: $SERVICE_NAME:$version -> $github_image"
    docker tag "$SERVICE_NAME:$version" "$github_image"
    
    # Push da imagem
    log_info "Fazendo push da imagem para o GitHub..."
    if docker push "$github_image"; then
        log_success "Push para GitHub concluído: $github_image"
        
        # Limpar tag local do GitHub
        docker rmi "$github_image" 2>/dev/null || true
    else
        log_error "Falha no push para o GitHub"
        exit 1
    fi
}

# Função para atualizar docker-compose
update_docker_compose() {
    local version=$1
    log_step "Atualizando docker-compose.yml..."
    
    cd "$DOCKER_COMPOSE_DIR"
    
    # Backup do arquivo original
    cp docker-compose.yml docker-compose.yml.backup
    
    # Atualizar a versão da imagem
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # Atualizar para usar GitHub Packages (macOS)
        sed -i '' "s|ghcr.io/keepguard/$SERVICE_NAME:[^ ]*|ghcr.io/keepguard/$SERVICE_NAME:$version|g" docker-compose.yml
        # Atualizar pull_policy para always (apenas para este serviço)
        sed -i '' "/$SERVICE_NAME:/,/restart:/ s|pull_policy: never|pull_policy: always|" docker-compose.yml
    else
        # Atualizar para usar GitHub Packages (Linux)
        sed -i "s|ghcr.io/keepguard/$SERVICE_NAME:[^ ]*|ghcr.io/keepguard/$SERVICE_NAME:$version|g" docker-compose.yml
        # Atualizar pull_policy para always (apenas para este serviço)
        sed -i "/$SERVICE_NAME:/,/restart:/ s|pull_policy: never|pull_policy: always|" docker-compose.yml
    fi
    
    log_success "Docker-compose atualizado para versão: $version"
    
    # Voltar para o diretório do serviço
    cd "$PROJECT_ROOT"
}

# Função para fazer deploy no Docker Compose
deploy_service() {
    local version=$1
    log_step "Fazendo deploy no Docker Compose..."
    
    cd "$DOCKER_COMPOSE_DIR"
    
    # Parar e remover o serviço completamente para garantir recriação
    log_info "Parando e removendo serviço $SERVICE_NAME completamente..."
    docker-compose stop "$SERVICE_NAME" 2>/dev/null || true
    docker-compose rm -f "$SERVICE_NAME" 2>/dev/null || true
    
    # Iniciar o serviço com recriação completa
    log_info "Iniciando serviço $SERVICE_NAME..."
    docker-compose up -d "$SERVICE_NAME"
    
    if [ $? -eq 0 ]; then
        log_success "Deploy do serviço $SERVICE_NAME concluído"
        
        # Aguardar um pouco para o serviço inicializar
        log_info "Aguardando inicialização do serviço..."
        sleep 5
        
        # Verificar saúde do serviço
        if check_service_health; then
            log_success "Serviço $SERVICE_NAME está saudável e funcionando"
        else
            log_error "Serviço $SERVICE_NAME não está respondendo corretamente"
        fi
    else
        log_error "Erro no deploy do serviço $SERVICE_NAME"
    fi
    
    # Sempre remover backup do docker-compose (independente do resultado)
    rm -f docker-compose.yml.backup
    
    # Limpeza final: remover diretório de backup se existir
    rm -rf "$BACKUP_DIR"
    
    # Voltar para o diretório do serviço
    cd "$PROJECT_ROOT"
}

# Função para verificar saúde do serviço
check_service_health() {
    local max_attempts=10
    local attempt=1
    
    log_step "Verificando saúde do serviço $SERVICE_NAME..."
    
    if [ -f "$DOCKER_COMPOSE_FILE" ]; then
        cd "$DOCKER_COMPOSE_DIR"
        while [ $attempt -le $max_attempts ]; do
            if docker-compose ps "$SERVICE_NAME" | grep -q "healthy"; then
                log_success "Serviço $SERVICE_NAME está saudável"
                cd "$PROJECT_ROOT"
                return 0
            elif docker-compose ps "$SERVICE_NAME" | grep -q "Up"; then
                log_success "Serviço $SERVICE_NAME está em execução"
                cd "$PROJECT_ROOT"
                return 0
            else
                log_warn "Serviço $SERVICE_NAME ainda não está rodando... (tentativa $attempt/$max_attempts)"
            fi
            
            sleep 2
            ((attempt++))
        done
        cd "$PROJECT_ROOT"
    fi
    return 0
}

# Função para mostrar informações finais
show_final_info() {
    local release_version=$1
    local next_version=$2
    local mode=$3
    
    log_success "🎉 Deploy do $SERVICE_NAME:$release_version concluído com sucesso!"
    echo ""
    echo -e "${CYAN}📋 Resumo do Deploy:${NC}"
    echo "  - Versão Release: $release_version"
    echo "  - Próxima Versão: $next_version"
    echo "  - Modo: $mode"
    echo "  - GitHub Packages: $GITHUB_REGISTRY/$NAMESPACE/$SERVICE_NAME"
    echo ""
    
    if [ "$mode" = "up" ]; then
        echo -e "${CYAN}🔗 Informações de Acesso:${NC}"
        echo "  - Health Check: http://localhost:8582/health"
        echo "  - Logs: docker-compose logs -f $SERVICE_NAME"
        echo ""
    fi
    
    echo -e "${CYAN}🚀 Próximos Passos:${NC}"
    echo "  - Config atualizado para: $next_version"
    echo "  - Pronto para desenvolvimento"
    echo "  - Execute: git add . && git commit -m 'Release $release_version'"
}

# Função principal
main() {
    echo -e "${CYAN}🚀 Deploy Automatizado - BFF Core${NC}"
    echo "=========================================="
    
    # Detectar modo de operação
    detect_deploy_mode "$@"
    
    # Determinar versão de release
    if [ -z "$RELEASE_VERSION" ]; then
        RELEASE_VERSION=$(get_base_version)
        log_info "Versão de release não especificada. Usando versão base: $RELEASE_VERSION"
    fi
    
    # Verificar se config está em versão SNAPSHOT
    if [ "$NON_INTERACTIVE" != "true" ] && ! is_snapshot_version; then
        log_warn "⚠️  Config não está em versão SNAPSHOT!"
        log_info "Versão atual: $(get_current_version)"
        log_info "Para deploy correto, o config deve estar em versão SNAPSHOT"
        log_info "Exemplo: 1.0.0-SNAPSHOT"
        echo ""
        read -p "Deseja continuar mesmo assim? (y/N): " -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Deploy cancelado pelo usuário"
            exit 0
        fi
        log_warn "Continuando com versão não-SNAPSHOT..."
    fi
    
    # Verificar se a versão é válida
    if [[ ! "$RELEASE_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        log_error "Versão inválida: $RELEASE_VERSION. Use o formato: X.Y.Z"
        exit 1
    fi
    
    # Verificar se versão RELEASE já existe
    if [ "$NON_INTERACTIVE" != "true" ] && check_version_exists "$RELEASE_VERSION" "false"; then
        log_warn "⚠️  Versão RELEASE $RELEASE_VERSION já existe no GitHub!"
        echo ""
        echo "Opções disponíveis:"
        echo "1. Sobrescrever versão existente (não recomendado)"
        echo "2. Usar próxima versão disponível"
        echo "3. Cancelar deploy"
        echo ""
        read -p "Escolha uma opção (1/2/3): " -n 1 -r
        echo ""
        
        case $REPLY in
            1)
                log_warn "⚠️  Sobrescrevendo versão existente..."
                ;;
            2)
                # Encontrar próxima versão disponível
                local next_version=$RELEASE_VERSION
                while check_version_exists "$next_version" "false"; do
                    next_version=$(increment_version "$next_version")
                done
                log_info "Usando próxima versão disponível: $next_version"
                RELEASE_VERSION=$next_version
                ;;
            3)
                log_info "Deploy cancelado pelo usuário"
                exit 0
                ;;
            *)
                log_error "Opção inválida. Deploy cancelado."
                exit 1
                ;;
        esac
    fi
    
    # Calcular próxima versão
    NEXT_VERSION=$(increment_version "$RELEASE_VERSION")
    
    echo "Versão Atual: $(get_current_version)"
    echo "Versão Release: $RELEASE_VERSION"
    echo "Próxima Versão: $NEXT_VERSION"
    echo "=========================================="
    echo ""
    
    # Executar pipeline de deploy
    if ! check_prerequisites; then
        # Se check_prerequisites retornou erro (serviço não existe), continuar apenas com GitHub
        log_info "Continuando apenas com deploy no GitHub..."
        DEPLOY_MODE="GitHub-only"
    fi
    
    create_backup
    
    # Deploy do SNAPSHOT
    log_step "1. Deploy do SNAPSHOT para GitHub..."
    SNAPSHOT_VERSION="${RELEASE_VERSION}-SNAPSHOT"
    build_image "$SNAPSHOT_VERSION"
    push_to_github "$SNAPSHOT_VERSION"
    
    # Deploy do RELEASE
    log_step "2. Deploy do RELEASE para GitHub..."
    build_image "$RELEASE_VERSION"
    push_to_github "$RELEASE_VERSION"
    
    # Deploy no Docker Compose (apenas se modo "up" e serviço existe)
    if [ "$DEPLOY_MODE" = "up" ]; then
        log_step "3. Atualizando docker-compose e fazendo deploy..."
        update_docker_compose "$RELEASE_VERSION"
        deploy_service "$RELEASE_VERSION"
        DEPLOY_RESULT=$?
        
        # Verificar saúde do serviço
        if [ $DEPLOY_RESULT -eq 0 ] && check_service_health; then
            log_success "✅ Deploy completo concluído com sucesso!"
        else
            log_error "❌ Deploy falhou, mas continuando com atualização do config..."
        fi
    else
        log_step "3. Deploy no GitHub concluído (modo GitHub-only)"
    fi
    
    # Atualizar config para próxima versão
    log_step "4. Atualizando config para próxima versão..."
    update_config_version "${NEXT_VERSION}-SNAPSHOT"
    
    # Limpeza
    log_step "5. Limpando arquivos temporários..."
    cleanup_backup
    
    # Mostrar informações finais
    show_final_info "$RELEASE_VERSION" "${NEXT_VERSION}-SNAPSHOT" "$DEPLOY_MODE"
}

# Tratamento de erros
trap 'log_error "Script interrompido. Restaurando backup..."; restore_backup; cleanup_backup; exit 1' INT TERM

# Executar função principal
main "$@"

