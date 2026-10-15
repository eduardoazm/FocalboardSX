# 🚀 FocalboardSX

[![Docker Build & Publish](https://github.com/eduardoazm/FocalboardSX/actions/workflows/docker-build.yml/badge.svg)](https://github.com/eduardoazm/FocalboardSX/actions/workflows/docker-build.yml)
[![Version](https://img.shields.io/badge/version-v1.0.4-blue.svg)](https://github.com/eduardoazm/FocalboardSX/releases)
[![Language](https://img.shields.io/badge/language-pt--BR-green.svg)](https://github.com/eduardoazm/FocalboardSX)
[![Docker Image](https://img.shields.io/badge/docker-ghcr.io-blue?logo=docker)](https://github.com/eduardoazm/FocalboardSX/pkgs/container/focalboardsx)

> **FocalboardSX** é uma versão personalizada e aprimorada do Focalboard para produtividade e gestão de projetos, totalmente em Português do Brasil (PT-BR), com controle de usuários administradores, tela de gerenciamento de usuários (CRUD) e compilação automatizada no GitHub Actions para **deploy instantâneo no Portainer**.

---

## ✨ Principais Funcionalidades e Modificações

* 🇧🇷 **100% em Português (Brasil)**:
  * Idioma padrão nativo definido como `pt-BR`.
  * Persistência definitiva da escolha de idioma gravada no banco de dados.
  * Todas as interfaces (Login, Cadastro, Menus, Modais, Cores e Configurações) traduzidas.

* 🛡️ **Painel e Usuário Administrador (`system_admin`)**:
  * Administrador padrão criado automaticamente no primeiro boot:
    * **Usuário**: `admin`
    * **Senha**: `admin123`
  * Middleware de segurança no backend (`systemAdminRequired`) com bloqueio de acesso a não-administradores (`403 Forbidden`).
  * Tela exclusiva de **Gerenciamento de Usuários** acessível em `/admin/users` e pelo menu do perfil.

* 👥 **CRUD Completo de Usuários**:
  * **Criação**: Cadastro de novos membros e administradores com validações.
  * **Listagem**: Tabela com busca em tempo real por nome/e-mail, badges (*Você*, *Root*, *Administrador*, *Membro*) e data de registro.
  * **Edição**: Alteração de nome, e-mail e função do usuário.
  * **Redefinição de Senha**: Troca de senha direta pelo administrador com confirmação.
  * **Exclusão**: Exclusão segura com derrubada imediata de sessões ativas (protegendo o usuário conectado e o `admin` raiz).

* ⚡ **Deploy Instantâneo no Portainer**:
  * O GitHub Actions compila o código e publica a imagem Docker no GitHub Container Registry (`ghcr.io/eduardoazm/focalboardsx:latest`).
  * O servidor com Portainer não precisa compilar nada: apenas baixa a imagem pronta em segundos!

---

## 🚀 Como Subir no Portainer (Passo a Passo)

### 1. Criar a Stack no Portainer
1. Abra o seu painel do **Portainer**.
2. Navegue até **Stacks** > **+ Add stack**.
3. Dê um nome para a stack, por exemplo: `focalboardsx`.
4. No campo **Web editor**, cole a configuração abaixo:

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

5. Clique no botão **Deploy the stack**.
6. Em poucos segundos, o container estará rodando em `http://SEU-IP:8000`!

### 2. Acesso Inicial
* Acesse `http://SEU-IP:8000`
* Entre com o login do administrador:
  * **Usuário**: `admin`
  * **Senha**: `admin123`
* *(Recomendado)*: Acesse o menu do usuário no canto superior esquerdo e altere a senha do `admin` ou gerencie novos usuários.

---

## 🔄 Como Atualizar a Aplicação no Portainer

Sempre que um novo commit for enviado para o repositório, o GitHub Actions gera uma nova imagem automaticamente. Para atualizar no seu servidor:

1. Acesse o **Portainer** > **Stacks** > `focalboardsx`.
2. Clique na aba **Editor**.
3. Marque a caixa de seleção **"Re-pull image and redeploy"**.
4. Clique em **Update the stack**.
5. O Portainer atualizará a versão instantaneamente sem perda de dados (os dados ficam seguros no volume `focalboard_data`).

---

## 🛠️ Executando Localmente com Docker Compose

Se preferir rodar localmente no seu computador:

```bash
# Iniciar o container em segundo plano
docker compose up -d

# Ver os logs
docker compose logs -f

# Parar o container
docker compose down
```

---

## 📖 Documentação de Engenharia

Para entender o histórico detalhado, decisões de projeto e diretrizes técnicas, consulte o arquivo [CONTEXTO.md](CONTEXTO.md).

---

## 👤 Autor

* **Eduardo** ([@eduardoazm](https://github.com/eduardoazm)) - `eduardo.azm@hotmail.com`
