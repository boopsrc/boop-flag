# Estrutura de projeto, injeção de dependência e nomenclatura

## Layout de diretórios

Go não impõe uma estrutura, mas a comunidade convergiu para um layout que funciona bem para APIs:

```
meu-servico/
├── cmd/
│   └── api/
│       └── main.go          # só monta dependências e chama Run(); sem lógica
├── internal/
│   ├── domain/               # entidades e regras de negócio puras, sem I/O
│   │   └── user/
│   │       ├── user.go
│   │       └── service.go
│   ├── http/                 # camada HTTP: handlers, middlewares, router
│   │   ├── handler_user.go
│   │   ├── middleware.go
│   │   └── router.go
│   ├── storage/               # implementações concretas (postgres, redis...)
│   │   └── postgres/
│   │       └── user_repo.go
│   └── config/
│       └── config.go
├── pkg/                       # só se houver código reutilizável por OUTROS projetos
├── api/
│   └── openapi.yaml
├── go.mod
└── Makefile
```

**Por quê `internal/`**: o compilador Go impede que outros módulos importem pacotes sob `internal/`. Isso deixa explícito o que é API pública do seu módulo (se houver) e o que é detalhe de implementação. Para a maioria dos serviços de API (não bibliotecas), quase tudo deve estar em `internal/`.

**Por quê não usar `pkg/` por padrão**: é comum copiar esse padrão de projetos que exportam bibliotecas, mas se seu serviço não é consumido como biblioteca por outro módulo Go, `pkg/` não cumpre função nenhuma — só adiciona um nível de indireção. Use `pkg/` apenas quando o conteúdo é genuinamente reutilizável fora do repositório.

**Organize por domínio, não por camada técnica**: evite pastas soltas como `models/`, `services/`, `controllers/` no nível raiz contendo arquivos de domínios não relacionados. Prefira agrupar por funcionalidade (`internal/domain/user/`, `internal/domain/order/`), o que reduz acoplamento e facilita navegação à medida que o projeto cresce.

## Injeção de dependência

Go não tem (nem precisa de) um framework de DI para a maioria dos casos. O padrão idiomático é:

1. Definir a dependência como uma **interface pequena**, no pacote que a consome (não no pacote que a implementa).
2. Implementar a interface em outro pacote (ex.: `storage/postgres`).
3. Passar a dependência via construtor.

```go
// internal/domain/user/service.go
package user

type Repository interface {
    FindByID(ctx context.Context, id string) (*User, error)
    Save(ctx context.Context, u *User) error
}

type Service struct {
    repo   Repository
    logger *slog.Logger
}

func NewService(repo Repository, logger *slog.Logger) *Service {
    return &Service{repo: repo, logger: logger}
}
```

```go
// cmd/api/main.go
repo := postgres.NewUserRepo(db)
svc := user.NewService(repo, logger)
handler := httpapi.NewUserHandler(svc)
```

Isso permite testar `Service` com um `Repository` fake sem tocar em banco de dados, e trocar a implementação de storage sem alterar a camada de domínio.

**Evite**: variáveis globais mutáveis para dependências (`var DB *sql.DB` no pacote raiz), `init()` que abre conexões, e singletons implícitos. Eles dificultam testes paralelos e escondem o grafo de dependências.

**Evite over-engineering**: não crie uma interface para cada struct "por precaução". Crie a interface no ponto de consumo, quando você realmente precisa trocar a implementação (produção vs. teste) ou desacoplar camadas. Uma interface com um único método e um único implementador real ainda vale a pena se o objetivo é permitir um mock em teste.

## Convenções de nomenclatura

- **Pacotes**: nomes curtos, minúsculos, sem underscore nem plural (`user`, não `users` ou `user_pkg`). O nome do pacote já dá contexto, então evite repetir: `user.New()` é melhor que `user.NewUser()`.
- **Exported vs. unexported**: exporte (`MixedCaps`) só o que precisa ser usado fora do pacote. Mantenha o resto em `mixedCaps` minúsculo.
- **Interfaces**: quando a interface descreve uma ação única, termine em `-er` (`Reader`, `Validator`, `TokenIssuer`). Para interfaces maiores/de domínio, um substantivo claro (`Repository`, `Notifier`) é aceitável.
- **Getters**: não prefixe com `Get`. Use `user.Name()`, não `user.GetName()`. Setters, quando necessários, usam `Set` (`user.SetName(...)`).
- **Erros**: variáveis de erro sentinela começam com `Err` (`ErrNotFound`), tipos de erro terminam em `Error` (`ValidationError`).
- **Arquivos**: `snake_case.go`; arquivos de teste `_test.go` ao lado do código testado.
- **Contexto no nome**: evite nomes genéricos como `data`, `info`, `obj` em código de domínio — prefira `user`, `order`, `payment`, que comunicam intenção.
