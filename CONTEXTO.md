# 📖 Contexto e Memória de Engenharia do Sistema FocalboardSX

> **🚨 AVISO MANDATÓRIO PARA O ANTIGRAVITY E AGENTES DE IA:**
> 1. **SEMPRE leia este arquivo com atenção** no início de cada nova sessão ou ao importar o projeto em outro computador para compreender o histórico, regras de negócio e decisões de arquitetura.
> 2. **SEMPRE ATUALIZE este arquivo** após concluir qualquer nova funcionalidade, refatoração ou correção de bug. Adicione os detalhes da nova versão na seção de *Histórico de Versões* e mantenha a documentação viva.
> 3. **Mantenha a sincronização de versão**: O número de versão deve estar consistente em `webapp/package.json`, `webapp/src/constants.ts`, `server/model/version.go`, `README.md` e neste arquivo.
> 4. **Assinatura de Commits**: Use sempre a identidade configurada: `eduardoazm <eduardo.azm@hotmail.com>`.

---

## 📌 1. Visão Geral do Sistema

* **Nome do Projeto**: FocalboardSX (Sistema Integrado de Gestão de Quadros, Tarefas e Projetos)
* **Repositório Git**: `https://github.com/eduardoazm/FocalboardSX.git`
* **Imagem Docker Pré-compilada**: `ghcr.io/eduardoazm/focalboardsx:latest`
* **Versão Atual**: `v1.0.4`
* **Tecnologias Principais**: Go 1.21 (Backend HTTP/WebSocket REST), React 17, TypeScript 4, Redux Toolkit, SCSS, SQLite3 (nativo em `/opt/focalboard/data`) / PostgreSQL opcional, Docker (Multi-stage build com suporte multi-arquitetura amd64 e arm64), GitHub Actions CI/CD (GHCR), Portainer.

### Propósito & Escopo de Negócio
O **FocalboardSX** é uma solução corporativa self-hosted para gestão visual de tarefas, quadros kanban, tabelas, galerias e calendários (alternativa ao Trello, Notion e Asana). 

Diferenciais e melhorias implementadas:
* **Localização Total em Português (Brasil - PT-BR)** como idioma padrão de fábrica do sistema, com gravação e persistência definitiva da preferência de idioma no banco de dados.
* **Usuário Administrador Nativo**: Criação automática do administrador padrão (`admin` / `admin123`) na inicialização do servidor.
* **Controle Estrito de Acesso (RBAC)**: Middleware de segurança no backend (`systemAdminRequired`) bloqueando qualquer acesso não autorizado com HTTP 403 Forbidden.
* **Painel de Gestão de Usuários Exclusivo (`/admin/users`)**: Tela administrativa com CRUD completo para gerenciar todos os usuários cadastrados (criar, editar, redefinir senhas e excluir contas).
* **Compatibilidade Multi-Arquitetura (amd64 e arm64)**: Imagem Docker compilada para servidores x86/amd64 e processadores ARM64 (Oracle Cloud Ampere A1, Raspberry Pi, AWS Graviton, Apple Silicon).
* **Compilação Automatizada no GitHub Actions**: Build em nuvem e publicação no GitHub Container Registry (`ghcr.io`), permitindo **deploy instantâneo no Portainer** em poucos segundos via `docker-compose.yml`, sem sobrecarregar a CPU ou memória do servidor.

---

## 🕒 2. Histórico Completo de Versões e Evolução do Projeto

### `v1.0.4` — Compartilhamento Obrigatório de Quadros, UI Limpa e Suporte Multi-Arquitetura (ARM64 / AMD64)
* **Compartilhamento Obrigatório de Quadro na Criação de Usuários**:
  * Adicionado campo obrigatório de seleção de quadro (*Quadro Inicial Compartilhado \**) no modal de cadastro de novo usuário.
  * O backend Go vincula automaticamente o novo usuário como membro ativo do quadro selecionado (`AddMemberToBoard`) e insere o quadro na categoria padrão da barra lateral, impedindo que novos usuários fiquem presos na tela vazia *"Nenhum quadro ainda"*.
  * Nova rota administrativa `GET /api/v2/admin/user_boards` para listar quadros disponíveis.
