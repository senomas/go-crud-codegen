#!/usr/bin/env python3
import sys, os, tempfile, shutil
from pathlib import Path

USAGE = """\
Usage:
  prop.py get <file> <KEY>
  prop.py set <file> <KEY> <VALUE>
  prop.py set-many <file> KEY=VALUE [KEY=VALUE ...]
  prop.py build-args <file>
Notes:
  - Matches keys as ^KEY= (no leading spaces).
  - If key not found, appends 'KEY=VALUE' at end.
  - Preserves unrelated lines/comments.
"""


def read_lines(p: Path):
    if not p.exists():
        return []
    txt = p.read_text(encoding="utf-8", errors="ignore").splitlines()
    return txt


def write_atomic(p: Path, lines):
    tmp = Path(str(p) + ".tmp")
    tmp.write_text("\n".join(lines) + "\n", encoding="utf-8")
    tmp.replace(p)


def normalize_line_endings(lines):
    # Already splitlines(); just strip trailing \r on each line.
    return [ln.rstrip("\r") for ln in lines]


def get_key(lines, key):
    prefix = f"{key}="
    for ln in lines:
        if ln.startswith(prefix):
            return ln[len(prefix) :]
    return ""


def set_key(lines, key, value):
    """Return updated lines; replace first ^KEY=... or append new line."""
    prefix = f"{key}="
    replaced = False
    out = []
    for ln in lines:
        if not replaced and ln.startswith(prefix):
            out.append(f"{key}={value}")
            replaced = True
        else:
            out.append(ln)
    if not replaced:
        out.append(f"{key}={value}")
    return out


def main(argv):
    if len(argv) < 2:
        print(USAGE, file=sys.stderr)
        return 2

    cmd = argv[1]
    fpath = Path(argv[2])
    lines = normalize_line_endings(read_lines(fpath))

    if cmd == "get":
        if len(argv) != 4:
            print(USAGE, file=sys.stderr)
            return 2
        key = argv[3]
        val = get_key(lines, key)
        print(val)
        return 0

    if cmd == "set":
        if len(argv) != 5:
            print(USAGE, file=sys.stderr)
            return 2
        key, val = argv[3], argv[4]
        new_lines = set_key(lines, key, val)
        write_atomic(fpath, new_lines)
        return 0

    if cmd == "set-many":
        if len(argv) < 4:
            print(USAGE, file=sys.stderr)
            return 2
        new_lines = lines[:]
        for pair in argv[3:]:
            if "=" not in pair:
                print(f"Invalid pair (expect KEY=VALUE): {pair}", file=sys.stderr)
                return 2
            k, v = pair.split("=", 1)
            new_lines = set_key(new_lines, k, v)
        write_atomic(fpath, new_lines)
        return 0

    if cmd == "build-args":
        with open(fpath) as f:
            for line in f:
                line = line.strip()
                if line and not line.startswith("#"):
                    key, value = line.split("=", 1)
                    print(f"--build-arg {key}={value}", end=" ")
        return 0

    if cmd == "envs":
        with open(fpath) as f:
            for line in f:
                line = line.strip()
                if line and not line.startswith("#"):
                    key, value = line.split("=", 1)
                    print(f"-e {key}={value}", end=" ")
        return 0

    print(f"Unknown command: {argv}", file=sys.stderr)
    print(USAGE, file=sys.stderr)
    return 2


if __name__ == "__main__":
    sys.exit(main(sys.argv))
