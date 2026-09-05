# Active Memory - GoWatch Session

**Data**: 2026-08-23  
**Status**: P5.1 entregue - Sistema de Alertas & Limites de Recursos (Alerting Engine) funcionando

---

## O que foi entregue

### P5.1 - Sistema de Alertas & Limites (Alerting Engine)
- `internal/alert/types.go` implementou os modelos de dados `Alert`, `AlertRule`, `Severity` e `AlertStatus`, além de regras padrão (`HIGH_CPU`, `HIGH_MEM`, `OOM_KILLED`, `CONTAINER_DOWN`).
- `internal/alert/engine.go` implementou `AlertEngine` com histerese (resolução apenas abaixo do limite inferior), contadores de amostras consecutivas (*hitCounters*), cálculo seguro de memória (trata `MemLimit == 0`), normalização de CPU por core e auto-resolução de alertas de containers efêmeros.
- `internal/alert/engine_test.go` adicionou cobertura de testes unitários testando disparos consecutivos, histerese, divisão por zero, containers efêmeros e concorrência (`go test -race`).
- `internal/docker/collector.go` integrou o `globalAlertEngine` ao ciclo de amostragem `WatchContainers`.
- `internal/ui/dashboard.go` adicionou resumo compacto de alertas na barra `System Resources`, atalho `w` / `W` para abrir/fechar o modal de Alertas (`alertsModal`) e atalho `c` / `C` para limpar alertas resolvidos.

### P4.0 (Anterior - AWS Integration)
- Módulos AWS CloudWatch, Lambda, XRay, CloudFormation e alternância na TUI (`a`/`A`).

---

## Validações executadas
- `make build` OK (binário compilado sem erros em `bin/gowatch`)
- `make test` OK (100% dos testes passando em todos os pacotes: `internal/alert`, `internal/docker`, `internal/ui`, `internal/aws`, `internal/filter`, `internal/trace`, `pkg/metrics`)
- `make go-sec` / `go vet` OK (0 erros de verificação)

---

## Estado atual
- Fase 5.1 entregue com sucesso.
- Próximas entregas: Canais externos de notificação (P5.2 - Webhooks, Slack, SNS) e Ações de Gerenciamento de Containers (Fase 7).

---

## Notas técnicas importantes
- O atalho `w` alterna exibição do modal `alertsModal`.
- O atalho `c` limpa alertas com status `RESOLVED`.
- REGRA ESTREITA: É estritamente proibido citar co-participação de IA, ferramentas de IA ou referências a IA na construção de código, mensagens de commit, pull requests ou documentação.
- REGRA ESTREITA: NUNCA criar commits ou executar git push sem a ordem expressa e individual do usuário.

---

**Encerrado em**: 2026-08-23  
**Retomar por**: Implementação das fases P5.2 ou Fase 7 conforme priorizado pelo usuário.