* **Estilização Limpa da Tabela de Usuários**:
  * Removidas as nomenclaturas e etiquetas visuais `Root` e `Você`.
  * Avatar com cores neutras e suaves integradas ao tema, sem gradientes fortes.
  * Coluna Função (`role-cell`) simplificada para texto puro legível (`Administrador` / `Membro`), removendo pills, molduras e cores chamativas.
* **Suporte Multi-Arquitetura Docker (ARM64 e AMD64)**:
  * Adicionado `docker/setup-qemu-action@v3` no fluxo do GitHub Actions (`docker-build.yml`).
  * Definido `platforms: linux/amd64,linux/arm64` no `build-push-action`, gerando manifesto multi-arch nativo compatível com VPS Oracle Cloud (processadores ARM64 Ampere) e servidores convencionais AMD64/x86.
  * Otimizado o estágio `nodebuild` no `Dockerfile` com `--platform=$BUILDPLATFORM` para compilação nativa ultra-rápida do frontend webapp.
* **Sincronização Rigorosa do Número de Versão**:
  * Atualizados `webapp/src/constants.ts` (`versionString = '1.0.4'`), `server/model/version.go` (`CurrentVersion = "1.0.4"`), `webapp/package.json` (`1.0.4`) e tag git `v1.0.4`.

### `v1.0.3` — Correção de Contraste no Campo Assignee e Localização Dinâmica de Propriedades e Placeholders para PT-BR
* **Correção de Contraste do Campo Assignee (PersonSelector) no Tema Escuro**:
  * Adicionado `color: 'rgb(var(--center-channel-color-rgb))'` no estilo `singleValue`, `input` e `placeholder` de `webapp/src/components/personSelector.tsx`.
  * Adicionada cor inline adaptativa no componente de renderização de opções (`formatOptionLabel`).
  * Estilizados `.react-select__single-value`, `.Person-item`, `.MultiPerson-item` e `.react-select__control` em `webapp/src/components/personSelector.scss` com `color: rgb(var(--center-channel-color-rgb)) !important`, eliminando o texto escuro sobre fundo escuro.
* **Localização Dinâmica de Nomes de Propriedades em Quadros Existentes e Novos**:
  * Criação do helper `Utils.localizePropertyName(name, intl)` em `webapp/src/utils.ts`.
  * Tradução automática e dinâmica dos rótulos de propriedades de quadros existentes (`Priority` -> *Prioridade*, `Assignee` -> *Responsável*, `Estimated Hours` -> *Horas Estimadas*, `Due Date` -> *Data de Entrega*, `Created By` -> *Criado por*, `Date Created` -> *Data de Criação*, `Last Updated By` -> *Última atualização por*, `Date Updated` -> *Data de atualização*, etc.).
  * Aplicado em: detalhes do cartão (`cardDetailProperties.tsx`), cabeçalhos da tabela (`tableHeaders.tsx`), tooltips de cartões kanban (`kanbanCard.tsx`), cabeçalhos de coluna (`kanbanColumnHeader.tsx`), menus de agrupamento (`viewHeaderGroupByMenu.tsx`), menus de ordenação (`viewHeaderSortMenu.tsx`), menus de propriedades (`viewHeaderPropertiesMenu.tsx`), menus de exibição por data (`viewHeaderDisplayByMenu.tsx`) e filtros (`filterEntry.tsx`).
