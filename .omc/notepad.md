# Notepad
<!-- Auto-managed by OMC. Manual edits preserved in MANUAL section. -->

## Priority Context
<!-- ALWAYS loaded. Keep under 500 chars. Critical discoveries only. -->

## Working Memory
<!-- Session notes. Auto-pruned after 7 days. -->

## MANUAL
<!-- User content. Never auto-pruned. -->
### 2026-04-22 07:02
## Server Architecture Exploration Report

### Directory Structure Findings

**Key Directories:**
- `/Users/cnhyk/Downloads/V0/server/internal/api/` - Route handlers (10 files)
- `/Users/cnhyk/Downloads/V0/server/internal/worker/` - Broadcaster (1 file)
- `/Users/cnhyk/Downloads/V0/server/internal/export/` - Export manager (1 file)
- `/Users/cnhyk/Downloads/V0/server/internal/service/` - Service layer (8 files)
- `/Users/cnhyk/Downloads/V0/server/internal/domain/` - Type definitions
- `/Users/cnhyk/Downloads/V0/docker-compose.yml` - Docker orchestration

### 1. API Handlers - Request Decoding & Validation

**Location:** `/Users/cnhyk/Downloads/V0/server/internal/api/`

**Key Findings:**
- **decodeJSON function** (line 145-148 in server.go):
  - Simple implementation: `json.NewDecoder(r.Body).Decode(target)`
  - NO `validate.Struct()` calls anywhere in the codebase
  - All JSON decoding uses this single pattern without validation

**Validation Pattern:**
- Struct tags in domain types DO include `validate` tags (e.g., `validate:"required"`)
- Examples from domain/types.go:
  - `ProblemWriteInput.Latex: validate:"required"`
  - `ProblemWriteInput.Type: validate:"required"`
  - `ProblemWriteInput.Difficulty: validate:"required"`
  - `PaperWriteInput.Title: validate:"required"`
  - `ExportCreateInput.PaperID: validate:"required"`
  - `ExportCreateInput.Format: validate:"required"`
  - `ExportCreateInput.Variant: validate:"required"`

**IMPORTANT:** Validation tags are defined in domain types but NOT actually validated in handlers - they're present but unused.

**Handlers that decode JSON (11 locations):**
1. `handleCreateProblem` (problems.go:60)
2. `handleUpdateProblem` (problems.go:76)
3. `handlePreviewImport` (problems.go:154)
4. `handleCommitImport` (problems.go:165)
5. `handleBatchTagProblems` (problems.go:185)
6. `handleBatchDeleteProblems` (problems.go:200)
7. `handleCreateExport` (exports.go:16)
8. `handleCreatePaper` (papers.go:43)
9. `handleUpdatePaper` (papers.go:58)
10. `handleUpdatePaperItems` (papers.go:77)
11. `handleCreateTag` (tags.go:23)
12. `handleUpdateTag` (tags.go:39)
13. `handleMergeTag` (tags.go:68)
14. `handleUpdateImage` (images.go:110)
15. `handleEditImage` (images.go:156)

All follow identical pattern: `decodeJSON(r, &input)` with NO validation.

### 2. Worker - Broadcaster Implementation

**Location:** `/Users/cnhyk/Downloads/V0/server/internal/worker/broadcaster.go`

**Implementation Details:**
- **Type:** `Broadcaster` struct with `sync.RWMutex` and `subscribers map[chan []byte]struct{}`
- **Subscribe method:**
  - Creates buffered channel: `make(chan []byte, 8)`
  - Registers subscriber with read lock
  - Returns receive-only channel: `<-chan []byte`
  - Cleanup goroutine unsubscribes on context cancellation

- **Publish method:**
  - JSON marshals the payload
  - Acquires read lock
  - Non-blocking send to all subscribers (uses `select` with `default`)
  - Silently drops messages if channel is full

**Usage in API:** 
- SSE stream endpoint (exports.go:98): `ch := s.svc.Broadcaster().Subscribe(ctx)`
- Receives `[]byte` messages on channel in event loop

