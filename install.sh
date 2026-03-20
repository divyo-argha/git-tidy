#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────────────
#  git-tidy installer
#
#  Usage:
#    curl -fsSL https://raw.githubusercontent.com/divyo-argha/git-tidy/main/install.sh | bash
#    curl -fsSL https://raw.githubusercontent.com/divyo-argha/git-tidy/main/install.sh | bash -s -- --version v1.2.0
#    curl -fsSL https://raw.githubusercontent.com/divyo-argha/git-tidy/main/install.sh | bash -s -- --dir ~/.local/bin
#
#  Supports: Linux (amd64, arm64), macOS (amd64, arm64 / Apple Silicon)
# ─────────────────────────────────────────────────────────────────────────────

set -euo pipefail

# ── Config ────────────────────────────────────────────────────────────────────

REPO="divyo-argha/git-tidy"
BINARY="git-tidy"
DEFAULT_INSTALL_DIR="/usr/local/bin"
GITHUB_API="https://api.github.com/repos/${REPO}"
GITHUB_RELEASES="https://github.com/${REPO}/releases/download"

# ── Colours ───────────────────────────────────────────────────────────────────

if [ -t 1 ] && [ "${NO_COLOR:-}" = "" ]; then
  RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[0;33m'
  CYAN='\033[0;36m'; BOLD='\033[1m'; RESET='\033[0m'
else
  RED=''; GREEN=''; YELLOW=''; CYAN=''; BOLD=''; RESET=''
fi

info()    { echo -e "  ${CYAN}→${RESET}  $*" >&2; }
success() { echo -e "  ${GREEN}✓${RESET}  $*" >&2; }
warn()    { echo -e "  ${YELLOW}!${RESET}  $*" >&2; }
die()     { echo -e "  ${RED}✗${RESET}  ${BOLD}Error:${RESET} $*" >&2; exit 1; }
header()  { echo -e "\n  ${BOLD}${CYAN}git-tidy installer${RESET}\n" >&2; }

# ── Argument parsing ──────────────────────────────────────────────────────────

REQUESTED_VERSION=""
INSTALL_DIR=""
DRY_RUN=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version|-V) REQUESTED_VERSION="${2:-}"; shift 2 ;;
    --dir|-d)     INSTALL_DIR="${2:-}";       shift 2 ;;
    --dry-run)    DRY_RUN=true;               shift   ;;
    --help|-h)
      echo "Usage: install.sh [--version <tag>] [--dir <path>] [--dry-run]"
      exit 0 ;;
    *) die "Unknown argument: $1" ;;
  esac
done

INSTALL_DIR="${INSTALL_DIR:-$DEFAULT_INSTALL_DIR}"

# ── Platform detection ────────────────────────────────────────────────────────

detect_os() {
  case "$(uname -s)" in
    Linux*)  echo "linux"  ;;
    Darwin*) echo "darwin" ;;
    *)       die "Unsupported OS: $(uname -s). Install manually from https://github.com/${REPO}/releases" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *) die "Unsupported architecture: $(uname -m). Install manually from https://github.com/${REPO}/releases" ;;
  esac
}

# ── Dependency checks ─────────────────────────────────────────────────────────

