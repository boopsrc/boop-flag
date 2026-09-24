---
name: golang-api-best-practices
description: Boas práticas para projetar, implementar e revisar APIs em Go (REST/gRPC), cobrindo estrutura de projeto, tratamento de erros, context.Context, injeção de dependência, handlers HTTP, validação, logging estruturado, concorrência, testes, documentação (OpenAPI/Swagger), versionamento, middlewares, segurança, configuração e graceful shutdown. Use esta skill sempre que o usuário estiver escrevendo, revisando ou planejando uma API em Golang — mesmo que peça algo pontual como "cria um handler", "como organizo esse projeto Go", "revisa esse código Go", ou mencione frameworks como net/http, chi, gin, echo, fiber, gRPC, ou pacotes internos de uma API Go.
---

# Boas práticas para APIs em Go

Esta skill orienta decisões de design e implementação de APIs em Go, do zero ou em código existente. O objetivo não é impor um único framework, e sim aplicar princípios idiomáticos de Go (simplicidade, composição, interfaces pequenas, erros explícitos) ao contexto específico de uma API HTTP ou RPC.

## Como usar

1. Identifique o que o usuário está fazendo: criar um projeto novo, adicionar um endpoint, revisar código existente, ou tirar uma dúvida pontual.
2. Aplique o checklist rápido abaixo para decisões imediatas.
3. Para um tópico específico em profundidade, leia o arquivo de referência correspondente em `references/` — eles contêm exemplos de código e o raciocínio por trás de cada prática. Não é necessário carregar todos de uma vez; leia só o que for relevante para a tarefa atual.
4. Ao revisar código, sinalize desvios das práticas abaixo como sugestões, explicando o porquê (não apenas "isso está errado") — isso ajuda o usuário a decidir quando o desvio é intencional e aceitável.

## Checklist rápido

- **Estrutura**: `cmd/` para binários, `internal/` para código não reutilizável fora do módulo, pacotes coesos por domínio (não por camada técnica tipo `handlers/`, `models/`, `services/` soltos). Ver `references/project-structure.md`.
- **Erros**: sempre `if err != nil { return fmt.Errorf("contexto: %w", err) }`; nunca engolir erro; erros de domínio como tipos/sentinelas comparáveis com `errors.Is`/`errors.As`. Ver `references/error-handling.md`.
- **Context**: todo handler e toda chamada I/O-bound recebe `context.Context` como primeiro parâmetro; nunca guardar `context.Context` em struct; respeitar cancelamento/timeout. Ver `references/error-handling.md`.
- **Injeção de dependência**: dependências (DB, clients, loggers) via interfaces pequenas passadas no construtor, não globais nem singletons implícitos. Ver `references/project-structure.md`.
- **Handlers HTTP**: handlers finos que só fazem parsing/validação de entrada, chamada à camada de domínio, e serialização de saída; lógica de negócio fica fora do handler. Ver `references/http-handlers-middleware.md`.
- **Validação de entrada**: validar e sanitizar tudo que vem de fora (path, query, body, headers) antes de usar; retornar 400 com mensagem clara em caso de falha. Ver `references/http-handlers-middleware.md`.
- **Logging**: `log/slog` estruturado (ou equivalente), com correlação por request ID, nunca logar segredos/PII. Ver `references/http-handlers-middleware.md`.
- **Concorrência**: goroutines sempre com forma clara de terminar (não vazar); `errgroup` para fan-out com propagação de erro; canais com dono único e fechamento explícito. Ver `references/concurrency.md`.
- **Testes**: table-driven tests, testes de handler via `httptest`, mocks via interfaces pequenas, evitar mocks de bibliotecas externas quando dá para testar o contrato real. Ver `references/testing.md`.
- **Documentação de API**: contrato OpenAPI como fonte de verdade (gerado ou gerando código), versionado junto com o código. Ver `references/api-docs-versioning.md`.
- **Versionamento**: versionar a API na URL ou header desde o início (`/v1/...`), planejar depreciação. Ver `references/api-docs-versioning.md`.
- **Middlewares**: cadeia explícita e ordenada (recover → request ID → logging → auth → rate limit → handler); nunca lógica de negócio em middleware. Ver `references/http-handlers-middleware.md`.
- **Segurança**: autenticação/autorização explícitas por rota, rate limiting, sanitização de entrada, headers de segurança, nunca confiar em entrada do cliente. Ver `references/security-config.md`.
- **Configuração**: via variáveis de ambiente com valores default sensatos e validação no boot (fail fast se config obrigatória faltar); nunca hardcode de segredo. Ver `references/security-config.md`.
- **Graceful shutdown**: `http.Server` com `Shutdown(ctx)` acionado por `signal.NotifyContext`, drenando requests em curso antes de encerrar. Ver `references/security-config.md`.
- **Nomenclatura**: seguir convenções Go padrão (`MixedCaps`, sem `Get`/`Set` redundante, nomes de interface terminando em `-er` quando fizer sentido, pacotes com nome curto e sem plural). Ver `references/project-structure.md`.

## Arquivos de referência

| Arquivo | Cobre |
|---|---|
| `references/project-structure.md` | Layout de diretórios, injeção de dependência, convenções de nomenclatura |
| `references/error-handling.md` | Erros idiomáticos, wrapping, sentinelas, `context.Context` |
| `references/http-handlers-middleware.md` | Design de handlers, validação, middlewares, logging estruturado |
| `references/concurrency.md` | Goroutines, channels, `errgroup`, cancelamento |
| `references/testing.md` | Table-driven tests, `httptest`, mocks, testes de integração |
| `references/api-docs-versioning.md` | OpenAPI/Swagger, versionamento de API |
| `references/security-config.md` | Autenticação, rate limiting, sanitização, config via env, graceful shutdown |

## Ao revisar código existente

Não reescreva o projeto inteiro para "se encaixar" nesta skill. Priorize:
1. Problemas de correção/segurança (erro engolido, SQL injection, falta de timeout, goroutine leak).
2. Problemas que afetam manutenibilidade a médio prazo (handler gordo, dependência global, falta de teste no caminho crítico).
3. Só depois, estilo e convenção (nomenclatura, organização de pacote).

Se o projeto já segue uma convenção consistente diferente da sugerida aqui (ex.: usa `pkg/` propositalmente, ou uma estrutura por camada que o time escolheu deliberadamente), respeite a convenção existente em vez de forçar a preferência desta skill.
