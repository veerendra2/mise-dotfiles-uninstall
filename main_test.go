package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestUninstall(t *testing.T) {
	tempWorkspace, err := os.MkdirTemp("", "mise-workspace-*")
	if err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempWorkspace) }()

	mockHome := filepath.Join(tempWorkspace, "home")
	mockSrcDir := filepath.Join(mockHome, "projects", "dotfiles-refactor")
	mockSrcFishDir := filepath.Join(mockSrcDir, "tools", "fish")

	_ = os.MkdirAll(mockHome, 0755)
	_ = os.MkdirAll(mockSrcFishDir, 0755)

	bashrcSrc := filepath.Join(mockSrcDir, "bash", "bashrc")
	_ = os.MkdirAll(filepath.Dir(bashrcSrc), 0755)
	_ = os.WriteFile(bashrcSrc, []byte("alias ll='ls -l'"), 0644)

	fishConfigSrc := filepath.Join(mockSrcFishDir, "config.fish")
	_ = os.WriteFile(fishConfigSrc, []byte("echo hello fish"), 0644)

	targetBashrc := filepath.Join(mockHome, ".bashrc")
	targetFishDir := filepath.Join(mockHome, ".config", "fish")

	_ = os.MkdirAll(targetFishDir, 0755)
	_ = os.Symlink(bashrcSrc, targetBashrc)

	targetFishConfig := filepath.Join(targetFishDir, "config.fish")
	_ = os.Symlink(fishConfigSrc, targetFishConfig)

	safeFile := filepath.Join(targetFishDir, "user_manual_config.fish")
	_ = os.WriteFile(safeFile, []byte("custom manual configs"), 0644)

	miseTomlContent := `
[settings]
experimental = true
dotfiles.root = "~/projects/dotfiles-refactor"
dotfiles.default_mode = "symlink"

[dotfiles]
"~/.bashrc" = "bash/bashrc"
"~/.config/fish" = { source = "tools/fish", mode = "symlink-each" }
`
	configFile := filepath.Join(tempWorkspace, "test-mise.toml")
	_ = os.WriteFile(configFile, []byte(miseTomlContent), 0644)

	// Build the real main.go to a temporary binary in the workspace
	binPath := filepath.Join(tempWorkspace, "mise-dotfiles-uninstall")
	buildCmd := exec.Command("go", "build", "-o", binPath, "main.go")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v, output: %s", err, string(output))
	}

	// 1. Run the compiled binary using our flags (with -d) and environment variables
	cmdDry := exec.Command(binPath, "-c", configFile, "-d")
	cmdDry.Dir = tempWorkspace
	cmdDry.Env = append(os.Environ(), "HOME="+mockHome) // Override HOME env to test tilde expansion

	dryOutput, err := cmdDry.CombinedOutput()
	if err != nil {
		t.Fatalf("dry run binary execution failed: %v, output: %s", err, string(dryOutput))
	}

	dryStr := string(dryOutput)
	if !strings.Contains(dryStr, "✓ Would unlink: ~/.bashrc") {
		t.Errorf("dry-run output should list ~/.bashrc, got:\n%s", dryStr)
	}
	if !strings.Contains(dryStr, "✓ Would unlink: ~/.config/fish/config.fish") {
		t.Errorf("dry-run output should list fish config, got:\n%s", dryStr)
	}
	if strings.Contains(dryStr, "All symlinks are removed.") {
		t.Errorf("dry-run output should not print completion message")
	}

	// Verify symlinks physically still exist
	if _, err := os.Lstat(targetBashrc); os.IsNotExist(err) {
		t.Errorf("target bashrc symlink was deleted during dry run")
	}
	if _, err := os.Lstat(targetFishConfig); os.IsNotExist(err) {
		t.Errorf("target fish config symlink was deleted during dry run")
	}

	// 2. Second execution: Real Run (no -d)
	cmdReal := exec.Command(binPath, "-c", configFile)
	cmdReal.Dir = tempWorkspace
	cmdReal.Env = append(os.Environ(), "HOME="+mockHome)

	realOutput, err := cmdReal.CombinedOutput()
	if err != nil {
		t.Fatalf("real run binary execution failed: %v, output: %s", err, string(realOutput))
	}

	realStr := string(realOutput)
	if strings.Contains(realStr, "✓ Would unlink:") || strings.Contains(realStr, "✓ Unlinked:") {
		t.Errorf("real run output should not list details, got:\n%s", realStr)
	}
	if !strings.Contains(realStr, "All symlinks are removed.") {
		t.Errorf("real run output should print completion message, got:\n%s", realStr)
	}

	// Verify files are physically removed/kept
	if _, err := os.Lstat(targetBashrc); !os.IsNotExist(err) {
		t.Errorf("target bashrc symlink was not deleted")
	}

	if _, err := os.Lstat(targetFishConfig); !os.IsNotExist(err) {
		t.Errorf("target fish config symlink was not deleted")
	}

	if _, err := os.Stat(safeFile); err != nil {
		t.Errorf("regular user file in fish dir was deleted or is inaccessible: %v", err)
	}
}
