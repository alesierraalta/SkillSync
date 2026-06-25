## Exploration: fix-cursor-and-metadata

### Current State
1. **Bug de Cursor Duplicado/Erróneo en `InstallerModel`**: En `internal/ui/installer_model.go`, la función `OptionsView()` renderiza los elementos de navegación. En la línea 266, el indicador de cursor para el botón "[ Execute Install ]" (`cursorAction`) verifica si `m.Cursor == 9+storageOffset` en lugar de `10+storageOffset`. Esto causa que tanto la opción global "Add shell aliases to profile" como el botón de instalación muestren el cursor de selección al mismo tiempo cuando `m.Cursor` vale `9+storageOffset`. Además, cuando `m.Cursor` vale `10+storageOffset`, el botón de instalación no muestra ningún cursor.
2. **Pérdida de metadatos en `Format()`**: En `internal/parser/parser.go`, la función `Format()` serializa un objeto `types.Skill` de vuelta a Markdown con frontmatter. Si `yamlStr == ""` (cuando el archivo original no tiene frontmatter o es nuevo), la función genera un frontmatter mínimo hardcodeado que solo incluye `scope` y `auto_invoke`. Campos críticos de metadatos como `name`, `description` y `local_only` se omiten por completo, provocando pérdida de información al guardar.

### Affected Areas
- `internal/ui/installer_model.go` — Corregir la condición de selección del cursor para `cursorAction` de `9+storageOffset` a `10+storageOffset` en la línea 266.
- `internal/parser/parser.go` — Modificar la rama `yamlStr == ""` de `Format()` para serializar correctamente todos los metadatos disponibles del skill (`Name`, `Description`, `Scope`, `AutoInvoke`, `LocalOnly`).

### Approaches
1. **Enfoque 1: Corrección de cursor directa y serialización de frontmatter estructurada**
   - **InstallerModel**: Cambiar la condición a `m.Cursor == 10+storageOffset` en `internal/ui/installer_model.go:266`.
   - **Parser Format**: Cuando `yamlStr == ""`, en lugar de construir un string manualmente de forma incompleta, podemos utilizar un mapa o estructura YAML temporal y codificarla usando `yaml.Marshal` para asegurar que `name`, `description`, `scope`, `auto_invoke` y `local_only` se serialicen de forma consistente y limpia.
   - Pros: Limpio, robusto, no propenso a errores de formato de texto manuales, conserva todos los campos de metadatos.
   - Cons: Ninguno evidente.
   - Effort: Low

### Recommendation
Se recomienda el Enfoque 1. Es la forma más limpia y robusta de corregir ambos problemas sin introducir lógica de strings frágil en el parser.

### Risks
- **Riesgo 1**: Alterar la indentación o el formato del frontmatter generado. Se mitigará verificando con los tests unitarios existentes en `internal/parser/parser_test.go` y añadiendo casos de test específicos para el guardado sin frontmatter previo.

### Ready for Proposal
Yes — Los problemas y sus soluciones están completamente claros y localizados.