check_deps() {
  local missing=()
  for cmd in curl git; do
    command -v "$cmd" &>/dev/null || missing+=("$cmd")
  done
  if [ ${#missing[@]} -gt 0 ]; then
    die "Missing required tools: ${missing[*]}"
  fi
}

# ── Version resolution ────────────────────────────────────────────────────────

resolve_version() {
  if [ -n "$REQUESTED_VERSION" ]; then
    echo "$REQUESTED_VERSION"
    return
  fi

  info "Fetching latest release…"
  local tag
  tag=$(curl -fsSL "${GITHUB_API}/releases/latest" \
    | grep '"tag_name"' \
    | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/') \
    || die "Could not fetch latest release. Check your internet connection."

  if [ -z "$tag" ]; then
    die "Could not determine latest version. Specify one with --version."
  fi
  echo "$tag"
}

# ── Download & install ────────────────────────────────────────────────────────

install_binary() {
  local tag="$1"
  local os="$2"
  local arch="$3"

  # ── Binary Download ─────────────────────────────────────────────────────────
  # Asset name must match what goreleaser produces.
  local asset="${BINARY}-${os}-${arch}"
  local url="${GITHUB_RELEASES}/${tag}/${asset}"
  local checksum_url="${GITHUB_RELEASES}/${tag}/checksums.txt"

  info "Version  : ${tag}"
  info "Platform : ${os}/${arch}"
  info "Target   : ${INSTALL_DIR}/${BINARY}"

  if $DRY_RUN; then
    warn "Dry-run mode — nothing will be downloaded or installed."
    info "Would download: ${url}"
    return
  fi

  # Create a temp dir that is cleaned up on exit.
  local tmp
  tmp=$(mktemp -d)
  trap 'rm -rf "$tmp"' EXIT

  info "Downloading binary…"
  curl -fsSL --progress-bar "$url" -o "${tmp}/${BINARY}" \
    || die "Download failed. Does release ${tag} exist?\n     Check: https://github.com/${REPO}/releases"

  # Verify checksum if available (non-fatal — warns instead of blocking).
  info "Verifying checksum…"
  if curl -fsSL "$checksum_url" -o "${tmp}/checksums.txt" 2>/dev/null; then
    local expected actual
    # Match the asset name at the end of the line, allowing for optional '*' or './' prefixes
    # and handling potential carriage returns (\r) from CRLF line endings.
    expected=$(sed 's/\r$//' "${tmp}/checksums.txt" | grep -E "[[:space:]\*./]*${asset}$" | awk '{print $1}')
    if [ -n "$expected" ]; then
      if command -v sha256sum &>/dev/null; then
        actual=$(sha256sum "${tmp}/${BINARY}" | awk '{print $1}')
      elif command -v shasum &>/dev/null; then
        actual=$(shasum -a 256 "${tmp}/${BINARY}" | awk '{print $1}')
      fi
      if [ -n "${actual:-}" ] && [ "$actual" != "$expected" ]; then
        die "Checksum mismatch!\n     Expected: ${expected}\n     Got:      ${actual}"
      fi
      success "Checksum verified"
    else
      warn "No checksum entry found for ${asset} — skipping verification"
    fi
  else
    warn "Checksums file not available — skipping verification"
  fi

  chmod +x "${tmp}/${BINARY}"

  # Install — try direct copy first, fall back to sudo.
  info "Installing to ${INSTALL_DIR}…"
  if [ -w "$INSTALL_DIR" ]; then
    cp "${tmp}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
  elif command -v sudo &>/dev/null; then
    warn "${INSTALL_DIR} requires sudo"
    sudo install -m 0755 "${tmp}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
  else
    die "Cannot write to ${INSTALL_DIR} and sudo is not available.\nRun with --dir to choose a writable directory:\n  install.sh --dir ~/.local/bin"
  fi
}

# ── Post-install checks ───────────────────────────────────────────────────────

verify_install() {
  local installed_path="${INSTALL_DIR}/${BINARY}"

  if [ ! -x "$installed_path" ]; then
    die "Binary not found at ${installed_path} after install."
  fi

  # Check that INSTALL_DIR is on PATH.
  if ! echo "$PATH" | tr ':' '\n' | grep -qx "$INSTALL_DIR"; then
    warn "${INSTALL_DIR} is not on your PATH."
    echo ""
    echo "  Add it to your shell profile:"
    echo "    ${BOLD}echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.bashrc${RESET}   # bash"
    echo "    ${BOLD}echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.zshrc${RESET}    # zsh"
    echo ""
  fi

  # Confirm the binary runs.
  local reported_version
  reported_version=$("${installed_path}" version --short 2>/dev/null || echo "unknown")
  success "Installed ${BINARY} ${reported_version}"
}

# ── Main ──────────────────────────────────────────────────────────────────────

main() {
  header
  check_deps

  local os arch version
  os=$(detect_os)
  arch=$(detect_arch)
  version=$(resolve_version)

  install_binary "$version" "$os" "$arch"

  if ! $DRY_RUN; then
    verify_install
    echo ""
    success "git-tidy is ready. Try it:"
    echo ""
    echo "    ${BOLD}git tidy --help${RESET}"
    echo "    ${BOLD}git tidy preview${RESET}"
    echo ""
  fi
}

main "$@"
