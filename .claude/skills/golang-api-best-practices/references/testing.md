# Testes: table-driven tests, httptest e mocks

## Table-driven tests

O padrão idiomático em Go para testar múltiplos cenários da mesma função é uma tabela de casos com `t.Run` por subteste — isso dá saída legível (`TestValidate/empty_email`) e facilita adicionar casos novos sem duplicar estrutura:

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {name: "valid email", email: "a@b.com", wantErr: false},
        {name: "missing @", email: "ab.com", wantErr: true},
        {name: "empty", email: "", wantErr: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("validateEmail(%q) error = %v, wantErr %v", tt.email, err, tt.wantErr)
            }
        })
    }
}
```

Para asserts mais legíveis, `testify/assert` e `testify/require` são aceitáveis e comuns, mas não obrigatórios — a stdlib sozinha já é suficiente e evita uma dependência a mais.

## Testando handlers HTTP com httptest

Teste handlers via `net/http/httptest`, sem precisar subir um servidor real na rede:

```go
func TestUserHandler_Create(t *testing.T) {
    svc := user.NewService(&fakeRepo{}, slog.Default())
    h := NewUserHandler(svc, slog.Default())

    body := strings.NewReader(`{"name":"Ana","email":"ana@example.com"}`)
    req := httptest.NewRequest(http.MethodPost, "/users", body)
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    h.Create(rec, req)

    if rec.Code != http.StatusCreated {
        t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusCreated, rec.Body.String())
    }
}
```

Para testar a cadeia completa (router + middlewares + handler), monte o `http.Handler` real e use `httptest.NewServer` só quando o teste precisar de uma conexão de rede de verdade (ex.: testar timeout, streaming); para a maioria dos casos, chamar o handler diretamente (como acima) é mais rápido e suficiente.

## Mocks via interfaces pequenas

Como as dependências já são interfaces pequenas (ver `project-structure.md`), criar um fake para teste é direto — não é necessário um framework de mock para isso funcionar:

```go
type fakeRepo struct {
    users map[string]*user.User
}

func (f *fakeRepo) FindByID(ctx context.Context, id string) (*user.User, error) {
    u, ok := f.users[id]
    if !ok {
        return nil, user.ErrNotFound
    }
    return u, nil
}

func (f *fakeRepo) Save(ctx context.Context, u *user.User) error {
    f.users[u.ID] = u
    return nil
}
```

Ferramentas de geração de mock (`mockgen`, `moq`) valem a pena quando a interface tem muitos métodos ou muitos testes precisam configurar comportamentos diferentes — mas para interfaces pequenas, um fake escrito à mão costuma ser mais legível.

**O que mockar**: mocke as bordas do seu sistema (repositório, cliente HTTP externo, fila) — não mocke tipos de domínio nem funções puras, que devem ser testados com seu comportamento real.

## Testes de integração

Para testar a implementação real de um repositório (ex.: queries SQL), prefira rodar contra um banco real (via `testcontainers-go`, ou um container Docker no CI) em vez de mockar o driver SQL — isso pega erros de SQL/schema que um mock não pegaria. Separe esses testes dos testes unitários rápidos com uma build tag ou nomenclatura (`_integration_test.go`), para que `go test ./...` no dia a dia continue rápido, e o CI rode a suíte completa incluindo integração.

## O que cobrir com teste, em ordem de prioridade

1. **Regras de domínio** (camada de serviço): é onde a lógica de negócio mora, e onde bugs custam mais caro.
2. **Handlers HTTP**: contrato de request/response, mapeamento de erro para status code, validação de entrada.
3. **Casos de erro**, não só o caminho feliz — é onde a maioria dos bugs de produção realmente vive (timeout, entrada inválida, dependência indisponível).
4. **Concorrência**: rode com `-race` sempre que houver goroutines/channels envolvidos.

Não persiga 100% de cobertura como métrica — meça se os caminhos de decisão importantes (branches de erro, regras de negócio) estão cobertos, não a porcentagem de linhas.