* **Tradução Completa de Placeholders Hardcoded em Toda a Interface**:
  * Título do cartão vazio em `cardDetail.tsx` e `tableRow.tsx`: de `"Untitled"` para `intl.formatMessage` (*"Sem nome"*).
  * Descrição do cartão em `cardDetailContents.tsx`: de `"Add a description..."` para *"Adicionar uma descrição..."*.
  * Descrição do quadro em `viewTitle.tsx`: de `"Add a description..."` para *"Adicionar uma descrição..."*.
  * Título de visualização em `viewHeader.tsx`: de `"Untitled View"` para *"Visualização sem título"*.
  * Cabeçalho de novo select em `kanbanColumnHeader.tsx` e `tableGroupHeaderRow.tsx`: de `"New Select"` para *"Nova opção"*.
  * Opções de ordenação em `viewHeaderSortMenu.tsx`: *"Nome"*, *"Manual"*, *"Reverter"*.
  * Campo de busca em `searchDialog.tsx`: de `"Search for boards"` para *"Pesquisar quadros"*.
  * Opções de cálculos de tabela e kanban em `options.tsx` (*Contar vazios*, *Contar preenchidos*, *Percentual vazio*, *Percentual preenchido*).
  * Placeholder do editor de blocos em `rootInput.tsx`: *"Insira seu texto ou comando"*.
  * Adição de todas as novas chaves em `webapp/i18n/pt_BR.json` e `webapp/i18n/en.json`.

### `v1.0.2` — Localização Total de Modelos (Templates), Quadros e Barra Lateral em PT-BR
* **Tradução Completa dos 13 Modelos Nativos (`templates-boardarchive`)**:
  * Tradução dos títulos, descrições, visualizações, propriedades, colunas de status e cartões de todos os 13 modelos embutidos (`Tarefas do Projeto`, `Bem-vindo ao Focalboard!`, `Pauta de Reunião`, `Pipeline de Vendas (CRM)`, `Tarefas Pessoais`, `Metas e OKRs da Empresa`, `Metas Pessoais`, `Planejamento de Sprint`, `Sessões de Pesquisa com Usuários`, `Análise Concorrencial`, `Calendário de Conteúdo`, `Retrospectiva da Equipe` e `Roadmap de Produto`).
  * Tradução das colunas de Status (*Não iniciado*, *Em andamento*, *Bloqueado*, *Concluído 🙌*, *Arquivado*, etc.) e prioridades (*Alta 🔥*, *Média*, *Baixa*).
  * Recompilação do pacote embutido `server/assets/templates.boardarchive` e incremento de `defaultTemplateVersion` para `7` em `server/app/templates.go`, forçando a reimportação automática dos novos modelos em português no boot.
* **Barra Lateral e Categorias (Sidebar)**:
  * Localização do cabeçalho da categoria padrão de `BOARDS` para `QUADROS` em `webapp/src/store/sidebar.ts` e `webapp/src/components/sidebar/sidebarCategory.tsx`.
* **Compatibilidade e Internacionalização**:
  * Suporte a nomes em português no tour inicial (`Bem-vindo ao Focalboard!` e `Criar um novo cartão`) no backend (`server/app/onboarding.go`) e frontend (`cardDetail.tsx`, `tourTipRenderer.tsx`, `add_properties.tsx`, `viewHeader.tsx`).
  * Inclusão de novas chaves de tradução em `webapp/i18n/pt_BR.json`.
* **Build Docker**:
  * Inclusão de `make templates-archive` no Dockerfile para garantir que o pacote de modelos seja sempre reempacotado durante o build.

### `v1.0.1` — Tema Adaptativo e Melhorias Visuais na Tela de Usuários
* **Adequação ao Sistema de Temas do Focalboard**:
  * Substituição de cores fixas/hardcoded em `webapp/src/pages/adminUsersPage.scss` por variáveis CSS dinâmicas nativas do Focalboard (`--center-channel-bg-rgb`, `--center-channel-color-rgb`, `--button-bg-rgb`, etc.).
  * Suporte automático e harmônico a temas claro, escuro e personalizados.
* **Correção de Contraste e Legibilidade**:
  * Ajuste de textos de descrição, labels, placeholders e dicas no formulário para manter legibilidade total em fundos escuros e claros.
  * Inputs e selects agora utilizam transparências dinâmicas sobre `--center-channel-color-rgb` e fundo adaptativo no foco.
* **Ajuste Dimensional dos Modais Administrativos**:
  * Aplicação da classe customizada `.AdminDialog` aos 4 modais administrativos (Criar Usuário, Editar Usuário, Redefinir Senha e Excluir Usuário).
  * Limitação de tamanho para proporções compactas (`max-width: 520px; height: auto`), eliminando espaços pretos/vazios excessivos e garantindo uma interface limpa e centralizada.
