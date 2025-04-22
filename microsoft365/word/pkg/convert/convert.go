package convert

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func MarkdownToDocx(in string) ([]byte, error) {
	tempfile, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("word-convsource-*.md"))
	if err != nil {
		return nil, err
	}
	defer tempfile.Close()

	p := tempfile.Name()
	_, err = tempfile.WriteString(in)
	if err != nil {
		return nil, err
	}
	_ = tempfile.Close()
	defer os.Remove(p)

	var cmd *exec.Cmd
	var outputFile string
	if _, err := exec.LookPath("pandoc"); err == nil {
		cmd, outputFile = pandocCmd(p)
		slog.Info("Used pandoc to convert markdown to docx", "input", p, "output", outputFile)
	} else if _, err := exec.LookPath("soffice"); err == nil {
		var cleanupFunc func()
		cmd, cleanupFunc, outputFile, err = sofficeCmd(p)
		if err != nil {
			return nil, err
		}
		slog.Info("Used soffice to convert markdown to docx", "input", p, "output", outputFile)
		defer cleanupFunc()
	} else {
		return nil, fmt.Errorf("neither pandoc nor soffice binary found")
	}

	// capture stdout and stderr in a buffer
	var outb, errb strings.Builder
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err = cmd.Run()
	if err != nil {
		slog.Error("Failed to run pandoc/soffice command", "error", err, "stderr", errb.String(), "stdout", outb.String())
		return nil, err
	}

	slog.Info("pandoc/soffice command output", "stdout", outb.String(), "stderr", errb.String())

	content, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, err
	}
	return content, os.Remove(outputFile)
}

func pandocCmd(p string) (*exec.Cmd, string) {
	outFile := fmt.Sprintf("%s.docx", p)
	return exec.Command(
		"pandoc",
		"-f", "markdown",
		"-t", "docx",
		"--output", outFile,
		p,
	), outFile
}

func sofficeCmd(p string) (*exec.Cmd, func(), string, error) {
	var err error
	p, err = markdownToHTML(p)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to convert markdown to html: %w", err)
	}

	profileDir, err := os.MkdirTemp(os.TempDir(), "libreoffice-profile-*")
	if err != nil {
		slog.Error("Failed to create soffice profile directory", "path", profileDir, "error", err)
		return nil, nil, "", fmt.Errorf("failed to create soffice profile directory: %w", err)
	}
	out := strings.TrimSuffix(p, filepath.Ext(p)) + ".docx"
	return exec.Command(
			"soffice",
			"--headless",
			fmt.Sprintf("-env:UserInstallation=file://%s", profileDir),
			"--convert-to", "docx:Office Open XML Text",
			"--outdir", filepath.Dir(out),
			p,
		), func() {
			_ = os.Remove(p)
			_ = os.RemoveAll(profileDir)
		}, out, nil
}

func markdownToHTML(p string) (string, error) {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	pars := parser.NewWithExtensions(extensions)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	md, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}

	outFile := fmt.Sprintf("%s.html", p)
	if err = os.WriteFile(outFile, markdown.ToHTML(md, pars, renderer), 0644); err != nil {
		return "", err
	}
	return outFile, nil
}
