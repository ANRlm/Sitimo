package service_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"mathlib/server/internal/domain"
	"mathlib/server/internal/service"
	"mathlib/server/internal/store"
	"mathlib/server/internal/worker"

	"github.com/rs/zerolog"
)

func TestDemoSeedEnvelopeBackwardCompatible(t *testing.T) {
	legacyPayload := []byte(`{
		"tags": [],
		"problems": [
			{
				"id": "demo-problem-legacy",
				"code": "P-2024-0001",
				"latex": "legacy",
				"type": "solve",
				"difficulty": "easy",
				"tagIds": [],
				"imageIds": [],
				"createdAt": "2026-04-01T00:00:00Z",
				"updatedAt": "2026-04-01T00:00:00Z",
				"version": 1,
				"isDeleted": false
			}
		],
		"images": [],
		"papers": [],
		"exportJobs": [],
		"settings": {}
	}`)

	var envelope service.SeedEnvelope
	if err := json.Unmarshal(legacyPayload, &envelope); err != nil {
		t.Fatalf("unmarshal legacy demo seed: %v", err)
	}
	if len(envelope.Problems) != 1 {
		t.Fatalf("expected 1 legacy problem, got %d", len(envelope.Problems))
	}
	if len(envelope.Problems[0].Versions) != 0 {
		t.Fatalf("expected legacy problem to omit versions, got %d", len(envelope.Problems[0].Versions))
	}
	if len(envelope.SearchHistory) != 0 || len(envelope.SavedSearches) != 0 {
		t.Fatal("expected legacy demo seed to omit search collections")
	}
}

