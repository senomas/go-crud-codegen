package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func gitStatus() {
	out, err := exec.Command("git", "--no-pager", "status", "--porcelain", "--", "model/*.go", "migrations/*.sql").CombinedOutput()
	if err != nil {
		log.Fatalf("Checking git status: %v", err)
	}
	scanner := bufio.NewScanner(bytes.NewBuffer(out))
	clean := true
	for scanner.Scan() {
		ln := scanner.Text()
		if strings.HasSuffix(ln, "_test.go") {
			// skip
		} else {
			clean = false
			fmt.Printf("MODIFIED: %s\n", ln)
		}
	}
	if !clean {
		log.Fatalf("Repository not clean, aborted")
	}
}

func diff(file string) {
	if _, err := os.Stat(file); err == nil {
		// continue
	} else if os.IsNotExist(err) {
		return
	} else {
		log.Fatalf("Error checking %s: %v\n", file, err)
	}
	out, err := exec.Command("git", "--no-pager", "diff", "--", file).CombinedOutput()
	if err != nil {
		log.Fatalf("Checking git diff for %s: %v", file, err)
	}
	scanner := bufio.NewScanner(bytes.NewBuffer(out))
	for scanner.Scan() {
		ln := scanner.Text()
		if strings.HasPrefix(ln, "@@ ") {
			break
		}
	}
	changed := false
	re := regexp.MustCompile(`^(?:[+-]\s*(?:--|//)\s*GENERATED:\s*\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})\s*|[+-]\s*)$`)
	for scanner.Scan() {
		ln := scanner.Text()
		if re.MatchString(ln) {
			// skip
		} else if !(strings.HasPrefix(ln, "+") || strings.HasPrefix(ln, "-")) {
			// skip
		} else {
			fmt.Printf("CHANGED: %s\n", ln)
			changed = true
		}
	}
	if !changed {
		out, err := exec.Command("git", "--no-pager", "restore", "--source=HEAD", "--staged", "--worktree", "--", file).CombinedOutput()
		if err != nil {
			log.Fatalf("Restoring file %s: %v\n%s", file, err, out)
		}
		scanner := bufio.NewScanner(bytes.NewBuffer(out))
		for scanner.Scan() {
			ln := scanner.Text()
			fmt.Println(ln)
		}
	}
}
