---
name: serena-mcp-tools
description: >
  Serena MCP for semantic navigation, symbol-based editing, and low-token-cost
  refactors. Trigger: using Serena MCP, searching declarations, references,
  implementations, diagnostics, or editing code by symbol.
metadata:
  author: a.sierra
  version: "1.2"
  scope: [root]
  auto_invoke: "Using Serena MCP, semantic code navigation, symbol editing, or refactors"
allowed-tools: Read, Edit, Write, Glob, Grep, Bash, Task, Serena MCP, Engram, Context7, delegate
---

# Serena MCP Tools

## Mission

Usa Serena como capa semantica de IDE: declaraciones, simbolos, referencias, implementaciones, diagnosticos y refactors reales del lenguaje.

No lo uses como reemplazo universal de `grep`, shell, git, lectura directa o parches pequenos. Esa es la trampa: una herramienta poderosa mal usada quema tokens, tiempo y confianza.

## Decision Gate

| Pregunta | Si | No |
| --- | --- | --- |
| Necesito saber que significa un identificador en el lenguaje? | Serena | `grep`/lectura directa |
| Necesito referencias reales, overrides o implementaciones? | Serena | `grep` si es texto literal |
| El cambio depende de una clase, metodo, funcion, interfaz o tipo? | Serena | parche directo |
| Es Markdown, JSON, YAML, TOML, `.env`, CI, Dockerfile o docs? | herramienta nativa | Serena solo si contiene codigo con simbolos |
| Ya conozco el archivo y son 1-5 lineas locales? | parche directo | Serena si hay impacto semantico |
| Es shell, test, build, lint, format o git? | shell nativo | no Serena |

Regla Gentle AI: usa Serena cuando reduce incertidumbre semantica. Si solo reduce pereza, NO.

## Token Harness

1. Define el objetivo semantico: simbolo, archivo probable, contenedor, firma o modulo.
2. Mapea antes de editar: overview, symbol, declaration o references.
3. Lee el minimo: cuerpo del simbolo antes que archivo completo.
4. Edita por simbolo si el punto de insercion o renombre es semantico.
5. Verifica con diff, diagnosticos y tests relevantes.
6. Si Serena devuelve ruido, ambiguedad o cambios amplios, vuelve a `grep` + patch controlado.

## Tool Routing

| Necesidad | Herramienta preferida |
| --- | --- |
| Mapa top-level de archivo desconocido | `serena_get_symbols_overview` |
| Buscar clase, funcion, metodo, tipo o interfaz | `serena_find_symbol` |
| Saltar de uso a definicion | `serena_find_declaration` |
| Encontrar implementaciones de abstracciones | `serena_find_implementations` |
| Medir impacto real de un cambio | `serena_find_referencing_symbols` |
| Insertar cerca de una entidad conocida | `serena_insert_before_symbol` / `serena_insert_after_symbol` |
| Reemplazar una funcion/clase localizada | `serena_replace_symbol_body` |
| Renombrar entidad real del lenguaje | `serena_rename_symbol` |
| Revisar errores de lenguaje | `serena_get_diagnostics_for_file` |
| Texto literal, logs, strings, rutas | `grep` / `glob` |
| Tests, build, lint, format, git | `bash` nativo |
| Docs o config | lectura/patch nativo |

## Operating Harness

### Bootstrap

- Llama `serena_initial_instructions` una vez por sesion si vas a usar Serena.
- Activa el proyecto con `serena_activate_project` cuando no este claro.
- Ejecuta `serena_check_onboarding_performed` antes de depender de conocimiento del repo.
- Si JetBrains es backend, confirma que la raiz abierta en IDE coincide con la raiz activa de Serena.

### Explore

- Empieza por `get_symbols_overview` en archivos candidatos.
- Usa `find_symbol` con `max_matches` bajo cuando el nombre sea ambiguo.
- Pide `include_body: false` hasta que necesites implementacion.
- Usa `find_referencing_symbols` antes de tocar APIs publicas, utilidades compartidas o interfaces.

### Edit

- Prefiere `rename_symbol` para renombres reales, no busqueda textual.
- Prefiere `insert_before_symbol` o `insert_after_symbol` cuando el punto de insercion sea estable por identidad.
- Usa `replace_symbol_body` solo despues de localizar exactamente el simbolo correcto.
- No mezcles rename, move y cambio de comportamiento en una sola operacion.

