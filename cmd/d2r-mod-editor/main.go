package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"d2r-multiboxing/internal/d2r"
	"d2r-multiboxing/internal/editor"
	"d2r-multiboxing/internal/modfile"
	"d2r-multiboxing/internal/process"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	modsDir := findModsDir()

	mods, err := modfile.DiscoverMods(modsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(mods) == 0 {
		fmt.Fprintln(os.Stderr, "No mods found in:", modsDir)
		os.Exit(1)
	}

	// If only one mod, load it directly; otherwise let user pick
	var modName string
	if len(mods) == 1 {
		modName = mods[0]
	} else {
		modName = pickMod(mods)
	}

	modDir := filepath.Join(modsDir, modName)
	mod, err := modfile.Load(modDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading mod: %v\n", err)
		os.Exit(1)
	}

	m := editor.New(mod)
	p := tea.NewProgram(m, tea.WithAltScreen())

	result, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Handle D2R restart request
	if finalModel, ok := result.(editor.Model); ok && finalModel.RestartRequested {
		restartD2R(modName)
	}
}

// restartD2R terminates running D2R processes and relaunches with the mod.
func restartD2R(modName string) {
	fmt.Println("\n🔄 Restarting D2R with mod:", modName)

	// Find and terminate existing D2R processes
	procs, err := process.FindProcessesByName(d2r.ProcessName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not find D2R processes: %v\n", err)
	}

	for _, p := range procs {
		fmt.Printf("  Terminating D2R (PID %d)...\n", p.PID)
		handle, err := os.FindProcess(int(p.PID))
		if err == nil {
			_ = handle.Kill()
		}
	}

	// Relaunch D2R with mod parameter
	d2rPath := d2r.DefaultGamePath
	cmd := exec.Command(d2rPath, "-mod", modName, "-txt")
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start D2R: %v\n", err)
		fmt.Fprintln(os.Stderr, "You can manually start D2R with:")
		fmt.Fprintf(os.Stderr, "  \"%s\" -mod %s -txt\n", d2rPath, modName)
		os.Exit(1)
	}

	fmt.Printf("  ✔ D2R started (PID %d) with -mod %s\n", cmd.Process.Pid, modName)
}

// findModsDir locates the mods/ directory.
// Priority: command-line arg > ./mods/ > executable-relative mods/
func findModsDir() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	}

	// Check current directory
	if info, err := os.Stat("mods"); err == nil && info.IsDir() {
		return "mods"
	}

	// Check next to executable
	exe, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exe)
		modsDir := filepath.Join(exeDir, "mods")
		if info, err := os.Stat(modsDir); err == nil && info.IsDir() {
			return modsDir
		}
	}

	// Fallback
	return "mods"
}

// pickMod shows a simple numbered menu for mod selection.
func pickMod(mods []string) string {
	fmt.Println("Available mods:")
	for i, name := range mods {
		fmt.Printf("  [%d] %s\n", i+1, name)
	}

	fmt.Print("\nSelect mod (number): ")
	var choice int
	if _, err := fmt.Scan(&choice); err != nil || choice < 1 || choice > len(mods) {
		fmt.Fprintln(os.Stderr, "Invalid selection")
		os.Exit(1)
	}

	return mods[choice-1]
}
