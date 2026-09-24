# Segurança, configuração e graceful shutdown

## Autenticação e autorização

- **Autentique na borda**, tipicamente em um middleware, validando token/sessão antes de o request chegar ao handler. Não duplique lógica de autenticação dentro de cada handler.
- **Autorização é por rota/recurso, explícita** — não assuma que "autenticado" implica "autorizado a fazer isso". Verifique permissão especificamente para a ação (ex.: usuário só pode editar o próprio recurso, ou precisa de uma role específica).
- **Nunca confie em um ID de usuário vindo do body/query do request** para decidir permissão — derive a identidade do token validado (JWT, sessão), não de um campo que o próprio cliente pode manipular.
- **JWTs**: valide assinatura, `exp` (expiração) e `aud`/`iss` quando aplicável. Nunca decodifique um JWT sem verificar a assinatura antes de confiar no conteúdo.
- **Senhas**: nunca armazene em texto plano; use `bcrypt` ou `argon2`. Nunca logue senha, token ou header `Authorization`.

## Rate limiting

Proteja a API contra abuso (intencional ou não) limitando taxa de requisições, tipicamente por IP, por usuário autenticado, ou por API key:

```go
limiter := rate.NewLimiter(rate.Every(time.Second), 20) // 20 req/s, burst 20

func RateLimit(limiter *rate.Limiter) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

Para múltiplas instâncias do serviço atrás de um load balancer, um limiter em memória por instância só limita por instância — para um limite global confiável, use um store compartilhado (Redis) com um algoritmo como token bucket ou sliding window.

## Sanitização e validação de entrada

- Trate toda entrada externa como hostil até validada (ver `http-handlers-middleware.md` para validação de formato).
- **SQL**: sempre queries parametrizadas, nunca concatenação de string com entrada do usuário.
- **HTML/JS ao renderizar conteúdo gerado por usuário** (se a API alimenta uma página): escape por padrão (`html/template`, não `text/template`, para saída HTML) para evitar XSS.
- **Path traversal**: ao aceitar um nome de arquivo/caminho do cliente, valide contra `..` e caracteres inesperados antes de usar em `os.Open` ou similar; prefira um identificador opaco (UUID) em vez do path real quando possível.
- **Tamanho de payload**: limite o tamanho do body aceito (`http.MaxBytesReader`) e de campos individuais, para evitar exhaustion de memória.

## Headers de segurança

Para APIs consumidas por browsers, configure ao menos:
- `Content-Type` correto e explícito em toda resposta (evita sniffing de MIME type).
- `X-Content-Type-Options: nosniff`.
- CORS restrito a origens conhecidas (`Access-Control-Allow-Origin` explícito, nunca `*` quando a API usa credenciais/cookies).
- TLS obrigatório em produção — nunca aceite tráfego HTTP puro para dados sensíveis; termine TLS no load balancer ou no próprio servidor.

## Configuração via variáveis de ambiente

- **Toda configuração externa (URLs de dependência, credenciais, feature flags, timeouts) vem de variáveis de ambiente**, nunca hardcoded no código.
- **Valide a configuração no boot e falhe rápido** (`log.Fatal` ou retorno de erro de `main` antes de subir o servidor) se algo obrigatório estiver faltando ou for inválido — é preferível o processo não subir a subir e falhar de forma confusa na primeira requisição.
- **Tenha defaults sensatos para o que for opcional** (timeout, porta), mas nunca default para segredo (chave de API, senha de banco) — segredo ausente deve ser erro, não default vazio.
- **Nunca commite segredo no repositório**, nem em arquivo `.env` versionado — use `.env.example` com os nomes das variáveis (sem valores reais) e mantenha o `.env` real no `.gitignore`.

```go
type Config struct {
    Port        string        `env:"PORT" envDefault:"8080"`
    DatabaseURL string        `env:"DATABASE_URL,required"`
    JWTSecret   string        `env:"JWT_SECRET,required"`
    ReadTimeout time.Duration `env:"READ_TIMEOUT" envDefault:"5s"`
}

func Load() (*Config, error) {
    var cfg Config
    if err := env.Parse(&cfg); err != nil {
        return nil, fmt.Errorf("load config: %w", err)
    }
    return &cfg, nil
}
```

(usando `caarlos0/env` como exemplo; a stdlib `os.Getenv` com validação manual também é uma opção totalmente válida para configs pequenas.)

## Graceful shutdown

Um servidor de API deve terminar requisições em andamento antes de encerrar o processo, em vez de cortá-las abruptamente — essencial para deploys sem downtime e para respeitar SIGTERM enviado por orquestradores (Kubernetes, systemd):

```go
func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }

    srv := &http.Server{
        Addr:         ":" + cfg.Port,
        Handler:      router,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    go func() {
        if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            logger.Error("server failed", "error", err)
            os.Exit(1)
        }
    }()

    <-ctx.Done()
    logger.Info("shutting down")

    shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()

    if err := srv.Shutdown(shutdownCtx); err != nil {
        logger.Error("graceful shutdown failed", "error", err)
    }
}
```

Pontos-chave:
- `signal.NotifyContext` captura `SIGTERM`/`SIGINT` e cancela o `ctx` automaticamente — não é preciso um canal de sinal manual.
- `srv.Shutdown(ctx)` para de aceitar novas conexões e espera as em andamento terminarem, até o timeout do `ctx` passado a ele.
- Se o serviço tem outras dependências para fechar (pool de conexão de banco, consumer de fila), feche-as depois do `Shutdown` do servidor HTTP, na mesma sequência de encerramento, para não cortar uma dependência que uma requisição em andamento ainda está usando.
- O timeout do shutdown deve ser maior que o tempo máximo esperado de uma requisição em andamento, mas finito — não deixe o processo pendurado indefinidamente esperando um cliente lento.
