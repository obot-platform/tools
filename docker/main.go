package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/google/safeopen"
	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/gptscript/pkg/env"
)

var ctx = context.Background()

func main() {
	if err := mainErr(); err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}
}

func writeFiles(ctx context.Context, c *gptscript.GPTScript, workspaceDir string) error {
	files, err := c.ListFilesInWorkspace(ctx)
	if err != nil {
		return fmt.Errorf("failed to list files in workspace: %w", err)
	}

	for _, file := range files {
		data, err := c.ReadFileInWorkspace(ctx, file)
		if err != nil {
			return fmt.Errorf("failed to read file %q: %w", file, err)
		}
		f, err := safeopen.CreateBeneath(workspaceDir, file)
		if err != nil {
			return err
		}
		_, err = f.Write(data)
		if err != nil {
			f.Close()
			return fmt.Errorf("failed to write file %q: %w", file, err)
		}

		if err := f.Close(); err != nil {
			return fmt.Errorf("failed to close file %q: %w", file, err)
		}
	}

	f, err := safeopen.CreateBeneath(workspaceDir, ".stamp")
	if err != nil {
		return fmt.Errorf("failed to create stamp file: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close stamp file: %w", err)
	}
	return nil
}

func getWorkspaceDir() (string, func(), error) {
	workspaceDir := os.Getenv("GPTSCRIPT_WORKSPACE_DIR")
	if workspaceDir != "" {
		return workspaceDir, func() {}, nil
	}

	workspaceDir, err := os.MkdirTemp("", "docker-tool-workspace")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temporary workspace directory: %w", err)
	}
	return workspaceDir, func() {
		os.RemoveAll(workspaceDir)
	}, nil
}

func mainErr() (err error) {
	workspaceDir, close, err := getWorkspaceDir()
	if err != nil {
		return err
	}
	defer close()

	c, err := gptscript.NewGPTScript()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := writeFiles(ctx, c, workspaceDir); err != nil {
		return err
	}

	var (
		image = gptscript.GetEnv("IMAGE", "")
		envs  = gptscript.GetEnv("ENVS", "")
	)

	if image == "" {
		return fmt.Errorf("missing IMAGE environment variable")
	}

	args := []string{
		"run", "--rm",
		"-v", workspaceDir + ":/mnt/data",
	}

	moreArgs := os.Getenv("OBOT_DOCKER_ARGS")
	for _, arg := range strings.Fields(moreArgs) {
		if arg != "" {
			args = append(args, arg)
		}
	}

	for _, env := range strings.Split(envs, ",") {
		env = strings.TrimSpace(env)
		if env == "" {
			continue
		}
		args = append(args, "-e", env)
	}

	inputString := gptscript.GetEnv("GPTSCRIPT_INPUT", "")
	input := map[string]any{}

	if inputString != "" {
		if err := json.Unmarshal([]byte(inputString), &input); err != nil {
			// ignore
		}
		args = append(args, "-e", "GPTSCRIPT_INPUT="+inputString)
	}

	for k, v := range input {
		if s, _ := v.(string); s != "" {
			k = strings.ToUpper(env.ToEnvLike(k))
			args = append(args, "-e", fmt.Sprintf("%s=%s", k, s))
		}
	}

	args = append(args, image)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
