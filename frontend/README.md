# Interface web pública (frontend)

Este diretório contém a aplicação de interface web baseada em React (SPA), com o objetivo de entregar a tela de votação principal para os espectadores do programa.

## Tecnologias

- **React** 20
- **Vite** (bundler de desenvolvimento e build de alta performance)
- **CSS Vanilla** (CSS nativo com uso de custom properties para tema e performance)
- **Google reCAPTCHA v3** (validação invisível de autenticidade humana)
- **Canvas Fingerprinting** (gerador de assinatura de hardware para detecção de bots)

## Padrão crítico do código (Hooks sem estado)

Uma armadilha comum em projetos React é atrelar funções de Fetch e manipulação de estado (`useState`) diretamente no corpo da UI visual (ex: no HTML/JSX do arquivo). Isso "chumba" as regras de negócios às tags `<div>` e `<button>`.

No diretório `src/hooks/useVote.js`, toda a lógica crítica de disparo para a API, invocação do canvas fingerprint e verificação invisível do reCAPTCHA estão perfeitamente contidas numa função agnóstica de View. O componente visual apenas invoca `submitVote()`.

- **O impacto:** O reaproveitamento da mesma engine de votação para outros canais (Smart TVs, Mobile Apps com React Native) se torna imediato. A complexidade fica testável de forma unitária, completamente independente da árvore do DOM.

## Decisões arquiteturais

1. **React.js:** Framework ideal para Single Page Applications (SPA), viabilizando componentes reativos e atualizações assíncronas fluidas sem recarregamento da página durante o processo de votação.
2. **Custom hooks agnósticos (Stateless):** A lógica de negócio, a comunicação com a Ingestion API e a injeção do reCAPTCHA estão totalmente extraídas para hooks personalizados (ex: `useVote`). O isolamento de estado e regras HTML facilita uma eventual migração do projeto para React Native com reaproveitamento de código quase absoluto.
3. **Vanilla CSS puro:** A ausência de frameworks robustos (como Bootstrap ou TailwindCSS) resulta em um bundle final extremamente leve. O carregamento de estilos reduzidos favorece um tempo menor de First Contentful Paint (FCP), característica crítica para a retenção do público em transmissões em massa na TV.

## Estrutura de pastas

```text
frontend/
├── src/
│   ├── assets/       # Imagens e ícones otimizados (WebP/SVG).
│   ├── components/   # Componentes visuais burros (Dumb/Presentational), ex: Botoes, Avatares.
│   ├── hooks/        # Lógica de negócio e comunicação com a Ingestion API (Custom Hooks).
│   ├── styles/       # Variáveis CSS globais, tokens de cores e temas.
│   └── App.jsx       # Componente raiz, injetor de providers e layout principal.
├── public/           # Arquivos estáticos (favicon, manifest).
├── index.html        # Ponto de montagem da raiz do DOM do React.
└── package.json      # Dependências (React, Vite, bibliotecas de canvas/fingerprint).
```

## Stack e ferramentas

- **React e Hooks**: Componentes 100% funcionais (Functional Components). A lógica de negócios e estado são extraídas do escopo da tela principal e encapsuladas em hooks customizados (custom hooks). Esse design permite o reaproveitamento imediato dessa mesma lógica caso uma versão nativa (Mobile) utilizando React Native venha a ser desenvolvida, minimizando o esforço de re-escrita.
- **Vite**: Usado como bundler moderno, proporcionando build nativo de módulos ultra-veloz no ambiente de desenvolvimento.
- **CSS Vanilla**: Garantia de máxima flexibilidade visual aliada a cores vibrantes para aderência de design moderno.

## Regras de segurança

Para proteger a integridade do voto e mitigar ataques por bots (script kiddies), o frontend encapsula validações em duas frentes antes de assinar a submissão para a API:
1. **Google reCAPTCHA v3**: Implementado na etapa de resolução de quebra-cabeça oculto, atestando comportamento humano no clique.
2. **Device Fingerprinting**: Análise via canvas API que coleta um identificador matemático das características nativas do hardware do usuário, limitando enxurradas de votos por máquinas fraudulentas sem bloquear sessões IP inteiras (comum em redes corporativas ou familiares).

## Comandos úteis

```bash
# Instalar dependências locais
npm install

# Iniciar servidor de desenvolvimento local
npm run dev

# Gerar o bundle de produção otimizado na pasta dist/
npm run build

# Executar a suíte de testes com Vitest
npm run test
```

## Troubleshooting (Resolução de problemas)

- **Erro de CORS (Cross-Origin Resource Sharing):**
  Se a requisição falhar indicando bloqueio de política CORS, certifique-se de que a Ingestion API em Go está configurada para aceitar requisições da URL do frontend. No arquivo `router.go` da API, verifique as configurações de `cors.New` e se a porta do frontend está na lista de origens aceitas.

- **Falha no carregamento do reCAPTCHA (sitekey inválida):**
  A aplicação do frontend faz uso do script oficial de reCAPTCHA do Google. Caso a validação retorne erros, certifique-se de que as chaves públicas configuradas no arquivo de variáveis de ambiente combinam com as chaves privadas carregadas na Ingestion API de Go. Em desenvolvimento local, chaves públicas de teste padrão do Google podem ser utilizadas.