### Verify

- Revisa `git diff --stat` y `git diff -- <archivos>`.
- Ejecuta diagnosticos Serena sobre archivos editados si el backend es fiable.
- Ejecuta tests minimos relevantes del stack.
- Si no puedes ejecutar tests, dilo explicitamente y limita la afirmacion a diff/diagnosticos/analisis.

## Task Playbooks

### Entender una funcion o clase

1. `find_symbol` para localizarla.
2. Lee solo firma/cuerpo necesario.
3. `find_referencing_symbols` si importa el impacto.
4. Explica invariantes, dependencias y riesgos.

### Cambiar una API publica

1. `find_symbol` o `find_declaration` del simbolo principal.
2. `find_referencing_symbols` para call sites.
3. `rename_symbol` solo si es renombre puro.
4. Edita definicion y usos en unidades pequenas.
5. Corre diagnosticos y tests especificos.

### Agregar funcionalidad

1. `get_symbols_overview` en archivos candidatos.
2. Localiza patrones existentes con `find_symbol` y `grep` complementario.
3. Inserta por simbolo cuando el punto sea semantico.
4. Usa herramientas nativas para tests, snapshots, docs y config.

### Refactor grande

1. Mapa de simbolos y referencias antes de editar.
2. Divide por unidad semantica: clase, metodo, modulo o interfaz.
3. Aplica una operacion semantica por vez.
4. Verifica diff y diagnosticos entre pasos.
5. Detente si el backend no esta indexado o los resultados no son confiables.

## Stop Conditions

Deten Serena y cambia de estrategia si:

- El simbolo no resuelve o resuelve otra entidad.
- El resultado trae demasiado contexto irrelevante.
- El backend/LSP no esta indexado o sus diagnosticos son dudosos.
- La edicion toca archivos no relacionados.
- La tarea es texto literal, configuracion, documentacion o git.
- El cliente MCP no descubre herramientas de Serena consistentemente.

## Safety Rules

- No ejecutes shell via Serena si el agente ya tiene shell auditable.
- No actives herramientas beta/opcionales salvo necesidad concreta.
- Trata hooks, scripts, generated files y dependencias como superficie de ejecucion.
- Mantene transportes HTTP de Serena en localhost salvo justificacion explicita.
- Si el cliente no pide confirmacion para ejecutar herramientas, evita operaciones con escritura o shell.

## Verification Commands

```bash
git diff --stat
git diff -- <archivos_modificados>
```

```bash
# Ajustar al stack real.
npm test -- --runInBand
pytest -q
go test ./...
cargo test
mvn test
```

## Agent Prompts

### Navegacion

> Usa Serena para localizar declaracion y referencias reales de `<symbol>`. No leas archivos completos salvo que el cuerpo del simbolo sea insuficiente. Devuelve impacto por archivo y simbolo contenedor.

### Refactor

> Usa Serena para renombrar `<old_symbol>` a `<new_symbol>` como simbolo semantico, no como busqueda textual. Despues revisa diff, diagnosticos y tests relevantes. Si hay ambiguedad, detente antes de editar.

### Auditoria de impacto

> Usa Serena para encontrar implementaciones y referencias de `<symbol>`. Clasifica usos por lectura, escritura, override, llamada directa, llamada indirecta o test. No modifiques codigo.

### Insercion segura

> Usa Serena para insertar el metodo/helper junto al simbolo relacionado mas cercano. No cambies imports ni call sites hasta verificar overview y referencias.

## Source Notes

- Serena aporta recuperacion, edicion, refactorizacion y diagnostico semantico a nivel de simbolos via MCP.
- Sus herramientas pueden estar parcialmente habilitadas segun contexto/modo; no asumas disponibilidad universal.
- Por defecto usa language servers; con JetBrains el backend se decide al iniciar y la raiz del IDE debe coincidir con la raiz Serena.
- Los contextos `claude-code`, `codex` e `ide` estan pensados para no duplicar capacidades nativas de clientes/agentes.
- Documentacion primaria: `https://oraios.github.io/serena/`.

## Maintenance

Revisa esta skill cuando cambien herramientas MCP, backend Serena, cliente MCP o topologia del proyecto. Mantene estable la decision central: Serena para semantica; herramientas nativas para texto, shell, git, config y parches pequenos.
