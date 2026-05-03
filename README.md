# Parte 1

Este serviço é responsável por receber e processar dados de telemetria de gases (H2, CH4, H2S) provenientes de dispositivos médicos.

## Como Executar

O serviço pode ser executado de duas formas principais:

### 1. Backend via Docker (Individual)
Caso queira subir apenas o backend utilizando o Dockerfile próprio:
```bash
cd backend
docker build -t healthgo-backend .
docker run -p 8080:8080 healthgo-backend
```
O Backend estará disponível em `http://localhost:8080`

### 2. Manualmente (Go local)
Na pasta `backend`, execute:
```bash
go run main.go
```

### 3. Frontend via Docker (Individual)
Caso queira subir o frontend (teste mais visual dos endpoints) utilizando o Dockerfile próprio:
```bash
cd frontend
docker build -t healthgo-frontend .
docker run -p 3000:3000 healthgo-frontend
```
O Frontend estará disponível em `http://localhost:3000`.

## Testes
Para rodar os testes unitários e de integração (Happy-path, Idempotência, Agregação):
```bash
go test ./internal/...
```

## Endpoints e Exemplos (curl)

### 1. Enviar Medições (POST)
```bash
curl -X POST http://localhost:8080/v1/measurements -H "Content-Type: application/json" -d '[{"device_id": "dev01","timestamp": "2024-05-03T12:00:00Z","gas": "H2","value_ppm": 15.5,"patient_ref": "user123"}]'
```

### 2. Consultar Série Temporal (GET)
Retorna a média por minuto.
```bash
curl "http://localhost:8080/v1/devices/dev01/series?gas=H2&start=2024-05-03T11:00:00Z&end=2024-05-03T13:00:00Z"
```

### 3. Consultar Saúde do Dispositivo (GET)
```bash
curl http://localhost:8080/v1/devices/dev01/health
```

### 4. Saúde do Sistema (Healthcheck)
```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
```

## Decisões Técnicas e Trade-offs

### 1. Observabilidade: Logs Estruturados e Métricas "Push-based" simplificadas
Implementei logs estruturados em formato JSON utilizando a biblioteca padrão `log/slog`. Para métricas, em vez de configurar um servidor Prometheus completo (que adicionaria complexidade de infraestrutura), optei por um log periódico de métricas de negócio (ex: total de medições em memória).
- **Motivo:** Facilita a integração com agregadores de log como ELK ou Datadog e resolve o requisito de observabilidade sem dependências externas.

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
- **CORS Permissivo:** Configurado para aceitar todas as origens para facilitar o teste entre containers Docker (Frontend <-> Backend).
- "Persistência" em memória por fins de simplicidade
- Log simplista


----

# Parte 2

### 1. Código
1. **Falta de Batch Insert:** 
O código executa uma inserção por iteração no loop, o que gera um overhead de rede e transações no banco.
2. **Ausência de ACID:** 
Se o loop falhar no meio, parte das medições terá sido persistida e a outra não, deixando o sistema em estado inconsistente sem possibilidade de rollback completo.
3. **Tratamento de Erros Genérico:** 
O erro do banco é retornado diretamente sem contexto ou estratégia de novas tentativas (ex: erros temporários de conexão), o que dificulta o rastreamento em produção.

### 2. Testes
Recuso porque "subir cobertura" artificialmente sob pressão de tempo geralmente resulta em testes frágeis (que testam implementação, não comportamento) e gera falsa sensação de segurança. Em 30 minutos, eu focaria em:
1. **Caminhos Críticos:** Teste de integração dos fluxos principais.
2. **Boundaries:** Casos de borda nas validações (ex: valores negativos ou timestamps futuros).
3. **Caminho feliz:** Caminho mais simples de testar com a maior probabilidade de uso no final para certificar que o crucial está funcional.

### 3. Banco de Dados
**(a) Investigação:** Verificar status do servidor do banco ou o tempo médio de resposta, pode ser questão de bloat na base necessitando um vacuum ou apenas uma instabilidade de rede dependendo da infraestrutura.
**(b) Soluções:**
- Executar vacuum (outro comando similar caso não PostgreSQL para limpeza de disco e organizar de dados fragmentados); Tradeoff necessita deixar a aplicação indisponível por alguns minutos até a execução finalizar.
- Reiniciar servidor do banco, irá causar indisponibilidade temporária da aplicação. Só usado em casos mais críticos em caso do servidor estar sobrecarregado e necessitando reset de estado dos processos em andamento.

### 4. Concorrência
A melhor abordagem é delegar a restrição ao banco de dados por uma Unique Key no trio `device_id + timestamp + gas`.
No código, utiliza-se a cláusula `ON CONFLICT DO NOTHING` (ou similar conforme o SQL escolhido). 
Isso evita a necessidade de locks de aplicação ou consultas de verificação ("select before insert") que dobrariam a carga no banco.

### 5. Code Review
**(a) Rejeito:** Possibilidade de introduzir regressões é grande, e 1200 linhas é muito código para revisar de uma só vez. Irei indicar o desenvolvedor a "quebrar" o PR em itens menores, isolados e que mantenha a aplicação funcionando mesmo que a funcionalidade em si esteja parcialmente implementada em cada PR. 
**(b) Comportamento:** Estabeleceria a cultura de pequenos PRs (máximo 200-300 linhas) e a obrigatoriedade de testes para novos comportamentos. Se um refactoring for grande, deve ser quebrado em etapas funcionais menores e revisado incrementalmente para garantir que o "código que outras pessoas conseguem manter" seja a prioridade.

### 6. Produção
1. **0-5 min:** Verifico os logs estruturados e métricas para identificar se o erro 500 está concentrado num tipo de dado específico ou por indisponibilidade parcial de algum componente da infraestrutura.
2. **5-10 min:** Analiso o stacktrace do log para ver se a falha ocorre na persistência (ex: banco lento ou conexão esgotada) ou na validação.
3. **10-15 min:** Se for um aumento de carga, necessário nivelar com o time e a equipe responsavel pela infra (caso exista) para definir mudanças nos servidores. Se for um bug de código recente, realizo o **Rollback** imediato para a versão estável anterior caso possivel, caso contrário, abro o bug para correção prioritária.