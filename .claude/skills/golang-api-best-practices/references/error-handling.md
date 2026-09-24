# Tratamento de erros e context.Context

## Erros são valores, trate-os explicitamente

Go não tem exceções: um erro é apenas um valor retornado. Isso é uma feature — força a decisão de tratamento a acontecer no ponto onde o erro ocorre, em vez de propagar implicitamente. Consequências práticas:

- **Nunca ignore um erro silenciosamente.** Se realmente não há nada a fazer, descarte explicitamente com `_ = err` e um comentário explicando por quê — isso sinaliza intenção, em vez de parecer um esquecimento.
- **Sempre adicione contexto ao propagar um erro**, com `%w` para preservar a cadeia (permite `errors.Is`/`errors.As` mais acima):

```go
func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    u, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get user %s: %w", id, err)
    }
    return u, nil
}
```

- **Não formate mensagens de erro como frases com maiúscula/pontuação** (`"User not found."`) — convenção Go é minúsculo e sem ponto final, já que a mensagem costuma ser concatenada por quem chama (`"get user: user not found"`).

## Erros de domínio: sentinelas e tipos

Para erros que o chamador precisa distinguir e tratar de forma diferente (não apenas logar), use sentinelas ou tipos comparáveis:

```go
var ErrNotFound = errors.New("not found")

// no repositório
func (r *PostgresUserRepo) FindByID(ctx context.Context, id string) (*User, error) {
    row := r.db.QueryRowContext(ctx, query, id)
    var u User
    if err := row.Scan(&u.ID, &u.Name); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("scan user: %w", err)
    }
    return &u, nil
}

// no handler HTTP, mapeando erro de domínio para status HTTP
u, err := h.svc.GetUser(ctx, id)
switch {
case errors.Is(err, user.ErrNotFound):
    writeError(w, http.StatusNotFound, "user not found")
case err != nil:
    h.logger.Error("get user failed", "error", err, "user_id", id)
    writeError(w, http.StatusInternalServerError, "internal error")
default:
    writeJSON(w, http.StatusOK, u)
}
```

Isso separa duas responsabilidades: erros que o cliente da API precisa entender (mapeados para status HTTP específicos, mensagem segura) e erros internos (logados com detalhe, mas expostos ao cliente de forma genérica — nunca vaze stack trace ou detalhe de infraestrutura na resposta).

Para erros com dados estruturados (ex.: validação com múltiplos campos), use um tipo:

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}
```

## context.Context

`context.Context` carrega prazo, cancelamento e valores de escopo de requisição através de chamadas — é o mecanismo padrão para propagar "essa operação deve parar" em uma API.

Regras práticas:

- **É sempre o primeiro parâmetro**, nomeado `ctx`, em qualquer função que faça I/O (chamada de rede, banco, arquivo) ou que possa ser cancelada: `func (s *Service) GetUser(ctx context.Context, id string) (*User, error)`.
- **Nunca armazene `context.Context` em um struct.** Passe-o explicitamente em cada chamada. Guardar contexto em struct esconde de qual requisição ele veio e facilita usar um contexto "morto" ou de outra requisição por engano.
- **Propague o contexto do handler até o banco de dados.** `http.Request` já tem `req.Context()`, que é cancelado automaticamente se o cliente desconectar. Use `db.QueryContext(ctx, ...)`, não `db.Query(...)`, para que uma query lenta seja abortada se o cliente desistir.
- **Use `context.WithTimeout`/`context.WithDeadline`** para limitar operações que podem travar (chamada a serviço externo, query sem índice), em vez de confiar apenas no timeout do servidor HTTP:

```go
ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
defer cancel()
resp, err := s.paymentClient.Charge(ctx, req)
```

- **`context.Value` é para dados de escopo de requisição que atravessam boundaries de API** (request ID, trace ID, usuário autenticado vindo de um middleware de auth) — não para passar parâmetros de negócio que deveriam ser argumentos explícitos da função. Se uma função precisa de um valor para funcionar corretamente, ele deve ser parâmetro, não `ctx.Value(...)`.
- **Chaves de contexto devem ser tipos não exportados**, nunca `string` crua, para evitar colisão entre pacotes:

```go
type ctxKey int

const requestIDKey ctxKey = iota

func WithRequestID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, requestIDKey, id)
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
    id, ok := ctx.Value(requestIDKey).(string)
    return id, ok
}
```

## Panics

`panic`/`recover` não substitui tratamento de erro em Go — reserve `panic` para erros de programação realmente irrecuperáveis (invariante violada, índice fora do range em código que não deveria permitir isso). Em uma API, sempre tenha um middleware de `recover` no topo da cadeia para transformar um panic inesperado em uma resposta 500 controlada, em vez de derrubar o processo inteiro (ver `http-handlers-middleware.md`).
