#!/bin/sh
set -e

REPO_URL="https://github.com/UreaLaden/log-sage"
TMP_ROOT="${TMPDIR:-/tmp}"
DRY_RUN=0
OS=""
ARCH=""
os=""
arch=""
version=""
filename=""
base_url=""
artifact_url=""
artifact_tmp=""
checksums_tmp=""
check_tmp=""
extract_dir=""
dest_dir=""
dest_path=""

usage() {
	cat <<'EOF'
Usage: sh install.sh [--dry-run] [--help]

Downloads the correct LogSage release artifact for the current macOS or Linux host,
verifies the SHA256 checksum, and installs the logsage binary.

Options:
  --dry-run  Print the resolved OS, arch, version, URL, and install path without downloading
  -h, --help Show this help text

Environment:
  VERSION    Release version to install (for example 1.0.0). If unset, the script
             resolves the latest GitHub release at runtime unless --dry-run is used.
EOF
}

say() {
	printf '%s\n' "$*"
}

fail() {
	printf '%s\n' "$*" >&2
	exit 1
}

cleanup() {
	rm -f "$artifact_tmp" "$checksums_tmp" "$check_tmp"
	rm -rf "$extract_dir"
}

run_or_echo() {
	if [ "$DRY_RUN" -eq 1 ]; then
		printf 'DRY-RUN:'
		for arg in "$@"; do
			printf ' %s' "$arg"
		done
		printf '\n'
		return 0
	fi

	"$@"
}

detect_os() {
	OS=$(uname -s)
	case "$OS" in
		Linux) os="linux" ;;
		Darwin) os="darwin" ;;
		*) fail "Unsupported OS: $OS" ;;
	esac
}

detect_arch() {
	ARCH=$(uname -m)
	case "$ARCH" in
		x86_64) arch="amd64" ;;
		arm64|aarch64) arch="arm64" ;;
		*) fail "Unsupported architecture: $ARCH" ;;
	esac
}

resolve_version() {
	if [ -n "${VERSION:-}" ]; then
		version=$VERSION
		return
	fi

	if [ "$DRY_RUN" -eq 1 ]; then
		version="latest"
		return
	fi

	say "Resolving latest release version..."
	effective_url=$(curl -fsSLI -o /dev/null -w '%{url_effective}' "${REPO_URL}/releases/latest") ||
		fail "Unable to resolve the latest release version. Set VERSION=... and retry."
	version=$(printf '%s' "$effective_url" | sed 's#.*/tag/v##')
	[ -n "$version" ] || fail "Unable to parse the latest release version from ${effective_url}"
}

resolve_paths() {
	filename="logsage_${version}_${os}_${arch}.tar.gz"
	base_url="${REPO_URL}/releases/download/v${version}"
	artifact_url="${base_url}/${filename}"
	artifact_tmp="${TMP_ROOT%/}/${filename}"
	checksums_tmp="${TMP_ROOT%/}/checksums.txt"
	check_tmp="${TMP_ROOT%/}/logsage-checksums-${os}-${arch}.txt"
	extract_dir="${TMP_ROOT%/}/logsage-install-${os}-${arch}.$$"

	if [ "$os" = "darwin" ]; then
		dest_dir="${HOME}/bin"
	else
		if [ -w /usr/local/bin ]; then
			dest_dir="/usr/local/bin"
		elif command -v sudo >/dev/null 2>&1; then
			dest_dir="/usr/local/bin"
		else
			dest_dir="${HOME}/.local/bin"
		fi
	fi

	dest_path="${dest_dir}/logsage"
}

print_plan() {
	say "OS: ${os}"
	say "Arch: ${arch}"
	say "Version: ${version}"
	say "Artifact: ${filename}"
	say "URL: ${artifact_url}"
	say "Install path: ${dest_path}"
}

download_artifacts() {
	say "Downloading ${filename}..."
	run_or_echo curl -fsSL -o "$artifact_tmp" "$artifact_url"

	say "Downloading checksums.txt..."
	run_or_echo curl -fsSL -o "$checksums_tmp" "${base_url}/checksums.txt"
}

verify_checksum() {
	say "Verifying checksum..."

	if [ "$DRY_RUN" -eq 1 ]; then
		say "DRY-RUN: checksum verification skipped"
		return
	fi

	grep "  ${filename}\$" "$checksums_tmp" >"$check_tmp" ||
		fail "Artifact ${filename} not found in checksums.txt"

	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum -c "$check_tmp"
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 -c "$check_tmp"
	else
		fail "No SHA256 verifier found. Install sha256sum or shasum and retry."
	fi
}

extract_artifact() {
	say "Extracting ${filename}..."
	run_or_echo mkdir -p "$extract_dir"
	run_or_echo tar -xzf "$artifact_tmp" -C "$extract_dir"
}

install_binary() {
	say "Installing logsage to ${dest_path}..."

	if [ "$os" = "linux" ] && [ "$dest_dir" = "/usr/local/bin" ] && [ ! -w /usr/local/bin ]; then
		run_or_echo sudo mkdir -p "$dest_dir"
		run_or_echo sudo cp "${extract_dir}/logsage" "$dest_path"
		run_or_echo sudo chmod +x "$dest_path"
		return
	fi

	run_or_echo mkdir -p "$dest_dir"
	run_or_echo cp "${extract_dir}/logsage" "$dest_path"
	run_or_echo chmod +x "$dest_path"
}

verify_install() {
	if [ "$DRY_RUN" -eq 1 ]; then
		say "DRY-RUN: install verification skipped"
		return
	fi

	say "Installed to ${dest_path}"
	say "Running version check..."
	"$dest_path" version
}

case "${1:-}" in
	--dry-run)
		DRY_RUN=1
		;;
	-h|--help)
		usage
		exit 0
		;;
	"")
		;;
	*)
		fail "Unknown argument: $1"
		;;
esac

trap cleanup EXIT INT TERM

say "Detecting platform..."
detect_os
detect_arch
resolve_version
resolve_paths
print_plan

download_artifacts
verify_checksum
extract_artifact
install_binary
verify_install
