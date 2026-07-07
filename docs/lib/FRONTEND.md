# Frontend: React, Vite e Playwright

Dependencias registradas:

- React + Vite, ultimas estaveis planejadas
- Playwright, ultima estavel planejada

## Papel no Morfeu

O frontend sera uma SPA para cliente e backoffice. A decisao registrada descarta
Next.js por SSR desnecessario e por evitar mais um servidor Node em producao.

## React

Regras esperadas:

- componentes pequenos, com responsabilidade clara;
- estado remoto separado de estado local de UI;
- formularios com validacao na borda;
- acessibilidade por roles, labels e foco;
- seletores E2E estaveis quando texto puder mudar por i18n.

Para mapa de assentos, modelar estado de forma explicita:

- assento livre;
- selecionado localmente;
- indisponivel;
- reservado;
- bloqueado por regra.

Nao confiar no estado do frontend para garantir reserva. A API e o banco devem
validar concorrencia.

## Vite

Uso esperado:

- build estatico servido por Caddy;
- proxy apenas em desenvolvimento;
- variaveis `VITE_*` somente para valores publicos;
- nenhuma secret no bundle.

Exemplo de separacao:

```text
VITE_API_BASE_URL=/api
```

Secrets de Stripe, TMDB, JWT e SMTP ficam no backend.

## Playwright

Playwright cobre 2 a 4 jornadas criticas conforme ADR 0006:

- login;
- escolher sessao e assento;
- iniciar pagamento em modo teste;
- fluxo administrativo essencial, se existir no MVP.

Padroes:

- Page Object Model quando uma tela for reutilizada;
- seletores por role sempre que possivel;
- `data-testid` para partes dinamicas ou textos sujeitos a i18n;
- testes independentes, com dados isolados por UUID;
- screenshots/videos apenas para falhas no CI.

Exemplo:

```ts
await page.getByRole('button', { name: 'Entrar' }).click()
await expect(page.getByRole('heading', { name: 'Em cartaz' })).toBeVisible()
```

## Contrato com API

Gerar ou documentar contratos antes de acoplar UI a payloads instaveis. Erros
devem ter shape previsivel, por exemplo:

```json
{
  "error": {
    "code": "seat_unavailable",
    "message": "Assento indisponivel",
    "request_id": "..."
  }
}
```

O frontend nao deve depender de mensagens internas para fluxo de controle; usar
`error.code`.

