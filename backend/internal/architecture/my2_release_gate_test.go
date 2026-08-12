//go:build unit

package architecture

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type my2ReleaseWorkflow struct {
	Jobs map[string]my2ReleaseJob `yaml:"jobs"`
}

type my2ReleaseJob struct {
	Needs my2ReleaseNeeds  `yaml:"needs"`
	Steps []my2ReleaseStep `yaml:"steps"`
}

type my2ReleaseNeeds []string

func (n *my2ReleaseNeeds) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		*n = my2ReleaseNeeds{value.Value}
		return nil
	}
	var values []string
	if err := value.Decode(&values); err != nil {
		return err
	}
	*n = my2ReleaseNeeds(values)
	return nil
}

type my2ReleaseStep struct {
	Name string            `yaml:"name"`
	Run  string            `yaml:"run"`
	Uses string            `yaml:"uses"`
	With map[string]string `yaml:"with"`
	Env  map[string]string `yaml:"env"`
}

func TestMy2ReleaseRequiresCompleteQualityGate(t *testing.T) {
	workflow := loadMy2ReleaseWorkflow(t)
	source := requireMy2ReleaseJob(t, workflow, "source-gate")
	backend := requireMy2ReleaseJob(t, workflow, "backend-gate")
	frontend := requireMy2ReleaseJob(t, workflow, "frontend-gate")
	lint := requireMy2ReleaseJob(t, workflow, "lint-gate")
	release := requireMy2ReleaseJob(t, workflow, "release")

	for _, dependency := range []string{"source-gate", "backend-gate", "frontend-gate", "lint-gate"} {
		if !slices.Contains([]string(release.Needs), dependency) {
			t.Errorf("release job must need %q", dependency)
		}
	}
	requireMy2ReleaseText(t, joinedMy2ReleaseRuns(source), "git merge-base --is-ancestor", "origin/my2.0")
	requireMy2ReleaseText(t, joinedMy2ReleaseRuns(backend), "make test-unit", "make test-integration")
	requireMy2ReleaseText(t, joinedMy2ReleaseRuns(frontend), "pnpm run test:run", "pnpm run typecheck", "pnpm run lint:check")
	requireMy2ReleaseText(t, joinedMy2ReleaseUses(lint), "golangci/golangci-lint-action")
	requireMy2ReleaseText(t, joinedMy2ReleaseWith(lint, "args"), "--build-tags=unit", "--concurrency=2")
}

func TestMy2ReleasePublishesLatestOnlyAfterValidatedRelease(t *testing.T) {
	workflow := loadMy2ReleaseWorkflow(t)
	release := requireMy2ReleaseJob(t, workflow, "release")

	versionedIndex := my2ReleaseStepIndex(t, release, "Build and publish versioned my2 image")
	offlineIndex := my2ReleaseStepIndex(t, release, "Create and validate offline Docker package")
	releaseIndex := my2ReleaseStepIndex(t, release, "Publish my2 prerelease")
	latestIndex := my2ReleaseStepIndex(t, release, "Publish my2 latest tag")
	if versionedIndex >= offlineIndex || offlineIndex >= releaseIndex || releaseIndex >= latestIndex {
		t.Fatalf("release steps must publish version, validate archive, create release, then update latest")
	}

	versionedRun := release.Steps[versionedIndex].Run
	requireMy2ReleaseText(t, versionedRun, `docker push "$IMAGE:my2-$VERSION"`, `org.opencontainers.image.revision=$COMMIT_SHA`)
	if commitSource := release.Steps[versionedIndex].Env["COMMIT_SHA"]; !strings.Contains(commitSource, "source-gate.outputs.commit_sha") {
		t.Fatalf("versioned image step must receive the validated commit SHA, got %q", commitSource)
	}
	if strings.Contains(versionedRun, `docker push "$IMAGE:my2-latest"`) {
		t.Fatal("versioned image step must not update my2-latest")
	}
	requireMy2ReleaseText(t, release.Steps[offlineIndex].Run, "gzip -t", "sha256sum -c", "docker image inspect", "linux/amd64")
	requireMy2ReleaseText(t, release.Steps[latestIndex].Run, `docker push "$IMAGE:my2-latest"`)
}

