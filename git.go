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

func diff(file string) {
	if _, err := os.Stat(file); err == nil {
		// continue
	} else if os.IsNotExist(err) {
		return
	} else {
		log.Fatalf("Error checking %s: %v\n", file, err)
	}
	out, err := exec.Command("pwd").CombinedOutput()
	if err != nil {
		log.Fatalf("Checking pwd: %v\n\n%s",
			err, string(out))
	}
	fmt.Printf("Checking git diff for %s in %s\n", file, strings.TrimSpace(string(out)))
	out, err = exec.Command("git", "--no-pager", "diff", "--", file).CombinedOutput()
	if err != nil {
		log.Fatalf("Checking git diff for %s: %v\n\n%s", file, err, string(out))
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
			estr := string(out)
			if strings.Contains(estr, "did not match any file") {
				// ignore
			} else {
				log.Fatalf("Restoring file %s: %v\n%s", file, err, estr)
			}
		}
		scanner := bufio.NewScanner(bytes.NewBuffer(out))
		if scanner.Scan() {
			ln := scanner.Text()
			if strings.HasPrefix(ln, "error: pathspec") {
				return
			}
			fmt.Println(ln)
			for scanner.Scan() {
				ln := scanner.Text()
				fmt.Println(ln)
			}
		}
	}
}
