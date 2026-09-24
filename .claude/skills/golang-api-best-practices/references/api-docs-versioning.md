# Documentação de API (OpenAPI/Swagger) e versionamento

## O contrato como fonte de verdade

Uma API sem contrato documentado obriga cada consumidor a ler o código-fonte (ou adivinhar) para saber o que esperar. Mantenha um arquivo OpenAPI (`api/openapi.yaml`) versionado junto com o código, e escolha uma das duas direções de forma consistente no projeto:

- **Design-first**: escreve-se o `openapi.yaml` antes do código, e gera-se structs/interfaces Go a partir dele (`oapi-codegen`, `ogen`). Garante que a implementação não pode divergir do contrato sem regenerar.
- **Code-first**: anota-se o código Go com comentários (`swaggo/swag`) e gera-se o `openapi.yaml` a partir deles. Mais rápido para prototipar, mas exige disciplina para manter as anotações atualizadas.

Qualquer uma é válida; o que não é aceitável é documentação manual desincronizada do código, que rapidamente fica desatualizada e passa a enganar quem consome a API.

Exemplo mínimo de anotação `swaggo`:

```go
// CreateUser godoc
// @Summary      Cria um novo usuário
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body createUserRequest true "Dados do usuário"
// @Success      201 {object} userResponse
// @Failure      400 {object} errorResponse
// @Failure      409 {object} errorResponse
// @Router       /v1/users [post]
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) { ... }
```

O contrato deve documentar, no mínimo: todos os status codes possíveis (não só o sucesso), o schema de erro, campos obrigatórios vs. opcionais, e exemplos de request/response.

## Versionamento de API

Defina a estratégia de versionamento **antes** do primeiro release público, porque trocar depois quebra todo consumidor existente. As duas abordagens mais comuns:

- **Versionamento na URL** (`/v1/users`, `/v2/users`): mais simples de rotear, visível em logs e no navegador, fácil de explicar a consumidores externos. É a escolha mais comum para APIs públicas.
- **Versionamento por header** (`Accept: application/vnd.myapi.v2+json`): mantém a URL estável, mas exige mais disciplina de cliente e é menos óbvio ao debugar. Mais comum em APIs internas com controle sobre todos os consumidores.

Independente da escolha:

- **Só crie uma nova versão para mudanças que quebram compatibilidade** (remover/renomear campo, mudar tipo, mudar semântica de um status code). Adicionar um campo novo opcional na resposta, ou um endpoint novo, não exige nova versão.
- **Anuncie depreciação com antecedência.** Use um header (`Deprecation: true`, `Sunset: <data>`) na versão antiga assim que a nova estiver disponível, e documente a data-limite de suporte.
- **Mantenha a versão antiga funcionando durante o período de transição** — depreciar não é o mesmo que desligar imediatamente. Rotas de versões diferentes podem compartilhar a mesma camada de domínio, com adapters finos por versão para o formato de request/response.

## Erros como parte do contrato

Padronize o formato de erro em toda a API (mesmo schema para 400, 404, 409, 500), para que o consumidor escreva um único parser de erro:

```json
{
  "error": {
    "code": "user_not_found",
    "message": "user not found"
  }
}
```

Documente cada `code` possível por endpoint no OpenAPI — isso é o que permite ao consumidor tratar erros programaticamente em vez de fazer parsing de string de mensagem.
