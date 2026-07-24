#!/bin/sh
set -eu

source_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
source_app="$source_dir/Space Rocks.app"
install_root=${1:-${SPACE_ROCKS_INSTALL_DIR:-"$HOME/Applications"}}

if [ ! -d "$source_app" ]; then
    echo "Space Rocks.app was not found beside install.command. Run the installer from an extracted Space Rocks release package." >&2
    exit 1
fi
if [ ! -x "$source_app/Contents/MacOS/SpaceRocks" ] && [ ! -x "$source_app/Contents/MacOS/Space Rocks" ]; then
    echo "The Space Rocks application bundle is incomplete." >&2
    exit 1
fi

mkdir -p "$install_root"
install_root=$(CDPATH= cd -- "$install_root" && pwd)
target_app="$install_root/Space Rocks.app"

if [ "$source_app" != "$target_app" ]; then
    case "$target_app/" in
        "$source_app"/*)
            echo "The installation directory cannot be inside the extracted release package." >&2
            exit 1
            ;;
    esac

    if [ -e "$target_app" ]; then
        backup_app="$install_root/Space Rocks.app.backup-$(date +%Y%m%d-%H%M%S)-$$"
        mv "$target_app" "$backup_app"
        echo "Previous installation moved to $backup_app"
    fi

    ditto "$source_app" "$target_app"
fi

chmod +x "$target_app/Contents/MacOS/"* 2>/dev/null || true

echo "Space Rocks installed to $target_app"
echo "Open it from Finder or run: open \"$target_app\""