* **Sincronização de Versão**:
  * Atualização coordenada em `webapp/package.json`, `webapp/src/constants.ts`, `server/model/version.go`, `README.md` e `CONTEXTO.md`.

### `v1.0.0` — Fundação, Localização PT-BR, Painel Admin CRUD & Pipeline Portainer
* **Tradução Completa e Persistência de Idioma**:
  * Configuração do idioma padrão inicial para `'pt-br'` em `webapp/src/store/language.ts` e `webapp/src/i18n.tsx`.
  * Persistência da preferência de idioma diretamente no banco de dados via `client.patchUserConfig` na troca de linguagem.
  * Sincronização automática na inicialização e login através das preferências do usuário (`myConfig.language.value`).
  * Tradução dos nomes de cores, etiquetas e mais de 100 chaves de localização em `webapp/i18n/pt_BR.json`.
  * Internacionalização completa dos formulários de Login, Cadastro e Redefinição de Senha.
* **Módulo de Administrador do Sistema (`system_admin`)**:
  * Criação automática do usuário administrador no startup do servidor (`admin` / `admin123`) via `app.EnsureAdminUser()`.
  * Primeiro usuário registrado ou usuário com login `admin` recebe automaticamente o papel de `system_admin`.
  * Validação de permissões em `server/services/permissions/localpermissions/localpermissions.go` verificando privilégio `manage_system`.
  * Atribuição da role `system_admin` no retorno de `/api/v2/users/me` para controle de visibilidade no frontend.
* **Middleware de Segurança e Endpoints Administrativos**:
  * Criação do middleware `systemAdminRequired` em `server/api/auth.go`.
  * Endpoints REST restritos:
    * `GET /api/v2/admin/users`: Listagem de todos os usuários com dados completos.
    * `POST /api/v2/admin/users`: Cadastro de novos usuários (username, email, senha inicial e papel).
    * `PUT /api/v2/admin/users/{userID}`: Edição de dados cadastrais e função.
    * `DELETE /api/v2/admin/users/{userID}`: Exclusão com invalidação atômica de sessões.
    * `POST /api/v2/admin/users/{userID}/password`: Redefinição de senha diretamente pelo administrador.
* **Tela de Gerenciamento de Usuários (`/admin/users`)**:
  * Componente `AdminUsersPage.tsx` e estilos `adminUsersPage.scss`.
  * Bloqueio client-side para não-administradores com tela de acesso restrito.
  * Tabela com busca em tempo real por nome e e-mail, badges (*"Você"*, *"Root"*, *"Administrador"*, *"Membro"*), data formatada e contadores.
  * Modais nativos para Criação, Edição, Redefinição de Senha e Confirmação de Exclusão.
  * Link direto "Gerenciar Usuários" incorporado no menu do perfil (`SidebarUserMenu`) apenas para administradores.
* **Pipeline CI/CD e Otimização para Portainer**:
  * GitHub Actions (`.github/workflows/docker-build.yml`) que compila o Webapp e o Go Server na nuvem e publica a imagem Docker no `ghcr.io/eduardoazm/focalboardsx:latest`.
  * `docker-compose.yml` na raiz configurado para baixar a imagem pré-compilada, eliminando a lentidão de compilação no servidor do Portainer (deploy instantâneo em segundos).

---

## 🏛️ 3. Arquitetura Técnica Detalhada

