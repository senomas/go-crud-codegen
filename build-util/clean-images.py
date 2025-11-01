#!/usr/bin/env python3
import argparse
import docker
from datetime import datetime


def main():
    parser = argparse.ArgumentParser(
        description="Delete all Docker images matching <repository> except the most recent ones."
    )
    parser.add_argument(
        "repository", help="Repository name (e.g., docker.mycom.com/example)"
    )
    parser.add_argument(
        "--keep",
        type=int,
        default=1,
        help="Number of most recent images to keep (default: 1)",
    )
    parser.add_argument(
        "--force",
        action="store_true",
        help="Force remove images even if used by containers",
    )
    args = parser.parse_args()

    client = docker.from_env()
    repo = args.repository

    print(f"🔍 Searching images for repository: {repo}")
    images = [
        img
        for img in client.images.list()
        if any(tag.startswith(repo) for tag in img.tags)
    ]

    if not images:
        print("No matching images found.")
        return

    # Sort by creation time (newest first)
    images.sort(key=lambda img: img.attrs["Created"], reverse=True)

    print(f"Keeping {args.keep} most recent image(s):")
    for img in images[: args.keep]:
        print(f"  🟢 {img.tags}")

    # Delete the rest
    for img in images[args.keep :]:
        for tag in img.tags:
            if tag.startswith(repo):
                try:
                    print(f"  🗑️ Removing {tag}...")
                    client.images.remove(tag, force=args.force)
                except Exception as e:
                    print(f"  ❌ Failed to remove {tag}: {e}")

    print("✅ Cleanup complete.")


if __name__ == "__main__":
    main()
