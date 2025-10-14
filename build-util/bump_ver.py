#!/usr/bin/env python3
import sys, re


def bump_last_number(ver: str) -> str:
    """
    Increment the last numeric run in a version string.
    Examples:
      1.0.8  -> 1.0.9
      v2.9   -> v2.10
      release-99 -> release-100
    """
    m = re.search(r"(.*?)(\d+)(\D*)$", ver)
    if not m:
        return ver + "1"
    prefix, num, tail = m.groups()
    return f"{prefix}{int(num) + 1}{tail}"


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: bump_version.py <version>", file=sys.stderr)
        sys.exit(1)
    old = sys.argv[1]
    print(bump_last_number(old))