func TestDemoSeedLifecycle(t *testing.T) {
	ctx := context.Background()
	repo, svc := newSeedTestService(t, ctx)

	if err := svc.SeedDemoData(ctx); err != nil {
		t.Fatalf("seed demo data: %v", err)
	}

	versions, err := repo.ListProblemVersions(ctx, "demo-problem-2")
	if err != nil {
		t.Fatalf("list seeded problem versions: %v", err)
	}
	if len(versions) < 2 {
		t.Fatalf("expected demo-problem-2 to have multiple versions, got %d", len(versions))
	}

	history, err := repo.ListSearchHistory(ctx, 20)
	if err != nil {
		t.Fatalf("list demo search history: %v", err)
	}
	if len(history) < 5 {
		t.Fatalf("expected demo search history entries, got %d", len(history))
	}

	saved, err := repo.ListSavedSearches(ctx)
	if err != nil {
		t.Fatalf("list saved searches: %v", err)
	}
	if len(saved) < 4 {
		t.Fatalf("expected demo saved searches, got %d", len(saved))
	}

	exports, err := repo.ListExportJobs(ctx, store.ExportListOptions{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list export jobs: %v", err)
	}
	statuses := make([]domain.ExportStatus, 0, len(exports.Items))
	for _, item := range exports.Items {
		statuses = append(statuses, item.Status)
	}
	for _, status := range []domain.ExportStatus{
		domain.ExportStatusPending,
		domain.ExportStatusProcessing,
		domain.ExportStatusDone,
		domain.ExportStatusFailed,
	} {
		if !slices.Contains(statuses, status) {
			t.Fatalf("expected export status %q in demo data, got %v", status, statuses)
		}
	}

	exported, err := repo.ExportAll(ctx)
	if err != nil {
		t.Fatalf("export all demo data: %v", err)
	}
	if len(exported.SearchHistory) != len(history) {
		t.Fatalf("expected exported search history count %d, got %d", len(history), len(exported.SearchHistory))
	}
	if len(exported.SavedSearches) != len(saved) {
		t.Fatalf("expected exported saved searches count %d, got %d", len(saved), len(exported.SavedSearches))
	}

	problem2 := findSeedProblem(exported.Problems, "demo-problem-2")
	if problem2 == nil || len(problem2.Versions) < 2 {
		t.Fatalf("expected exported problem versions for demo-problem-2, got %+v", problem2)
	}

	if err := repo.ImportAll(ctx, exported); err != nil {
		t.Fatalf("round-trip import all: %v", err)
	}

	reloadedHistory, err := repo.ListSearchHistory(ctx, 20)
	if err != nil {
		t.Fatalf("list reloaded search history: %v", err)
	}
	if len(reloadedHistory) != len(history) {
		t.Fatalf("expected reloaded history count %d, got %d", len(history), len(reloadedHistory))
	}
}

func TestLoadDemoDataIdempotentAndPreservesSettings(t *testing.T) {
	ctx := context.Background()
	repo, svc := newSeedTestService(t, ctx)

	customSettings := domain.SettingsPayload{
		"profile": map[string]any{
			"name":  "自定义用户",
			"notes": "不要被增量示例数据覆盖",
		},
	}
	if err := repo.UpsertSettings(ctx, customSettings); err != nil {
		t.Fatalf("seed custom settings: %v", err)
	}

	firstStats, err := svc.LoadDemoData(ctx)
	if err != nil {
		t.Fatalf("first load demo data: %v", err)
	}
	secondStats, err := svc.LoadDemoData(ctx)
	if err != nil {
		t.Fatalf("second load demo data: %v", err)
	}
	if firstStats["problems"] == 0 || secondStats["problems"] == 0 {
		t.Fatalf("expected load demo data to report problem inserts, got %v and %v", firstStats, secondStats)
	}

	statusLoaded, statusStats, err := svc.GetDemoDataStatus(ctx)
	if err != nil {
		t.Fatalf("get demo data status: %v", err)
	}
	if !statusLoaded {
		t.Fatal("expected demo data status to be loaded")
	}
	if statusStats["problems"] == 0 || statusStats["problemVersions"] == 0 || statusStats["searchHistory"] == 0 || statusStats["savedSearches"] == 0 {
		t.Fatalf("expected extended demo stats after load, got %v", statusStats)
	}

	settings, err := repo.GetSettings(ctx)
	if err != nil {
		t.Fatalf("get settings after load demo data: %v", err)
	}
	profile, ok := settings["profile"].(map[string]any)
	if !ok {
		t.Fatalf("expected profile settings to remain present, got %#v", settings["profile"])
	}
	if profile["name"] != "自定义用户" {
		t.Fatalf("expected custom settings to survive incremental load, got %#v", profile)
	}
}

func TestClearDemoDataOnlyRemovesDemoRecords(t *testing.T) {
	ctx := context.Background()
	repo, svc := newSeedTestService(t, ctx)

	if err := svc.SeedDemoData(ctx); err != nil {
		t.Fatalf("seed demo data: %v", err)
	}

	tag := mustCreateTag(t, repo, "保留标签")
	created, _, err := svc.CreateProblem(ctx, domain.ProblemWriteInput{
		Latex:      "A retained non-demo problem.",
		Type:       domain.ProblemTypeSolve,
		Difficulty: domain.DifficultyMedium,
		TagIDs:     []string{tag.ID},
	})
	if err != nil {
		t.Fatalf("create retained problem: %v", err)
	}
	if err := repo.CreateSearchHistory(ctx, "retain", map[string]any{"formula": "\\sin x"}, 1); err != nil {
		t.Fatalf("create retained search history: %v", err)
	}
	if _, err := repo.CreateSavedSearch(ctx, "retain", "retain", map[string]any{"conditions": []map[string]any{{"field": "grade", "operator": "eq", "value": "高二"}}}); err != nil {
		t.Fatalf("create retained saved search: %v", err)
	}

	stats, err := svc.ClearDemoData(ctx)
	if err != nil {
		t.Fatalf("clear demo data: %v", err)
	}
	if stats["problems"] == 0 || stats["searchHistory"] == 0 || stats["savedSearches"] == 0 {
		t.Fatalf("expected clear demo data to remove extended demo entities, got %v", stats)
	}

	loaded, demoStats, err := svc.GetDemoDataStatus(ctx)
	if err != nil {
		t.Fatalf("get demo data status after clear: %v", err)
	}
	if loaded {
		t.Fatalf("expected demo data status to be clear, got %v", demoStats)
	}

	problems, err := repo.ListProblems(ctx, store.ProblemListOptions{Page: 1, PageSize: 50, IncludeDeleted: true})
	if err != nil {
		t.Fatalf("list retained problems: %v", err)
	}
	if !slices.ContainsFunc(problems.Items, func(item domain.Problem) bool { return item.ID == created.ID }) {
		t.Fatalf("expected retained problem %s to survive clear", created.ID)
	}

	history, err := repo.ListSearchHistory(ctx, 20)
	if err != nil {
		t.Fatalf("list retained search history: %v", err)
	}
	if len(history) != 1 || history[0].Query != "retain" {
		t.Fatalf("expected retained search history to survive clear, got %+v", history)
	}

	saved, err := repo.ListSavedSearches(ctx)
	if err != nil {
		t.Fatalf("list retained saved searches: %v", err)
	}
	if len(saved) != 1 || saved[0].Name != "retain" {
		t.Fatalf("expected retained saved search to survive clear, got %+v", saved)
	}
}

func newSeedTestService(t *testing.T, ctx context.Context) (*store.Repository, *service.Service) {
	t.Helper()

	cfg, cleanup := loadIsolatedTestConfig(t)
	cfg.PublicBaseURL = "http://localhost:8080"
	cfg.StorageRoot = t.TempDir()
	t.Cleanup(cleanup)

	repo, err := store.NewRepository(ctx, cfg)
	if err != nil {
		t.Fatalf("create test repository: %v", err)
	}
	t.Cleanup(repo.Close)

	applyMigrations(t, repo.DB())

	svc := service.New(cfg, repo, zerolog.Nop(), worker.NewBroadcaster(), nil)
	return repo, svc
}

func findSeedProblem(problems []domain.SeedProblem, id string) *domain.SeedProblem {
	for idx := range problems {
		if problems[idx].ID == id {
			return &problems[idx]
		}
	}
	return nil
}

func TestDemoSeedJSONDecodes(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "testdata", "demo-data.json"))
	if err != nil {
		t.Fatalf("read demo data json: %v", err)
	}
	var envelope service.SeedEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("decode demo data json: %v", err)
	}
	if len(envelope.SearchHistory) == 0 || len(envelope.SavedSearches) == 0 {
		t.Fatalf("expected decoded demo seed to include search collections, got history=%d saved=%d", len(envelope.SearchHistory), len(envelope.SavedSearches))
	}
	if problem := findSeedProblem(envelope.Problems, "demo-problem-2"); problem == nil || len(problem.Versions) < 2 {
		t.Fatalf("expected decoded demo seed to include nested problem versions, got %+v", problem)
	}
}
