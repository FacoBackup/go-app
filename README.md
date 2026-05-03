# HealthGo Backend

Este serviço é responsável por receber e processar dados de telemetria de gases (H2, CH4, H2S) provenientes de dispositivos médicos.

## Como Executar

O serviço pode ser executado de duas formas principais:

### 1. Docker Compose (Recomendado)
A forma mais simples de subir o ambiente completo (Backend + Frontend) é usando o Docker Compose na raiz do projeto:
```bash
docker-compose up --build
```
O Backend estará disponível em `http://localhost:8080` e o Frontend em `http://localhost:3000`.

### 2. Manualmente (Go local)
Na pasta `backend`, execute:
```bash
go run main.go
```

## Testes
Para rodar os testes unitários e de integração (Happy-path, Idempotência, Agregação):
```bash
go test ./internal/...
```

## Endpoints e Exemplos (httpie)

### 1. Enviar Medições (POST)
```bash
http POST :8080/v1/measurements << '[
  {
    "device_id": "dev01",
    "timestamp": "2024-05-03T12:00:00Z",
    "gas": "H2",
    "value_ppm": 15.5,
    "patient_ref": "user123"
  }
]'
```

### 2. Consultar Série Temporal (GET)
Retorna a média por minuto.
```bash
http ":8080/v1/devices/dev01/series?gas=H2&start=2024-05-03T11:00:00Z&end=2024-05-03T13:00:00Z"
```

### 3. Consultar Saúde do Dispositivo (GET)
```bash
http :8080/v1/devices/dev01/health
```

### 4. Saúde do Sistema (Healthcheck)
```bash
http :8080/healthz
http :8080/readyz
```

## Decisões Técnicas e Trade-offs

### 1. Observabilidade: Logs Estruturados e Métricas "Push-based" simplificadas
Implementei logs estruturados em formato JSON utilizando a biblioteca padrão `log/slog`. Para métricas, em vez de configurar um servidor Prometheus completo (que adicionaria complexidade de infraestrutura), optei por um log periódico de métricas de negócio (ex: total de medições em memória).
- **Motivo:** Facilita a integração com agregadores de log como ELK ou Datadog e resolve o requisito de observabilidade sem dependências externas pesadas.

### 2. Validação Rigorosa no Handler
Adicionei validações explícitas no `MeasurementHandler` para rejeitar:
- Payloads malformados (400 Bad Request).
- Valores de `value_ppm` negativos.
- Campos obrigatórios ausentes (`device_id`, `gas`, `timestamp`).
- **Motivo:** Garantir a integridade dos dados antes que eles cheguem à camada de serviço/repositório, fornecendo feedback útil ao cliente.

### 3. Reprodutibilidade com Docker
Utilizei Docker multi-stage para gerar imagens leves e Docker Compose para orquestrar o ambiente.
- **Motivo:** Facilita o onboarding de novos desenvolvedores e garante que o ambiente de teste seja idêntico ao de produção.

### 4. Schema do Banco de Dados (Em Memória)
O armazenamento é feito em um `InMemoryRepository` utilizando:
- `measurements []domain.Measurement`: Um slice para armazenar os dados brutos.
- `dedup map[string]bool`: Um mapa para garantir idempotência, onde a chave é `device_id + timestamp + gas`.
- `sync.RWMutex`: Para garantir thread-safety em acessos concorrentes.

## Trade-offs anteriores mantidos:
- **Agregação em leitura:** Para volumes moderados, a agregação on-the-fly simplifica o sistema. Para escala maciça, usaríamos um banco de séries temporais.
- **CORS Permissivo:** Configurado para aceitar todas as origens para facilitar o teste entre containers Docker (Frontend <-> Backend).
