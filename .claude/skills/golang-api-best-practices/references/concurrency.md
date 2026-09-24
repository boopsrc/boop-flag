# Concorrência: goroutines, channels e cancelamento

Concorrência é um dos pontos fortes de Go, mas também uma fonte comum de bugs sutis em APIs (goroutine leak, race condition, deadlock). A regra geral: **toda goroutine que você inicia deve ter uma forma clara e garantida de terminar.**

## Goroutines não devem vazar

Uma goroutine "vaza" quando fica bloqueada para sempre (esperando em um channel que nunca recebe valor, por exemplo) porque quem a criou já desistiu de esperar por ela. Em uma API de longa duração, goroutines vazadas se acumulam e consomem memória/CPU até degradar o serviço.

```go
// ERRADO: se ninguém ler de resultCh, esta goroutine vaza para sempre.
func process(items []Item) <-chan Result {
    resultCh := make(chan Result)
    go func() {
        for _, item := range items {
            resultCh <- expensiveOp(item) // bloqueia se ninguém consumir
        }
    }()
    return resultCh
}
```

```go
// CERTO: a goroutine respeita o contexto e pode ser cancelada.
func process(ctx context.Context, items []Item) <-chan Result {
    resultCh := make(chan Result)
    go func() {
        defer close(resultCh)
        for _, item := range items {
            select {
            case resultCh <- expensiveOp(item):
            case <-ctx.Done():
                return
            }
        }
    }()
    return resultCh
}
```

## errgroup para fan-out com propagação de erro

Quando uma requisição precisa disparar múltiplas chamadas concorrentes (ex.: buscar dados em três serviços diferentes) e falhar se qualquer uma falhar, `golang.org/x/sync/errgroup` é o padrão idiomático — ele cancela o contexto compartilhado assim que a primeira goroutine retorna erro, e as outras podem checar `ctx.Done()` para parar mais cedo:

```go
func (s *Service) GetDashboard(ctx context.Context, userID string) (*Dashboard, error) {
    g, ctx := errgroup.WithContext(ctx)

    var profile *Profile
    var orders []Order

    g.Go(func() error {
        var err error
        profile, err = s.profileClient.Get(ctx, userID)
        return err
    })
    g.Go(func() error {
        var err error
        orders, err = s.orderClient.List(ctx, userID)
        return err
    })

    if err := g.Wait(); err != nil {
        return nil, fmt.Errorf("get dashboard: %w", err)
    }
    return &Dashboard{Profile: profile, Orders: orders}, nil
}
```

Isso é preferível a orquestrar `sync.WaitGroup` + channel de erro manualmente na maioria dos casos — menos código, comportamento de cancelamento já correto.

## Channels: dono único, fechamento explícito

- **Só quem escreve em um channel deve fechá-lo.** Fechar um channel do lado do consumidor (ou fechar duas vezes) causa panic.
- **Use `defer close(ch)`** na goroutine produtora para garantir que o channel feche mesmo se a função retornar cedo por erro.
- **Buffered channels evitam bloqueio desnecessário** quando o produtor pode gerar mais rápido que o consumidor processa, mas não são solução para backpressure real — se o consumidor não acompanha, o buffer eventualmente enche e volta a bloquear (o que é o comportamento correto: aplique limite, não ignore o sinal).
- Prefira `select` com `ctx.Done()` em qualquer operação de channel que possa bloquear indefinidamente.

## Estado compartilhado

- Prefira **não compartilhar estado mutável entre goroutines**; passe dados por channel ("share memory by communicating") sempre que a lógica permitir.
- Quando estado compartilhado é necessário (ex.: cache em memória, contador de métricas), proteja com `sync.Mutex`/`sync.RWMutex` e mantenha a seção crítica o menor possível — não faça I/O ou chamadas de rede dentro de um lock.
- Rode testes com `go test -race` regularmente (idealmente no CI) — o detector de race condition do Go pega uma classe inteira de bugs que só apareceriam em produção sob carga.

## Limitar concorrência

Disparar uma goroutine por item de uma lista grande sem limite pode esgotar recursos (conexões de banco, file descriptors, memória). Use um semáforo (channel com buffer, ou `golang.org/x/sync/semaphore`) para limitar quantas operações concorrentes rodam ao mesmo tempo:

```go
sem := make(chan struct{}, 10) // no máximo 10 concorrentes
for _, item := range items {
    sem <- struct{}{}
    go func(item Item) {
        defer func() { <-sem }()
        process(item)
    }(item)
}
```

Isso é especialmente importante ao processar batches vindos de uma requisição de API — sem limite, um único request malicioso ou mal dimensionado pode derrubar o serviço.
