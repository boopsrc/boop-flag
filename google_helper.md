# Guia: configurar o Login com Google (OAuth2)

Este guia leva você, do zero, a habilitar o botão **"Continuar com Google"** do
Boop. Ele é bem detalhado de propósito — não é preciso ter experiência com o
Google Cloud. Ao final, um usuário real conseguirá entrar com a conta Google.

> Enquanto você não configura nada, o app continua 100% utilizável em **modo dev**
> (`AUTH_DEV_MODE=true`), que usa um login de mentira. Este guia é só para ligar o
> login **real** do Google.

---

## O que vamos fazer e por quê

O app usa o fluxo **OAuth2 Authorization Code**: quando o usuário clica em
"Continuar com Google", o backend o envia para a tela de consentimento do Google;
depois de autorizar, o Google devolve um código para o backend, que troca esse
código pelos dados do perfil (e-mail, nome, foto) e cria a sessão. Para isso
funcionar, o Google precisa **conhecer o seu app** — é o que criamos aqui: um
**Client ID** e um **Client Secret**, mais o endereço de retorno autorizado.

---

## Pré-requisitos

- Uma conta Google (a mesma que você vai usar para testar serve).
- O projeto rodando localmente:
  - Backend em `http://localhost:8080`
  - Frontend em `http://localhost:5173`
- Acesso ao arquivo `backend/.env` (copie de `backend/.env.example` se ainda não
  existir).

Valores que este projeto espera (já batem com o código, não precisa mudar nada
no app):

| Item | Valor |
|---|---|
| Redirect URI (endereço de retorno) | `http://localhost:8080/api/v1/auth/google/callback` |
| Origem do frontend | `http://localhost:5173` |
| Origem do backend | `http://localhost:8080` |
| Scopes (permissões pedidas) | `openid`, `email`, `profile` |

---

## Passo 1 — Criar (ou escolher) um projeto no Google Cloud

