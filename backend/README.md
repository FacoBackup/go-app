# HealthGo

Este serviço é responsável por receber e processar dados de telemetria de gases (H2, CH4, H2S) provenientes de dispositivos médicos.

## Como Executar (Docker)

A forma recomendada de subir o ambiente completo (Backend + Frontend) é utilizando o Docker Compose na raiz do projeto:

```bash
docker-compose up --build
```

- **Backend:** `http://localhost:8080`
- **Frontend:** `http://localhost:3000`

## Desenvolvimento e Testes

Se desejar rodar os testes unitários ou o backend localmente para desenvolvimento:

### Backend
```bash
go test ./internal/...
go run main.go
```

### Frontend
```bash
cd frontend
npm install
npm start
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

### 1. Observabilidade: Logs Estruturados e Métricas "Push-based"
Implementei logs estruturados em formato JSON utilizando `log/slog`. Para métricas, optei por um log periódico de métricas de negócio (ex: total de medições em memória).
- **Motivo:** Facilita a integração com agregadores de log (ELK/Datadog) sem dependências externas complexas para um MVP.

### 2. Validação Rigorosa no Handler
O handler rejeita payloads malformados, valores negativos ou campos obrigatórios ausentes.
- **Motivo:** Garantir integridade dos dados na entrada da API.

### 3. Reprodutibilidade Total com Docker
O projeto foi estruturado para ser executado via Docker, eliminando problemas de "funciona na minha máquina".
- **Motivo:** Simplifica o deploy e o onboarding de novos desenvolvedores.

### 4. Schema do Banco de Dados (Em Memória)
- `measurements`: Slice para dados brutos.
- `dedup`: Mapa para garantir idempotência (`device_id + timestamp + gas`).
- `sync.RWMutex`: Garantia de thread-safety.

## Trade-offs:
- **Agregação em leitura:** Média calculada sob demanda para simplificar o código.
- **CORS Permissivo:** Facilita a comunicação Frontend <-> Backend em ambiente Docker.
