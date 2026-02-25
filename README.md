# Video Processor - Clean Architecture

Este projeto implementa um processador de vídeos seguindo os princípios da Clean Architecture.

## 📁 Estrutura do Projeto

```
video-processor/
├── cmd/
│   └── server/
│       └── main.go                 # Entry point da aplicação
├── internal/
│   ├── domain/                     # Camada de Domínio (Entities)
│   │   ├── entities.go            # Entidades de negócio
│   │   └── repository.go          # Interfaces dos repositórios
│   ├── usecase/                    # Camada de Casos de Uso (Use Cases)
│   │   └── video_usecase.go       # Lógica de negócio
│   ├── repository/                 # Camada de Interface (Adapters)
│   │   ├── video_repository.go    # Implementação do repositório de vídeo
│   │   └── frame_repository.go    # Implementação do repositório de frames
│   ├── infrastructure/             # Camada de Infraestrutura
│   │   └── ffmpeg_processor.go    # Implementação do processador FFmpeg
│   └── delivery/                   # Camada de Entrega (Frameworks & Drivers)
│       └── http/
│           ├── video_handler.go   # Handlers HTTP
│           ├── router.go          # Configuração de rotas
│           └── templates.go       # Templates HTML
├── pkg/
│   └── utils/
│       └── filesystem.go           # Utilitários
├── go.mod
└── README.md
```

## 🏗️ Arquitetura

### Camadas da Clean Architecture

#### 1. **Domain Layer (Entities)**
- **Localização**: `internal/domain/`
- **Responsabilidade**: Contém as entidades de negócio e interfaces dos repositórios
- **Características**:
  - Não depende de nenhuma outra camada
  - Define as regras de negócio essenciais
  - Interfaces dos repositórios (Dependency Inversion Principle)

#### 2. **Use Case Layer**
- **Localização**: `internal/usecase/`
- **Responsabilidade**: Contém a lógica de negócio da aplicação
- **Características**:
  - Depende apenas da camada Domain
  - Orquestra o fluxo de dados entre as camadas
  - Implementa os casos de uso do sistema

#### 3. **Interface Adapters Layer**
- **Localização**: `internal/repository/` e `internal/delivery/`
- **Responsabilidade**: Converte dados entre as camadas externas e internas
- **Componentes**:
  - **Repositories**: Implementações concretas das interfaces de repositório
  - **HTTP Handlers**: Adaptadores para requisições HTTP

#### 4. **Infrastructure Layer**
- **Localização**: `internal/infrastructure/`
- **Responsabilidade**: Implementações de ferramentas e serviços externos
- **Características**:
  - FFmpeg processor
  - Serviços de terceiros
  - Frameworks externos

#### 5. **Frameworks & Drivers Layer**
- **Localização**: `cmd/` e `internal/delivery/`
- **Responsabilidade**: Entry points e configuração de frameworks
- **Componentes**:
  - Main application
  - Router configuration
  - Middlewares

## 🔄 Fluxo de Dados

```
HTTP Request → Router → Handler → Use Case → Repository → Infrastructure
                  ↓         ↓         ↓            ↓
              Templates  Validation Domain     Database/FFmpeg
```

## 💡 Princípios Aplicados

### 1. **Dependency Rule**
- As dependências apontam sempre para dentro (em direção ao domínio)
- Camadas externas dependem de camadas internas
- Camadas internas não conhecem camadas externas

### 2. **Dependency Inversion Principle (DIP)**
- Use cases dependem de interfaces, não de implementações concretas
- Repositories implementam interfaces definidas no domain

### 3. **Single Responsibility Principle (SRP)**
- Cada camada tem uma única responsabilidade
- Handlers apenas lidam com HTTP
- Use cases apenas lógica de negócio
- Repositories apenas persistência

### 4. **Interface Segregation Principle (ISP)**
- Interfaces específicas para cada tipo de operação
- VideoRepository, FrameRepository, VideoProcessor

## 🚀 Como Executar

1. **Instalar dependências**:
```bash
go mod download
```

2. **Executar a aplicação**:
```bash
go run cmd/server/main.go
```

3. **Acessar**:
```
http://localhost:8080
```

## 📦 Dependências

- **Gin**: Framework HTTP
- **FFmpeg**: Processamento de vídeo (deve estar instalado no sistema)

## 🧪 Testabilidade

A estrutura Clean Architecture facilita:

- **Unit Tests**: Testar use cases isoladamente usando mocks
- **Integration Tests**: Testar repositórios com banco de dados de teste
- **E2E Tests**: Testar handlers HTTP

Exemplo de teste de use case:
```go
// Mock repository
type MockVideoRepository struct{}
func (m *MockVideoRepository) Save(file multipart.File, filename string) (string, error) {
    return "/fake/path/video.mp4", nil
}

// Test
func TestProcessVideo(t *testing.T) {
    mockRepo := &MockVideoRepository{}
    useCase := usecase.NewVideoUseCase(mockRepo, ...)
    // Test logic...
}
```

## 📝 Benefícios da Arquitetura

1. **Independência de Frameworks**: Fácil trocar Gin por outro framework
2. **Testabilidade**: Cada camada pode ser testada isoladamente
3. **Independência de UI**: Pode adicionar CLI, gRPC sem afetar lógica de negócio
4. **Independência de Banco de Dados**: Trocar implementação sem afetar use cases
5. **Manutenibilidade**: Código organizado e fácil de entender
6. **Escalabilidade**: Fácil adicionar novos recursos

## 🔧 Extensões Futuras

- Adicionar testes unitários e de integração
- Implementar logging estruturado
- Adicionar métricas e monitoring
- Implementar circuit breaker para FFmpeg
- Adicionar suporte a diferentes formatos de output
- Implementar sistema de filas para processamento assíncrono
