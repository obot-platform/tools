package tests

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/gptscript/pkg/sdkserver"
	"sigs.k8s.io/yaml"
)

func TestGPTScriptLoadTools(t *testing.T) {
	ctx := context.Background()

	registryURL := "github.com/obot-platform/tools"
	url, err := sdkserver.EmbeddedStart(ctx)
	if err != nil {
		t.Fatal(err)
	}

	client, err := gptscript.NewGPTScript(gptscript.GlobalOptions{
		URL: url,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Add assertions and test cases here
	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	idx, err := readRegistry(ctx, registryURL, client)
	if err != nil {
		t.Fatal(err)
	}
	allTools := []map[string]indexEntry{
		idx.Tools,
		idx.StepTemplates,
		idx.KnowledgeDataSources,
		idx.KnowledgeDocumentLoaders,
		idx.System,
		idx.ModelProviders,
		idx.AuthProviders,
	}

	var errs []error
	for _, toolMap := range allTools {
		for name, tool := range toolMap {
			if ref, ok := strings.CutPrefix(tool.Reference, "./"); ok {
				tool.Reference = registryURL + "/" + ref
			}
			_, err := client.LoadFile(ctx, "* from "+tool.Reference, gptscript.LoadOptions{})
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to load tool %s: %w", name, err))
			}
		}
	}
	if len(errs) > 0 {
		t.Fatal(fmt.Errorf("failed to load tools: %v", errs))
	}
}

type index struct {
	Tools                    map[string]indexEntry `json:"tools,omitempty"`
	StepTemplates            map[string]indexEntry `json:"stepTemplates,omitempty"`
	KnowledgeDataSources     map[string]indexEntry `json:"knowledgeDataSources,omitempty"`
	KnowledgeDocumentLoaders map[string]indexEntry `json:"knowledgeDocumentLoaders,omitempty"`
	System                   map[string]indexEntry `json:"system,omitempty"`
	ModelProviders           map[string]indexEntry `json:"modelProviders,omitempty"`
	AuthProviders            map[string]indexEntry `json:"authProviders,omitempty"`
}

type indexEntry struct {
	Reference string `json:"reference"`
}

func readRegistry(ctx context.Context, registryURL string, client *gptscript.GPTScript) (index, error) {
	run, err := client.Run(ctx, registryURL, gptscript.Options{})
	if err != nil {
		return index{}, err
	}

	out, err := run.Text()
	if err != nil {
		return index{}, err
	}

	var in index
	if err := yaml.Unmarshal([]byte(out), &in); err != nil {
		return index{}, err
	}

	return in, nil
}