### 3.1. Estrutura de Pastas Principais
```text
├── .github/
│   └── workflows/
│       └── docker-build.yml     # Pipeline de build e push da imagem Docker para o GHCR
├── docker/
│   ├── Dockerfile               # Dockerfile multi-stage
│   ├── server_config.json       # Configurações do servidor (telemetria desativada, porta 8000)
│   └── docker-compose.yml       # Exemplo de compose para docker local
├── server/
│   ├── api/
│   │   ├── auth.go              # Autenticação, sessões e middleware systemAdminRequired
│   │   └── users.go             # Endpoints /users e /admin/users (CRUD de usuários)
│   ├── app/
│   │   ├── auth.go              # Lógica de registro com auto-atribuição de admin
│   │   └── user.go              # Operações de CRUD de usuários e EnsureAdminUser()
│   ├── model/
│   │   ├── user.go              # Modelo User com permissões e sanitização
│   │   └── version.go           # Lista de versões do backend
│   └── services/
│       ├── permissions/         # Verificação de permissões locais e system_admin
│       └── store/sqlstore/      # Implementação SQLite3 / Postgres
├── webapp/
│   ├── i18n/
│   │   └── pt_BR.json           # Dicionário de traduções em Português Brasil
│   ├── src/
│   │   ├── components/
│   │   │   └── sidebar/sidebarUserMenu.tsx # Menu do usuário com link "Gerenciar Usuários"
│   │   ├── pages/
│   │   │   ├── adminUsersPage.tsx          # Tela administrativa completa de usuários
│   │   │   ├── adminUsersPage.scss         # Estilos da tela administrativa
│   │   │   ├── loginPage.tsx
│   │   │   └── registerPage.tsx
│   │   ├── octoClient.ts        # Cliente HTTP com métodos adminGetUsers, adminCreateUser, etc.
│   │   └── router.tsx           # Roteamento React com rota /admin/users
│   └── package.json             # Manifest do frontend (v1.0.0)
├── Dockerfile                   # Dockerfile raiz multi-stage otimizado
├── docker-compose.yml           # Compose pronto para deploy instantâneo no Portainer
├── CONTEXTO.md                  # Este documento de memória de engenharia
└── README.md                    # Documentação do projeto para usuários e deploy
```

---

## ⚠️ 4. Regras de Projeto e Armadilhas Conhecidas (Gotchas)

1. **Proteção do Usuário Root Admin**:
   * O usuário `admin` não pode ser excluído por nenhum administrador.
   * O nome de usuário do `admin` não pode ser renomeado e sua função não pode ser rebaixada para membro.
2. **Auto-Exclusão Bloqueada**:
   * O administrador conectado não pode excluir a sua própria conta.
3. **Exclusão de Sessões Ativas**:
   * Ao excluir um usuário, todas as linhas correspondentes na tabela `sessions` são deletadas imediatamente para derrubar acessos em andamento.
4. **Deploy no Portainer sem Compilação Local**:
   * *Regra de Ouro*: Nunca use blocos `build: .` na stack de produção do Portainer. O Portainer deve apenas puxar a imagem pronta `ghcr.io/eduardoazm/focalboardsx:latest`.

---

## 🚀 5. Procedimentos de Deploy e Operação

### Credenciais Padrão do Administrador
* **Usuário**: `admin`
* **Senha**: `admin123`

### Deploy Inicial no Portainer (Produção)
1. Acesse o painel do **Portainer**.
2. Vá em **Stacks** > **Add stack**.
3. Nomeie a stack (ex: `focalboardsx`).
4. Selecione a opção **Web editor** e cole o conteúdo do `docker-compose.yml`:
```yaml
version: "3.8"

services:
  focalboardsx:
    image: ghcr.io/eduardoazm/focalboardsx:latest
    container_name: focalboardsx
    restart: unless-stopped
    ports:
      - "8000:8000"
    environment:
      - VIRTUAL_HOST=localhost
      - VIRTUAL_PORT=8000
    volumes:
      - focalboard_data:/opt/focalboard/data

volumes:
  focalboard_data:
    driver: local
```
5. Clique em **Deploy the stack**. O download da imagem pronta leva apenas alguns segundos!
6. Acesse no navegador: `http://SEU_IP_OU_DOMINIO:8000`.

### Atualização no Portainer
1. Acesse a Stack `focalboardsx` no **Portainer**.
2. Clique na aba **Editor** da stack.
3. Certifique-se de marcar a opção **"Re-pull image and redeploy"**.
4. Clique em **Update the stack**.
5. Em segundos, o Portainer baixa a nova imagem gerada pelo GitHub Actions e reinicia o serviço.

### Configuração Git Obrigatória
Para qualquer commit no repositório:
```bash
git config user.name "eduardoazm"
git config user.email "eduardo.azm@hotmail.com"
```

---

*Documento atualizado em: 13 de Outubro de 2026 — Versão v1.0.3*
