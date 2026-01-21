# Flujo de Exportación CSV en Footprint

## Resumen

El comando `fp export` extrae eventos pendientes de la base de datos SQLite, los enriquece con metadata de Git, y los guarda en un archivo CSV plano con rotación por año. Los CSVs se almacenan en un repositorio Git local que puede sincronizarse con un remoto.

---

## Flujo Completo

```
fp export [flags]
    │
    ├─→ Verifica intervalo de exportación (config: export_interval)
    │
    ├─→ Query: SELECT * FROM repo_events WHERE status_id = 0 (PENDING)
    │
    ├─→ Ordena eventos por timestamp (oldest first)
    │
    ├─→ Para cada evento:
    │   ├─ Determina año del evento
    │   ├─ Selecciona archivo destino:
    │   │   ├─ Año actual → commits.csv
    │   │   └─ Año anterior → commits-{año}.csv
    │   ├─ Enriquece con Git metadata (author, stats, parents)
    │   └─ Append de fila al CSV correspondiente
    │
    ├─→ Git commit de archivos modificados
    │
    ├─→ UPDATE repo_events SET status_id = 1 WHERE id IN (exportados)
    │
    ├─→ Guarda timestamp en config (export_last)
    │
    └─→ Git push a origin (si existe remote configurado)
```

---

## Estructura de Archivos

```
~/.config/Footprint/exports/     (o ~/Library/Application Support/Footprint/exports/)
├── commits.csv          ← Año actual (activo)
├── commits-2024.csv     ← Eventos del 2024
├── commits-2023.csv     ← Eventos del 2023
└── ...
```

**Rotación por año:**
- Eventos del año actual van a `commits.csv`
- Eventos de años anteriores van a `commits-{año}.csv`
- Todos los repositorios se mezclan en el mismo archivo (columna `repo_id` para filtrar)
- Al iniciar nuevo año, los eventos nuevos van a `commits.csv` (ahora vacío para el nuevo año)

---

## Columnas del CSV Exportado

| Columna | Fuente | Descripción |
|---------|--------|-------------|
| `authored_at` | Git | Fecha del autor (RFC3339) |
| `repo` | BD | Identificador del repositorio |
| `branch` | BD | Rama donde ocurrió |
| `commit` | BD | Hash completo (40 chars) |
| `subject` | Git | Primera línea del mensaje |
| `author` | Git | Nombre del autor |
| `author_email` | Git | Email del autor |
| `files` | Git | Archivos modificados |
| `additions` | Git | Líneas agregadas |
| `deletions` | Git | Líneas eliminadas |
| `parents` | Git | Hashes de commits padre (separados por espacio) |
| `committer` | Git | Nombre del committer |
| `committer_email` | Git | Email del committer |
| `committed_at` | BD | Timestamp RFC3339 del evento |
| `source` | BD | Origen (post-commit, backfill, etc.) |
| `machine` | Sistema | Hostname de la máquina donde se registró |

---

## Flags Disponibles

| Flag | Descripción |
|------|-------------|
| `--force` | Ignora el intervalo de exportación |
| `--dry-run` | Muestra qué se exportaría sin ejecutar |
| `--open` | Abre el directorio de exports en file manager |

---

## Estados de Eventos

| Estado | ID | Descripción |
|--------|-----|-------------|
| PENDING | 0 | Listo para exportar |
| EXPORTED | 1 | Ya fue exportado |
| SKIPPED | 2 | Saltado (no usado) |
| FAILED | 3 | Falló (no usado) |

Solo se exportan eventos con `status_id = 0`.

---

## Configuración Relacionada

Configurar vía `fp config set <key> <value>`:

| Key | Descripción |
|-----|-------------|
| `export_remote` | URL del repositorio remoto para sync |
| `export_interval` | Segundos entre exports automáticos (default: 3600) |
| `export_repo` | Path al directorio de exportación |
| `export_last` | Unix timestamp del último export (interno) |

Ejemplo:
```bash
fp config set export_remote git@github.com:user/my-exports.git
fp config set export_interval 1800
```

---

## Auto-Export

La función `MaybeExport()` se llama después de cada `fp record`. Si ha pasado el intervalo configurado, ejecuta el export automáticamente en background.

---

# Historial de cambios

## Mejoras implementadas

Las siguientes mejoras fueron implementadas tras la evaluación inicial:

1. **Eliminado `commit_short`** - Redundante, se puede derivar del commit completo
2. **Eliminado `is_merge`** - Redundante, se puede derivar de `parents` (si tiene espacios = merge)
3. **Agregado `authored_at`** - Fecha del autor (puede diferir de commit date en rebases/cherry-picks)
4. **Agregado `machine`** - Hostname de la máquina donde se registró el commit
5. **Renombrado columnas** para mayor claridad:
   - `timestamp` → `committed_at`
   - `repo_id` → `repo`
   - `message` → `subject`
   - `files_changed` → `files`
   - `insertions` → `additions`
   - `parent_commits` → `parents`
   - `author_name` → `author`
   - `committer_name` → `committer`
6. **Reordenado columnas** por importancia (datos más relevantes primero)
7. **Parents separados por espacio** en lugar de coma (evita conflictos con CSV)

## Mejoras pendientes (baja prioridad)

| Mejora | Descripción |
|--------|-------------|
| `event_id` | Referencia para debugging o correlación con BD local |
| `--retry-failed` | Re-exportar eventos que fallaron por metadata incompleta |
| `--include-body` | Exportar mensaje completo opcionalmente |
| `files_list` | Lista de archivos modificados (archivo separado) |
| `gpg_signature` | Indicar si el commit estaba firmado |
