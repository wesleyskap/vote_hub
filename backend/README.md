# Main API (backend administrativo)

Este diretório contém a API core do projeto, focada nas regras de negócios globais e entrega de dados para o painel de gestão. Desenvolvida sob o poderoso framework MVC do ecossistema Ruby.

## Tecnologias

- **Ruby** 3.3
- **Rails** 8.0 (API mode)
- **PostgreSQL** 16 (Interface relacional via ActiveRecord)

## Mapeamento de rotas (Main API)

Desenvolvida em Ruby on Rails 8, fornece endpoints altamente cacheados para o front-end e para o painel da produção.

- `GET /api/v1/participants`
  - **Função:** Retorna a listagem dos participantes "emparedados" e suas fotos para montar a tela de votação (Frontend).
  - **Mecanismo crítico:** Dados estáticos para a semana do paredão, ideais para serem servidos a partir de cache Redis para não onerar o banco.

- `GET /api/v1/results/current`
  - **Função:** Retorna o consolidado (total agrupado) de votos daquele paredão, agrupado por participante.

- `GET /admin/v1/stats`
  - **Função:** Rota de consumo do Dashboard da produção, exibindo QPS calculado e status global da votação ao vivo.

## Decisões arquiteturais

1. **Ruby on Rails para a regra de negócio:** O framework entrega extrema velocidade de desenvolvimento (Convention over Configuration). O ActiveRecord abstrai a modelagem do banco de dados e agiliza a exposição de rotas RESTful para os dashboards administrativos.
2. **Arquitetura inspirada em CQRS:** A Ingestion API isola o fluxo de escrita (Commands) para suportar o tráfego pesado. A API Rails foca exclusivamente na leitura (Queries) e administração, garantindo consistência e isolando o consumo interno de instabilidades externas causadas por eventuais ataques na borda pública.
3. **Gerenciamento centralizado do banco de dados (Database owner):** A API Rails atua como a dona exclusiva do schema do Postgres. Ela gerencia o `db/schema.rb`, as migrations e a integridade estrutural das tabelas. O ecossistema em Go apenas consome e insere dados nessas tabelas já consolidadas.

## Estrutura de pastas

```text
backend/
├── app/
│   ├── controllers/  # Controladores API que respondem ao front-end ou dashboards.
│   ├── models/       # ActiveRecord: Regras de negócio, relacionamentos (Participant, Vote).
│   └── views/        # (Se aplicável) Views da administração ou serializadores JSON.
├── config/
│   ├── database.yml  # Configuração de conexão ao PostgreSQL.
│   └── routes.rb     # Declaração das rotas RESTful do negócio.
├── db/               # Controle absoluto do Schema do banco de dados (migrations e seeds).
└── Dockerfile        # Imagem Ruby enxuta (Alpine/Debian slim) pronta para K8s.
```

## Componentes

- **Ruby on Rails 8 (Ruby 3.3)**: Expõe os serviços de leitura e orquestração administrativa.
- **Dona do schema**: É a API Rails quem determina a composição das tabelas, mantendo as migrações do PostgreSQL (`db/migrate`) e cuidando do esboço principal da base de dados.
- **Relatórios e dashboards**: Não é responsabilidade do Rails gravar os picos de votos diretos, quem faz a gravação pesada em lote é o Ingestion Worker. A aplicação Rails compartilha o acesso à mesma base PostgreSQL, otimizada primariamente para leituras rápidas que compõem a telemetria em tempo real.

## Mapeamento de rotas (Main API)

- `GET /api/v1/participants`
  - **Função:** Retorna a listagem dos participantes "emparedados" e suas fotos para montar a tela de votação (Frontend).
  - **Mecanismo crítico:** Dados estáticos para a semana do paredão, ideais para serem servidos a partir de cache Redis para não onerar o banco.

- `GET /api/v1/results/current`
  - **Função:** Retorna o consolidado (total agrupado) de votos daquele paredão, agrupado por participante.

- `GET /admin/v1/stats`
  - **Função:** Rota de consumo do Dashboard da produção, exibindo QPS calculado e status global da votação ao vivo.

## Fluxo
1. Produtores acessam o Dashboard via rotas do Rails.
2. A aplicação Rails consolida queries nos agregados de bancos e cache, respondendo de forma ágil ao dashboard oficial.

## Comandos úteis

```bash
# Executar console interativo dentro do container
docker compose exec main-api bundle exec rails c

# Rodar migrations pendentes
docker compose exec main-api bundle exec rails db:migrate

# Popular o banco de dados com os seeds padrão (Paredão e Participantes)
docker compose exec main-api bundle exec rails db:seed

# Listar todas as rotas ativas da aplicação
docker compose exec main-api bundle exec rails routes
```

## Troubleshooting (Resolução de problemas)

- **Erro "A server is already running" (server.pid):**
  Se o container da Main API cair abruptamente, o arquivo `tmp/pids/server.pid` pode persistir e travar novas inicializações.
  *Resolução:* Execute `rm tmp/pids/server.pid` na pasta `backend/` ou destrua o container com `docker compose down` para remover os temporários montados.

- **Conexão com PostgreSQL recusada:**
  Certifique-se de que a variável de ambiente `DATABASE_URL` aponta corretamente para o container do banco. Em ambiente local externo aos containers, configure o host como `localhost`. Dentro do Docker Compose, a string de conexão deve usar o nome do host `postgres` definido nos serviços.