func TestMy2IntegrationTargetExcludesNonIntegrationPackages(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve architecture test path")
	}
	path := filepath.Join(filepath.Dir(filename), "..", "..", "Makefile")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	makefile := string(content)
	for _, packagePath := range []string{
		"./internal/repository",
		"./internal/middleware",
		"./internal/server/routes",
		"./internal/pkg/tlsfingerprint",
		"./ent/migrate",
		"./migrations",
		"./cmd/...",
	} {
		if !strings.Contains(makefile, packagePath) {
			t.Errorf("integration target is missing %s", packagePath)
		}
	}
	if strings.Contains(makefile, "go test -tags=integration ./...") {
		t.Fatal("integration target must not compile unrelated packages with ./...")
	}
	if !strings.Contains(makefile, "go test -p 1 -tags=integration") {
		t.Fatal("integration target must serialize package execution to avoid testcontainer resource contention")
	}
}

func TestMy2LintGateFocusesOnProductionCorrectness(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve architecture test path")
	}
	path := filepath.Join(filepath.Dir(filename), "..", "..", ".golangci.yml")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	config := string(content)
	for _, linter := range []string{"errcheck", "govet", "ineffassign", "staticcheck"} {
		if !strings.Contains(config, "    - "+linter+"\n") {
			t.Errorf("release lint gate must enable %s", linter)
		}
	}
	if strings.Contains(config, "    - unused\n") {
		t.Fatal("unused must remain disabled while deferred protocol adapters are intentionally retained")
	}
	for _, exclusion := range []string{"'(.+)_test\\.go'", "'^(ST|QF)[0-9]+:'", "'^SA1008:'", "'^SA1019:'"} {
		if !strings.Contains(config, exclusion) {
			t.Errorf("release lint gate is missing scoped exclusion %s", exclusion)
		}
	}
}

func loadMy2ReleaseWorkflow(t *testing.T) my2ReleaseWorkflow {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve architecture test path")
	}
	path := filepath.Join(filepath.Dir(filename), "..", "..", "..", ".github", "workflows", "my2-release.yml")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var workflow my2ReleaseWorkflow
	if err := yaml.Unmarshal(content, &workflow); err != nil {
		t.Fatalf("parse my2 release workflow: %v", err)
	}
	return workflow
}

func requireMy2ReleaseJob(t *testing.T, workflow my2ReleaseWorkflow, name string) my2ReleaseJob {
	t.Helper()
	job, ok := workflow.Jobs[name]
	if !ok {
		t.Fatalf("my2 release workflow is missing %q job", name)
	}
	return job
}

func joinedMy2ReleaseRuns(job my2ReleaseJob) string {
	var runs []string
	for _, step := range job.Steps {
		runs = append(runs, step.Run)
	}
	return strings.Join(runs, "\n")
}

func joinedMy2ReleaseUses(job my2ReleaseJob) string {
	var uses []string
	for _, step := range job.Steps {
		uses = append(uses, step.Uses)
	}
	return strings.Join(uses, "\n")
}

func joinedMy2ReleaseWith(job my2ReleaseJob, key string) string {
	var values []string
	for _, step := range job.Steps {
		values = append(values, step.With[key])
	}
	return strings.Join(values, "\n")
}

func requireMy2ReleaseText(t *testing.T, text string, values ...string) {
	t.Helper()
	for _, value := range values {
		if !strings.Contains(text, value) {
			t.Errorf("my2 release workflow is missing %q", value)
		}
	}
}

func my2ReleaseStepIndex(t *testing.T, job my2ReleaseJob, name string) int {
	t.Helper()
	for index, step := range job.Steps {
		if step.Name == name {
			return index
		}
	}
	t.Fatalf("my2 release job is missing %q step", name)
	return -1
}