1. Acesse o **Google Cloud Console**: <https://console.cloud.google.com/>.
2. No topo da página, clique no seletor de projeto (ao lado do logo "Google
   Cloud") e depois em **"New Project" / "Novo projeto"**.
3. Dê um nome (ex.: `boop-flag`) e clique em **Create / Criar**.
4. Aguarde alguns segundos e **selecione o projeto recém-criado** no mesmo
   seletor do topo. Tudo dos próximos passos deve acontecer dentro dele.

---

## Passo 2 — Configurar a tela de consentimento OAuth

Antes de criar as credenciais, o Google exige configurar a tela que o usuário vê
ao autorizar o app.

1. No menu lateral, vá em **APIs & Services → OAuth consent screen**
   (**APIs e serviços → Tela de permissão OAuth**).
   Link direto: <https://console.cloud.google.com/apis/credentials/consent>.
2. Em **User Type**, escolha **External / Externo** e clique em **Create / Criar**.
   (É a opção correta para contas Google comuns, fora de uma organização Workspace.)
3. Preencha a aba **"App information"**:
   - **App name**: `Boop` (ou o nome que quiser mostrar ao usuário).
   - **User support email**: seu e-mail.
   - **Developer contact information** (no fim da página): seu e-mail.
   - Os demais campos (logo, domínios) são opcionais para desenvolvimento — pode
     deixar em branco.
   - Clique em **Save and Continue / Salvar e continuar**.
4. Na aba **"Scopes"**:
   - Clique em **Add or Remove Scopes** e marque os três básicos:
     `.../auth/userinfo.email`, `.../auth/userinfo.profile` e `openid`.
   - Clique em **Update**, depois **Save and Continue**.
   > Esses são exatamente os scopes que o backend pede (`openid`, `email`,
   > `profile`). Não é preciso adicionar mais nada.
5. Na aba **"Test users"**:
   - Clique em **Add users** e adicione **o(s) e-mail(s) Google que você vai usar
     para testar**. Enquanto o app estiver em modo "Testing", só essas contas
     conseguem entrar.
   - Clique em **Save and Continue**.
6. Revise o resumo e finalize. O app fica no estado **"Testing"** — perfeito para
   desenvolvimento.

> **Nota (Google Auth Platform novo):** dependendo da conta, a interface pode
> aparecer reorganizada em "Branding", "Audience" e "Clients". A ideia é a mesma:
> preencha as informações do app em *Branding*, mantenha o público em *Testing* e
> adicione seus e-mails de teste em *Audience → Test users*.

---

## Passo 3 — Criar o OAuth Client ID

1. No menu lateral, vá em **APIs & Services → Credentials**
   (**APIs e serviços → Credenciais**).
   Link direto: <https://console.cloud.google.com/apis/credentials>.
2. Clique em **+ Create Credentials → OAuth client ID**.
3. Em **Application type**, escolha **Web application**.
4. Dê um nome (ex.: `boop-flag-web`).
5. Em **Authorized JavaScript origins**, clique em **+ Add URI** e adicione as
   duas origens (uma de cada vez):
   ```
   http://localhost:5173
   http://localhost:8080
   ```
6. Em **Authorized redirect URIs**, clique em **+ Add URI** e adicione
   **exatamente** este endereço:
   ```
   http://localhost:8080/api/v1/auth/google/callback
   ```
   ⚠️ Precisa ser idêntico ao do backend: mesmo esquema (`http`), mesma porta
   (`8080`), mesmo caminho, **sem barra no final**. Qualquer diferença causa o
   erro `redirect_uri_mismatch`.
7. Clique em **Create**. Uma janela mostrará o **Client ID** e o **Client
   Secret**. Copie os dois (dá para revê-los depois na mesma tela de Credentials).

---

## Passo 4 — Preencher o `.env` do backend

Abra `backend/.env` e defina:

```bash
GOOGLE_CLIENT_ID=cole-aqui-o-client-id
GOOGLE_CLIENT_SECRET=cole-aqui-o-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback

# Desligue o modo dev para usar o login real do Google:
AUTH_DEV_MODE=false
```

Confira também que `JWT_SECRET` está preenchido (qualquer valor longo e
aleatório; gere um com `openssl rand -hex 32`) e que `FRONTEND_URL` aponta para
`http://localhost:5173`.

> **Nunca** faça commit do `.env` com as credenciais reais — ele já está no
> `.gitignore`. Compartilhe apenas o `.env.example`, sem valores.

---

## Passo 5 — Reiniciar e testar

1. Reinicie o backend para carregar as novas variáveis:
   ```bash
   cd backend
   go run ./cmd/api
   ```
   Se alguma credencial obrigatória faltar, o backend avisa e não sobe (isso é
   proposital — falha rápido). Com `AUTH_DEV_MODE=false`, `GOOGLE_CLIENT_ID` e
   `GOOGLE_CLIENT_SECRET` passam a ser obrigatórios.
2. Garanta que o frontend está rodando:
   ```bash
   cd frontend
   npm run dev
   ```
3. Abra <http://localhost:5173>, clique em **"Continuar com Google"**.
4. Escolha a conta (precisa ser uma das cadastradas em **Test users**), aceite as
   permissões e você será levado ao **onboarding** (primeiro acesso) ou direto à
   **home**.

Pronto — o login real com Google está funcionando. 🎉

---

## Solução de problemas

| Erro / sintoma | Causa provável | Como corrigir |
|---|---|---|
| `Error 400: redirect_uri_mismatch` | O redirect URI registrado no Google não é idêntico ao usado pelo backend. | Confira em **Credentials → seu client → Authorized redirect URIs** se está **exatamente** `http://localhost:8080/api/v1/auth/google/callback` (sem barra final, porta `8080`). Precisa bater com `GOOGLE_REDIRECT_URL` no `.env`. |
| `Acesso bloqueado: o app não concluiu o processo de verificação` / `access_blocked` | A conta usada não está na lista de **Test users** (app em "Testing"). | Adicione o e-mail em **OAuth consent screen → Test users**. Ou publique o app (veja "Indo para produção"). |
| `Error 401: invalid_client` | Client ID e/ou Secret errados ou trocados. | Reveja os valores em **Credentials** e cole novamente no `.env`. Reinicie o backend. |
| Resposta `403 email_unverified` do backend | A conta Google usada não tem o e-mail verificado. | Use uma conta Google com e-mail verificado (o backend valida `email_verified` por segurança). |
| Clicou em Google e voltou para a landing sem logar | `AUTH_DEV_MODE` ainda está `true`, ou o backend não foi reiniciado após editar o `.env`. | Defina `AUTH_DEV_MODE=false` e reinicie `go run ./cmd/api`. |
| Backend não sobe: "GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET are required" | `AUTH_DEV_MODE=false` sem as credenciais preenchidas. | Preencha as duas variáveis no `.env`, ou volte para `AUTH_DEV_MODE=true` enquanto não tiver as credenciais. |
| `This app isn't verified` (aviso, não erro) | Normal em desenvolvimento. | Clique em **Advanced → Go to Boop (unsafe)** para prosseguir como test user. |

---

## Indo para produção

Quando publicar o app num domínio real (ex.: `https://boop.seudominio.com`):

1. **Adicione os URIs de produção** no mesmo OAuth client (ou crie um client
   separado para produção):
   - Authorized JavaScript origins: `https://boop.seudominio.com` e o domínio do
     backend.
   - Authorized redirect URIs: `https://SEU-BACKEND/api/v1/auth/google/callback`.
2. **Atualize o `.env` de produção**: `GOOGLE_REDIRECT_URL` e `FRONTEND_URL` com
   os domínios `https://`. O backend liga automaticamente a flag `Secure` do
   cookie de sessão quando `FRONTEND_URL` começa com `https://`.
3. **Publique a tela de consentimento**: em **OAuth consent screen**, mude o
   status de "Testing" para **In production** (**Publish app**). Assim qualquer
   usuário — não só os test users — pode entrar. Dependendo dos scopes, o Google
   pode pedir verificação; para `openid/email/profile` (não sensíveis), o
   processo costuma ser simples.
4. **Sempre HTTPS** em produção — cookies de sessão e OAuth não devem trafegar em
   HTTP puro.

---

## Referências oficiais

- Console de credenciais: <https://console.cloud.google.com/apis/credentials>
- Tela de consentimento OAuth: <https://console.cloud.google.com/apis/credentials/consent>
- Documentação OAuth 2.0 do Google: <https://developers.google.com/identity/protocols/oauth2/web-server>