### 3. Export Manager - Progress Update Logic

**Location:** `/Users/cnhyk/Downloads/V0/server/internal/export/manager.go`

**Progress States:**
- 10% - After marking job as Processing (line 63)
- 40% - After LaTeX rendering completes (line 87)
- 100% - After file write/PDF compilation (line 115 for success, line 122 for failure)

**Key Methods:**
- `Start(ctx)`: Initializes worker goroutine (runs once via `sync.Once`)
- `Enqueue(jobID)`: Non-blocking enqueue with fallback goroutine
- `process(jobID)`: Main processing flow:
  1. Update status to Processing (10%)
  2. Publish job state
  3. Load paper details
  4. Render LaTeX template
  5. Update progress (40%)
  6. Create export directory
  7. Write file (PDF via xelatex or LaTeX)
  8. Update status to Done (100%)

**Progress Publishing:**
- After each state update, calls `publishJob()` which:
  - Retrieves full job record
  - Calls `m.broadcaster.Publish(job)` to notify all SSE clients

**Error Handling:**
- `fail()` method: Sets status to Failed, progress 100%, error message
- Also publishes job state on failure

### 4. Service Layer - All Files

**Location:** `/Users/cnhyk/Downloads/V0/server/internal/service/`

**Files:**
1. `service.go` - Service initialization
2. `exports.go` - Export service logic
3. `types.go` - Service parameter types
4. `problems.go` - Problem operations
5. `papers.go` - Paper operations
6. `search.go` - Search functionality
7. `images.go` - Image operations
8. `meta.go` - Metadata endpoints
9. `helpers.go` - Utility functions

**Export Service Details** (exports.go):
- `ListExports()`: Filters exports by status from snapshot
- `CreateExport()`: Validates paper exists, creates job record, enqueues export job

### 5. Domain Types - Struct Tags

**Location:** `/Users/cnhyk/Downloads/V0/server/internal/domain/types.go`

**Input Types with validate tags:**
- `ProblemWriteInput`:
  - `Latex: validate:"required"`
  - `Type: validate:"required"`
  - `Difficulty: validate:"required"`

- `PaperWriteInput`:
  - `Title: validate:"required"`

- `ExportCreateInput`:
  - `PaperID: validate:"required"`
  - `Format: validate:"required"`
  - `Variant: validate:"required"`

- `ImportPreviewRequest`:
  - No validate tags

**Response/Data Types:** No validate tags (as expected)

### 6. Docker Setup

**File:** `/Users/cnhyk/Downloads/V0/docker-compose.yml`

**Services:**
1. **PostgreSQL 16** (mathlib-postgres):
   - DB: mathlib / User: mathlib / Password: mathlib
   - Port: 5432
   - Health check configured

2. **Go Server** (mathlib-server):
   - Port: 8080
   - Depends on postgres health
   - Mounts ./server:/app
   - Environment: development config

**Environment Variables:**
- APP_ENV=development
- PORT=8080
- DATABASE_URL=postgres://mathlib:mathlib@postgres:5432/mathlib?sslmode=disable
- STORAGE_ROOT=/app/storage
- PUBLIC_BASE_URL=http://localhost:8080
- ALLOWED_ORIGINS=http://localhost:3000
- LOG_LEVEL=debug

### 7. Configuration Files

**File:** `/Users/cnhyk/Downloads/V0/server/.env.example`

**Contains:**
- APP_ENV, PORT, DATABASE_URL, STORAGE_ROOT
- PUBLIC_BASE_URL, ALLOWED_ORIGINS, LOG_LEVEL
- All matching docker-compose.yml values

---

## CRITICAL INSIGHTS

1. **Validation Gap:** Struct tags exist but are not validated in handlers
2. **Broadcaster Pattern:** Simple pub/sub with sync.RWMutex, 8-item buffer
3. **Progress Updates:** 10% → 40% → 100% workflow, published after each step
4. **No Explicit Validator Library:** No imports of validator/validation packages detected
5. **All JSON handlers:** Use single decodeJSON pattern (no variation)



