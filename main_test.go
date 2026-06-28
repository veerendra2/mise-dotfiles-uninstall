package main

import (
	"os"
	"os/exec"
	"path/filepath"
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

	// Run the compiled binary using our flags and environment variables
	cmd := exec.Command(binPath, "-c", configFile)
	cmd.Dir = tempWorkspace
	cmd.Env = append(os.Environ(), "HOME="+mockHome) // Override HOME env to test tilde expansion

	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("binary execution failed: %v, output: %s", err, string(output))
	}

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
