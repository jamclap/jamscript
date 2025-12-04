package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

var (
	targetDir  = "target"
	targetFile = exe(filepath.Join(targetDir, "jams"))
)

func main() {
	if !build() {
		os.Exit(1)
	}
	report()
}

func build() bool {
	start := time.Now()
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		fmt.Println("mkdir failed:", err)
		return false
	}
	cmd := exec.Command("go", "build", "-ldflags=-s -w", "-o", targetFile, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	end := time.Now()
	fmt.Printf("Build time: %.2f s\n", end.Sub(start).Seconds())
	return err == nil
}

func exe(p string) string {
	if runtime.GOOS == "windows" {
		return p + ".exe"
	}
	return p
}

func report() {
	reportSize(targetFile)
}

func reportSize(path string) {
	info, err := os.Stat(path)
	if err != nil {
		fmt.Println("stat failed:", err)
		return
	}
	size := info.Size()
	fmt.Printf("Size of %s: %d B or %.2f MiB\n",
		path, size, float64(size)/(1<<20))
}
