# Flujo de Aplicación Footprint (`fp`)

## Qué es Footprint

Footprint es una herramienta CLI que registra automáticamente la actividad de Git en repositorios locales. Usa hooks de Git para capturar eventos (commits, merges, pushes) y los almacena en una base de datos SQLite local.

---

## Flujo General

```
Usuario ejecuta comando → Parsing → Dispatch → Acción → Respuesta/BD
```

---

## 1. Punto de Entrada

**Archivo:** `cmd/fp/main.go`

Cuando ejecutas `fp [comando] [flags]`:

1. Se extraen las flags y comandos de los argumentos
2. Se normaliza la entrada (ej: `-5` → `--limit=5`)
3. Se construye el árbol de comandos disponibles
4. Se despacha al handler correspondiente

---

## 2. Árbol de Comandos

**Archivo:** `internal/cli/tree.go`

Define la jerarquía de comandos como nodos padre-hijo:

```
fp (raíz)
├── track / untrack / repos    → Gestión de repositorios
├── setup / teardown / check   → Instalación de hooks
├── activity / watch           → Visualización de eventos
├── config (get|set|unset|list)→ Configuración
├── record                     → Plumbing (llamado por hooks)
└── export / backfill          → Import/Export de datos
```

Cada nodo tiene: nombre, descripción, flags válidas y función de acción.

---

## 3. Dispatcher

**Archivo:** `internal/dispatchers/dispatch.go`

Traduce los tokens del usuario en una resolución ejecutable:

1. Recorre el árbol buscando coincidencias con los tokens
2. Valida que las flags sean reconocidas
3. Valida que haya suficientes argumentos
4. Retorna un objeto `Resolution` con la acción a ejecutar

---

## 4. Acciones

**Directorio:** `internal/actions/`

Implementan la lógica de cada comando. Usan inyección de dependencias para ser testeables:

```go
// Punto de entrada público
func Track(args, flags) error {
    return track(args, flags, DefaultDeps())
}

// Función con dependencias inyectables
func track(args, flags, deps) error {
    // Lógica real
}
```

---

## 5. Sistema de Hooks

**Directorio:** `internal/hooks/`

`fp setup` instala 5 hooks globales de Git:

- `post-commit` - después de cada commit
- `post-merge` - después de merge/pull
- `post-checkout` - cambio de rama
- `post-rewrite` - después de rebase/amend
- `pre-push` - antes de push

Cada hook es un script simple:

```bash
#!/bin/sh
FP_SOURCE=post-commit /path/to/fp record >/dev/null
```

---

## 6. Flujo de Registro de Evento

Cuando haces `git commit` en un repo trackeado:

```
git commit
    ↓
post-commit hook ejecuta: fp record
    ↓
record():
  - Lee FP_SOURCE del ambiente
  - Verifica que el repo esté trackeado
  - Obtiene commit hash, rama actual
  - Inserta evento en SQLite
    ↓
Evento guardado ✓
```

---

## 7. Almacenamiento

**Base de datos:** `$XDG_CONFIG_HOME/Footprint/store.db` (SQLite)

Tabla principal `repo_events`:

| Campo       | Descripción                            |
| ----------- | -------------------------------------- |
| repo_id     | Identificador único (ej: `github.com/user/repo`) |
| commit_hash | SHA del commit                         |
| branch      | Rama donde ocurrió                     |
| timestamp   | Cuándo ocurrió                         |
| status      | pending/exported/orphaned              |
| source      | post-commit/post-merge/etc             |

**Configuración:** `~/.fprc` (formato key=value)

```
trackedRepos=github.com/user/repo1,github.com/user/repo2
theme=default-dark
```

---

## 8. Identificación de Repositorios

**Archivo:** `internal/repo/repo.go`

Deriva IDs únicos de las URLs de remoto:

```
git@github.com:user/repo.git  → github.com/user/repo
https://github.com/user/repo  → github.com/user/repo
(sin remoto)                  → local:/full/path/to/repo
```

---

## 9. Ejemplo Completo: `fp activity -5`

```
1. Parsing
   - flags: ["--limit=5"]
   - commands: ["activity"]

2. Dispatch
   - Encuentra nodo "activity" en el árbol
   - Valida flag --limit
   - Retorna Resolution con acción Activity

3. Ejecución
   - Abre BD SQLite
   - Query: SELECT ... LIMIT 5 ORDER BY timestamp DESC
   - Formatea resultados

4. Output
   - Muestra eventos en pantalla (via pager)
```

---

## Capas de la Arquitectura

```
┌─────────────────────────────────────┐
│      ENTRADA (main.go, args.go)     │  ← Parsing de CLI
├─────────────────────────────────────┤
│      ROUTING (tree.go, dispatch.go) │  ← Resolución de comandos
├─────────────────────────────────────┤
│      ACCIONES (actions/*)           │  ← Lógica de negocio
├─────────────────────────────────────┤
│      SERVICIOS                      │
│  git/ repo/ config/ hooks/ store/   │  ← Interacción con sistema
├─────────────────────────────────────┤
│      DATOS                          │
│  SQLite (store.db) + Config (.fprc) │  ← Persistencia
└─────────────────────────────────────┘
```

---

## Puntos Clave de Diseño

- **Sin daemons:** Cada invocación de `fp` es independiente
- **Hooks transparentes:** Scripts mínimos que delegan a `fp record`
- **Testeable:** Inyección de dependencias en todas las acciones
- **Multiplataforma:** Rutas configuradas según SO (macOS/Linux/Windows)
- **Migraciones automáticas:** La BD se actualiza al abrir
