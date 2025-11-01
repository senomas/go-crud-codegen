#!/usr/bin/env python3
import sys

# usage: python keep_last.py N
# reads lines like: "docker.mycom.com/example:tag 2025-10-21 15:20:12 +0000 UTC"

if len(sys.argv) < 2:
    print("Usage: keep_last.py <N>", file=sys.stderr)
    sys.exit(1)

keep = int(sys.argv[1])
lines = [line.strip() for line in sys.stdin if line.strip()]

# sort newest first (since docker images output is not always sorted)
lines.sort(key=lambda l: " ".join(l.split()[1:]), reverse=True)

for line in lines[keep:]:
    tag = line.split()[0]
    print(tag)
