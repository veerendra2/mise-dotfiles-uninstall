package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

var Version = "dev"

type Config struct {
	Settings map[string]any `toml:"settings"`
	Dotfiles map[string]any `toml:"dotfiles"`
}

func main() {
	var configFile string
	var showVersion bool
	var dryRun bool

	flag.StringVar(&configFile, "c", "mise.toml", "Path to mise.toml configuration file")
	flag.BoolVar(&dryRun, "d", false, "Show what would be unlinked without deleting (Dry run)")
	flag.BoolVar(&showVersion, "v", false, "Display version information")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "The missing uninstaller companion for mise's dotfiles (Until native support).\n\nUsage:\n  mise-dotfiles-uninstall [options]\n\nOptions:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if showVersion {
		fmt.Println(Version)
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if absConfig, err := filepath.Abs(configFile); err == nil {
		configFile = absConfig
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	expand := func(p string) string {
		if p == "~" {
			return home
		}
		if strings.HasPrefix(p, "~/") {
			return filepath.Join(home, p[2:])
		}
		return p
	}

	// Extract and expand dotfiles.root
	var rootDir string
	if settings, ok := config.Settings["dotfiles"].(map[string]any); ok {
		if r, ok := settings["root"].(string); ok {
			rootDir = r
		}
	}
	if rootDir == "" {
		if r, ok := config.Settings["dotfiles.root"].(string); ok {
			rootDir = r
		}
	}

	var expandedRootDir string
	if rootDir != "" {
		expandedRootDir = expand(rootDir)
	} else {
		expandedRootDir = filepath.Dir(configFile)
	}

	unlink := func(path string) {
		if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
			pretty := path
			if home != "" && strings.HasPrefix(path, home) {
				pretty = "~" + strings.TrimPrefix(path, home)
			}

			if dryRun {
				fmt.Printf("✓ Would unlink: %s\n", pretty)
			} else {
				_ = os.Remove(path)
			}
		}
	}

	for target, valRaw := range config.Dotfiles {
		src := ""
		mode := "symlink"

		switch val := valRaw.(type) {
		case string:
			src = val
		case map[string]any:
			if s, ok := val["source"].(string); ok {
				src = s
			} else if p, ok := val["path"].(string); ok {
				src = p
			}
			if m, ok := val["mode"].(string); ok {
				mode = m
			}
		}

		targetPath := expand(target)

		var srcPath string
		if src != "" {
			expandedSrc := expand(src)
			if filepath.IsAbs(expandedSrc) || strings.HasPrefix(src, "~") {
				srcPath = expandedSrc
			} else {
				srcPath = filepath.Join(expandedRootDir, expandedSrc)
			}
		}

		if mode == "symlink-each" {
			entries, _ := os.ReadDir(srcPath)
			for _, entry := range entries {
				unlink(filepath.Join(targetPath, entry.Name()))
			}
		} else {
			unlink(targetPath)
		}
	}

	if !dryRun {
		fmt.Println("✓ All symlinks are removed.")
	}
}
