# Handlers HTTP, validação, middlewares e logging

## Handlers finos

Um handler HTTP deve fazer só quatro coisas: decodificar a entrada, validar, chamar a camada de domínio, codificar a saída. Lógica de negócio (regras, cálculos, decisões) pertence à camada de domínio/serviço, não ao handler — isso permite testar a regra sem subir um servidor HTTP e reutilizá-la em outro transporte (gRPC, CLI, job) se necessário.

```go
type UserHandler struct {
    svc    *user.Service
    logger *slog.Logger
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    var req createUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    if err := req.Validate(); err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    u, err := h.svc.CreateUser(ctx, req.ToDomain())
    switch {
    case errors.Is(err, user.ErrAlreadyExists):
        writeError(w, http.StatusConflict, "user already exists")
    case err != nil:
        h.logger.ErrorContext(ctx, "create user failed", "error", err)
        writeError(w, http.StatusInternalServerError, "internal error")
    default:
        writeJSON(w, http.StatusCreated, toResponse(u))
    }
}
```

Pontos-chave:
- **Sempre limite o tamanho do body** lido (`http.MaxBytesReader`) para evitar que um cliente mande um payload gigante e esgote memória.
- **Sempre defina timeouts no `http.Server`** (`ReadTimeout`, `WriteTimeout`, `IdleTimeout`) — sem eles, um cliente lento pode segurar uma conexão indefinidamente.
- **Separe o tipo de request/response (DTO) do tipo de domínio.** Decodificar direto para a struct de domínio acopla o contrato da API ao modelo interno e vaza campos internos sem querer.

## Validação de entrada

Tudo que vem de fora (body, query params, path params, headers) é não confiável até validado. Duas camadas:

1. **Validação de formato/sintaxe** na borda (handler ou DTO): campos obrigatórios presentes, tipos corretos, tamanhos dentro do limite, formato de e-mail/data válido. Bibliotecas como `go-playground/validator` ajudam via struct tags, mas validação manual explícita também é aceitável e às vezes mais clara:

```go
func (r createUserRequest) Validate() error {
    if strings.TrimSpace(r.Name) == "" {
        return &ValidationError{Field: "name", Message: "required"}
    }
    if !isValidEmail(r.Email) {
        return &ValidationError{Field: "email", Message: "invalid format"}
    }
    return nil
}
```

2. **Validação de regra de negócio** na camada de domínio (ex.: "e-mail já cadastrado", "saldo insuficiente") — não depende do transporte e deve ser testada isoladamente.

Nunca confie em validação feita só no frontend/cliente. A API é o boundary de confiança real.

Ao interpolar entrada do usuário em queries SQL, **sempre use queries parametrizadas** (`db.QueryContext(ctx, "... WHERE id = $1", id)`), nunca concatenação de string — isso previne SQL injection por construção, não por disciplina.

## Middlewares

Middlewares em Go seguem o padrão `func(http.Handler) http.Handler`, compondo funcionalidades transversais sem tocar na lógica do handler. Ordem importa — a ordem recomendada, de fora para dentro:

```go
router.Use(
    middleware.Recover(logger),      // captura panic, sempre primeiro
    middleware.RequestID(),          // gera/propaga X-Request-ID
    middleware.Logging(logger),      // loga request/response com request ID
    middleware.Timeout(30*time.Second),
    middleware.RateLimit(limiter),
    middleware.Auth(tokenVerifier),  // autenticação, o mais perto possível do handler
)
```

Regras:
- **`recover` sempre primeiro na cadeia**, para capturar panic de qualquer middleware/handler abaixo dele e devolver 500 em vez de derrubar o processo.
- **Um middleware faz uma coisa.** Não misture logging com autenticação no mesmo middleware — dificulta testar e reordenar.
- **Nunca lógica de negócio em middleware.** Middleware é para o que é verdadeiramente transversal (auth, logging, rate limit, CORS, compressão); regra de negócio pertence à camada de domínio.
- **Auth por rota, não global**, quando a API tem rotas públicas e privadas — aplique o middleware de auth só no grupo de rotas que precisa, em vez de um `if` dentro do handler.

## Logging estruturado

Use logging estruturado (`log/slog` na stdlib desde Go 1.21, ou `zap`/`zerolog`) em vez de `fmt.Println`/`log.Printf` — permite filtrar e agregar logs por campo em produção.

```go
logger.InfoContext(ctx, "user created",
    "user_id", u.ID,
    "request_id", requestIDFromContext(ctx),
)
```

- **Inclua um request ID** (gerado no middleware, propagado via contexto) em todo log de uma requisição, para correlacionar múltiplas linhas de log com uma única chamada.
- **Nunca logue segredos, senhas, tokens ou PII sensível** (CPF, cartão de crédito) em texto plano. Se precisar logar um identificador de usuário, prefira um ID interno a dado pessoal.
- **Erros esperados (validação, not found) geram log em nível `info`/`warn`**, não `error` — reserve `error` para falhas inesperadas que exigem atenção operacional (erro de infraestrutura, bug).
- **Não logue e retorne o erro ao mesmo tempo em múltiplas camadas** — isso duplica a mesma falha várias vezes no log. Logue uma vez, no ponto mais alto que tem contexto suficiente para decidir a resposta (tipicamente o handler).
