# Boop 🤍

Sistema de login e cadastro com **Login com Google (OAuth2)**, backend em **Go**,
frontend em **React (Vite + TypeScript + Tailwind + Framer Motion)** e um mascote
branco sorridente na home — com direito a dar um _boop_ no nariz dele.

Com OAuth2, login e cadastro são a mesma porta: o primeiro acesso de uma conta
Google cria o usuário automaticamente e leva a uma tela de **onboarding** (escolher
o nome de exibição) antes de cair na home.

## Estrutura

```
boop-flag/
├── backend/     # API em Go (chi, pgx, oauth2, JWT em cookie HttpOnly)
├── frontend/    # SPA em React + Vite + TS + Tailwind + Framer Motion
└── docker-compose.yml   # Postgres 16 para desenvolvimento
```

O backend segue as boas práticas documentadas em
`.claude/skills/golang-api-best-practices/` (cmd/internal, handlers finos,
injeção de dependência por interface, erros com wrap, `context.Context`, logging
estruturado, graceful shutdown, config por variáveis de ambiente).

## Pré-requisitos

- Go 1.24+
- Node 20+ / npm
- Docker (para o Postgres) — ou um Postgres 16 local

## Configuração

### 1. Banco de dados

```bash
docker compose up -d          # sobe o Postgres em localhost:5432
```

O schema é criado automaticamente pelo backend no boot (idempotente).

### 2. Backend

```bash
cd backend
cp .env.example .env          # ajuste os valores conforme necessário
go run ./cmd/api
```

Gere um `JWT_SECRET` forte, por exemplo:

```bash
openssl rand -hex 32
```

### 3. Frontend

```bash
cd frontend
npm install
npm run dev                   # http://localhost:5173
```

O dev server do Vite faz proxy de `/api` para o backend (`localhost:8080`),
então o cookie de sessão é tratado como same-origin em desenvolvimento.

## Testar sem credenciais do Google (modo dev)

Com `AUTH_DEV_MODE=true` no `.env` (padrão do `.env.example`), a landing mostra o
link **"Entrar em modo dev (sem Google)"**, que cria/loga um usuário fake. Assim
você exercita todo o fluxo (onboarding → home → boop → logout) sem precisar de
client ID/secret. **Nunca deixe `AUTH_DEV_MODE=true` em produção.**

## Configurar o Login com Google (fluxo real)

1. Acesse **Google Cloud Console → APIs & Services → Credentials**.
2. Crie um **OAuth client ID** do tipo **Web application**.
3. Em **Authorized redirect URIs**, adicione:
   `http://localhost:8080/api/v1/auth/google/callback`
4. Copie o **Client ID** e o **Client Secret** para o `.env`
   (`GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`) e defina `AUTH_DEV_MODE=false`.
5. Reinicie o backend. O botão **"Continuar com Google"** passa a funcionar.

## Variáveis de ambiente (backend)

| Variável | Descrição | Obrigatória |
|---|---|---|
| `PORT` | Porta do servidor HTTP (padrão 8080) | não |
| `DATABASE_URL` | Conexão Postgres | sim |
| `FRONTEND_URL` | Origem do frontend (CORS + redirects) | não (padrão :5173) |
| `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` | Credenciais OAuth | sim, se `AUTH_DEV_MODE=false` |
| `GOOGLE_REDIRECT_URL` | Redirect URI registrado no Google | não (tem padrão) |
| `JWT_SECRET` | Segredo para assinar sessão e state OAuth | sim |
| `SESSION_TTL` | Duração da sessão (ex.: `24h`) | não |
| `AUTH_DEV_MODE` | Habilita login mock | não (padrão false) |

## Endpoints da API (`/api/v1`)

| Método | Rota | Descrição |
|---|---|---|
| GET | `/healthz` | Health check |
| GET | `/auth/google/login` | Inicia o fluxo OAuth (redirect) |
| GET | `/auth/google/callback` | Callback do Google |
| GET | `/auth/dev/login` | Login mock (só em `AUTH_DEV_MODE`) |
| POST | `/auth/logout` | Encerra a sessão |
| GET | `/me` | Usuário autenticado |
| PATCH | `/me` | Conclui o onboarding (nome de exibição) |

## Testes

```bash
cd backend && go test ./...      # JWT, state, service (fakeRepo), handler /me
cd frontend && npm run build     # type-check + build de produção
```

## Segurança

- Sessão em cookie **HttpOnly**, **SameSite=Lax**, **Secure** em produção (HTTPS).
- Parâmetro `state` do OAuth assinado (proteção CSRF).
- CORS restrito à origem do frontend, com credenciais.
- Validação de `email_verified` do Google.
- Nenhum segredo versionado — apenas `.env.example`.
