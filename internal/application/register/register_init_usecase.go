package register

import (
	"fmt"

	"github.com/keepguard/bff-core/internal/adapters/inbound/http/dto"
	userDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/domain/enums"
	"github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/domain/ports/messaging"
	"go.uber.org/zap"
)

// registerInitUseCaseImpl implementa o caso de uso de inicialização de registro
type registerInitUseCaseImpl struct {
	authClient          client.AuthClient
	userClient          client.UserClient
	companyClient       client.CompanyClient
	communicationClient client.CommunicationClient
	messagePublisher    messaging.MessagePublisher
	logger              *zap.Logger
}

// NewRegisterInitUseCase cria um novo caso de uso de inicialização de registro
func NewRegisterInitUseCase(
	authClient client.AuthClient,
	userClient client.UserClient,
	companyClient client.CompanyClient,
	communicationClient client.CommunicationClient,
	messagePublisher messaging.MessagePublisher,
	logger *zap.Logger,
) RegisterInitUseCase {
	return &registerInitUseCaseImpl{
		authClient:          authClient,
		userClient:          userClient,
		companyClient:       companyClient,
		communicationClient: communicationClient,
		messagePublisher:    messagePublisher,
		logger:              logger,
	}
}

// Execute executa o caso de uso de inicialização de registro
func (uc *registerInitUseCaseImpl) Execute(command appdto.RegisterInitCommand) (dto.RegisterInitResponseDTO, error) {

	// Passo 1: Verificar se a empresa existe consultando o Company Service
	company, err := uc.companyClient.GetByXApplication(command.Context, command.XApplication, command.CorrelationID)
	if err != nil {
		return dto.RegisterInitResponseDTO{}, err
	}

	// Passo 2: Criar requisição de registro
	registerRequest := userDto.MSUserRegisterInitRequestDTO{
		Email:                      command.Email,
		NameFull:                   command.NameFull,
		Password:                   command.Password,
		Phone:                      command.Phone,
		HasAcceptedTermsAndPrivacy: command.HasAcceptedTermsAndPrivacy,
		AcceptedMarketing:          command.AcceptedMarketing,
		IPAddress:                  command.IPAddress,
		UserAgent:                  command.UserAgent,
		Geolocation:                command.Geolocation,
		Type:                       command.Type,
	}

	// Passo 3: Inicializar registro no User Service
	registerResponse, err := uc.userClient.InitRegister(command.Context, registerRequest, command.XApplication, command.CorrelationID)
	if err != nil {
		return dto.RegisterInitResponseDTO{}, err
	}

	// Passo 4: Preparar variáveis para o template
	variables := map[string]string{
		"userName":  command.NameFull,
		"token":     registerResponse.Token,
		"expiresIn": fmt.Sprintf("%d", registerResponse.ExpiresIn/60),
		"appName":   company.Name,
	}

	// Passo 5: Criar requisição de envio de mensagem
	// Converte map[string]string para map[string]interface{}
	interfaceVariables := make(map[string]interface{})
	for k, v := range variables {
		interfaceVariables[k] = v
	}

	messageReq := messaging.MessageDTO{
		XApplication:      command.XApplication,
		XCorrelationID:    command.CorrelationID,
		MessageType:       enums.MessageTypeEmail.String(),
		CommunicationType: enums.CommunicationTypeEmail.String(),
		TemplateType:      enums.TemplateTypeAutenticacaoEmailToken.String(),
		Recipient:         registerResponse.Email,
		CodeUser:          registerResponse.RegistrationSessionID,
		Variables:         interfaceVariables,
	}

	// Passo 6: Enviar mensagem através do Message Publisher (RabbitMQ com fallback HTTP)
	err = uc.messagePublisher.PublishMessage(command.Context, messageReq)
	if err != nil {
		// Não falha o registro se o email não for enviado
		// Log será feito pelo decorator
	}

	// Passo 7: Retornar resposta
	response := dto.RegisterInitResponseDTO{
		RegistrationSessionID: registerResponse.RegistrationSessionID,
		Email:                 registerResponse.Email,
		ExpiresIn:             registerResponse.ExpiresIn,
	}

	return response, nil
}
