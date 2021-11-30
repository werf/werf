#!/usr/bin/env bash
set -uo pipefail

main() {
  declare -r REQUIRED_BASH_VERSION="3.2.57"
  declare -r REQUIRED_GIT_VERSION="2.18.0"

  declare -r DEFAULT_TRDL_INITIAL_VERSION="0.3.1"
  declare -r DEFAULT_TRDL_TUF_REPO="https://tuf.trdl.dev"
  declare -r DEFAULT_TRDL_GPG_PUBKEY_URL="https://trdl.dev/trdl.asc"
  declare -r DEFAULT_WERF_TUF_REPO_URL="https://tuf.werf.io"
  declare -r DEFAULT_WERF_TUF_REPO_ROOT_VERSION="1"
  declare -r DEFAULT_WERF_TUF_REPO_ROOT_SHA="b7ff6bcbe598e072a86d595a3621924c8612c7e6dc6a82e919abe89707d7e3f468e616b5635630680dd1e98fc362ae5051728406700e6274c5ed1ad92bea52a2"
  declare -r DEFAULT_WERF_AUTOACTIVATE_VERSION="1.2"
  declare -r DEFAULT_WERF_AUTOACTIVATE_CHANNEL="ea"

  declare -r NON_INTERACTIVE_VALID_REGEX='^(yes|no)$'
  declare -r CI_VALID_REGEX='^(yes|no)$'
  declare -r TRDL_SKIP_SIGNATURE_VERIFICATION_VALID_REGEX='^(yes|no)$'
  declare -r SHELL_VALID_REGEX='^(bash|zsh|auto)?$'
  declare -r WERF_JOIN_DOCKER_GROUP_VALID_REGEX='^(yes|no|auto)$'
  declare -r TRDL_ENABLE_SETUP_BIN_PATH_VALID_REGEX='^(yes|no|auto)$'
  declare -r WERF_ENABLE_AUTOACTIVATION_VALID_REGEX='^(yes|no|auto)$'
  declare -r TRDL_INITIAL_VERSION_VALID_REGEX='^[0-9]+.[0-9]+.[0-9]+$'
  declare -r TRDL_TUF_REPO_VALID_REGEX='^https?://.+'
  declare -r TRDL_GPG_PUBKEY_URL_VALID_REGEX='^https?://.+'
  declare -r WERF_TUF_REPO_URL_VALID_REGEX='^https?://.+'
  declare -r WERF_TUF_REPO_ROOT_VERSION_VALID_REGEX='^[0-9]+$'
  declare -r WERF_TUF_REPO_ROOT_SHA_VALID_REGEX='^[a-fA-F0-9]+$'
  declare -r WERF_AUTOACTIVATE_VERSION_VALID_REGEX='^[.0-9]+$'
  declare -r WERF_AUTOACTIVATE_CHANNEL_VALID_REGEX='.+'

  validate_bash_version "$REQUIRED_BASH_VERSION"

  declare NON_INTERACTIVE="${WI_NON_INTERACTIVE:-"no"}"
  declare ci="${WI_CI:-"no"}"
  declare trdl_skip_signature_verification="${WI_TRDL_SKIP_SIGNATURE_VERIFICATION:-"no"}"
  declare shell="${WI_SHELL:-"auto"}"
  declare werf_join_docker_group="${WI_WERF_JOIN_DOCKER_GROUP:-"auto"}"
  declare trdl_enable_setup_bin_path="${WI_TRDL_ENABLE_SETUP_BIN_PATH:-"auto"}"
  declare werf_enable_autoactivation="${WI_WERF_ENABLE_AUTOACTIVATION:-"auto"}"
  declare trdl_initial_version="${WI_TRDL_INITIAL_VERSION:-$DEFAULT_TRDL_INITIAL_VERSION}"
  declare trdl_tuf_repo="${WI_TRDL_TUF_REPO:-$DEFAULT_TRDL_TUF_REPO}"
  declare trdl_gpg_pubkey_url="${WI_TRDL_GPG_PUBKEY_URL:-$DEFAULT_TRDL_GPG_PUBKEY_URL}"
  declare werf_tuf_repo_url="${WI_WERF_TUF_REPO_URL:-$DEFAULT_WERF_TUF_REPO_URL}"
  declare werf_tuf_repo_root_version="${WI_WERF_TUF_REPO_ROOT_VERSION:-$DEFAULT_WERF_TUF_REPO_ROOT_VERSION}"
  declare werf_tuf_repo_root_sha="${WI_WERF_TUF_REPO_ROOT_SHA:-$DEFAULT_WERF_TUF_REPO_ROOT_SHA}"
  declare werf_autoactivate_version="${WI_WERF_AUTOACTIVATE_VERSION:-$DEFAULT_WERF_AUTOACTIVATE_VERSION}"
  declare werf_autoactivate_channel="${WI_WERF_AUTOACTIVATE_CHANNEL:-$DEFAULT_WERF_AUTOACTIVATE_CHANNEL}"

  declare OPTIND opt
  while getopts ":qxgi:t:p:w:r:y:v:c:a:d:b:s:h" opt; do
    case "$opt" in
      q) NON_INTERACTIVE="yes" ;;
      x) ci="yes" ;;
      g) trdl_skip_signature_verification="yes" ;;
      s) shell="$OPTARG" ;;
      d) werf_join_docker_group="$OPTARG" ;;
      b) trdl_enable_setup_bin_path="$OPTARG" ;;
      a) werf_enable_autoactivation="$OPTARG" ;;
      i) trdl_initial_version="$OPTARG" ;;
      t) trdl_tuf_repo="$OPTARG" ;;
      p) trdl_gpg_pubkey_url="$OPTARG" ;;
      w) werf_tuf_repo_url="$OPTARG" ;;
      r) werf_tuf_repo_root_version="$OPTARG" ;;
      y) werf_tuf_repo_root_sha="$OPTARG" ;;
      v) werf_autoactivate_version="$OPTARG" ;;
      c) werf_autoactivate_channel="$OPTARG" ;;
      h) usage && exit 0 ;;
      *) printf 'Unknown option: -%s\n\n' "$OPTARG" && usage && exit 1 ;;
    esac
  done

  if [[ $ci == "yes" ]]; then
    NON_INTERACTIVE="yes"
    werf_join_docker_group="no"
    trdl_enable_setup_bin_path="no"
    werf_enable_autoactivation="no"
  fi

  validate_option_by_regex "$NON_INTERACTIVE" "non-interactive mode" "$NON_INTERACTIVE_VALID_REGEX"
  validate_option_by_regex "$ci" "CI mode" "$CI_VALID_REGEX"
  validate_option_by_regex "$trdl_skip_signature_verification" "trdl skip signature verification" "$TRDL_SKIP_SIGNATURE_VERIFICATION_VALID_REGEX"
  validate_option_by_regex "$shell" "shell" "$SHELL_VALID_REGEX"
  validate_option_by_regex "$werf_join_docker_group" "join Docker group" "$WERF_JOIN_DOCKER_GROUP_VALID_REGEX"
  validate_option_by_regex "$trdl_enable_setup_bin_path" "setup trdl bin path" "$TRDL_ENABLE_SETUP_BIN_PATH_VALID_REGEX"
  validate_option_by_regex "$werf_enable_autoactivation" "enable werf autoactivation" "$WERF_ENABLE_AUTOACTIVATION_VALID_REGEX"
  validate_option_by_regex "$trdl_initial_version" "initial trdl version" "$TRDL_INITIAL_VERSION_VALID_REGEX"
  validate_option_by_regex "$trdl_tuf_repo" "trdl tuf repo" "$TRDL_TUF_REPO_VALID_REGEX"
  validate_option_by_regex "$trdl_gpg_pubkey_url" "trdl GPG public key url" "$TRDL_GPG_PUBKEY_URL_VALID_REGEX"
  validate_option_by_regex "$werf_tuf_repo_url" "werf TUF-repo URL" "$WERF_TUF_REPO_URL_VALID_REGEX"
  validate_option_by_regex "$werf_tuf_repo_root_version" "werf TUF-repo root version" "$WERF_TUF_REPO_ROOT_VERSION_VALID_REGEX"
  validate_option_by_regex "$werf_tuf_repo_root_sha" "werf TUF-repo root SHA" "$WERF_TUF_REPO_ROOT_SHA_VALID_REGEX"
  validate_option_by_regex "$werf_autoactivate_version" "werf autoactivation version" "$WERF_AUTOACTIVATE_VERSION_VALID_REGEX"
  validate_option_by_regex "$werf_autoactivate_channel" "werf autoactivation channel" "$WERF_AUTOACTIVATE_CHANNEL_VALID_REGEX"

  ensure_cmds_available uname docker git grep tee install
  [[ $trdl_skip_signature_verification == "no" ]] && ensure_cmds_available gpg
  validate_git_version "$REQUIRED_GIT_VERSION"

  declare arch
  arch="$(get_arch)" || abort "Failure getting system architecture."

  declare os
  os="$(get_os)" || abort "Failure getting OS."

  [[ $shell == "auto" ]] && get_shell "shell"

  # declare linux_distro
  # if [[ $os == "linux" ]]; then
  #   linux_distro="$(get_linux_distro)" || abort "Failure getting Linux distro."
  # fi

  # declare darwin_version
  # if [[ $os == "darwin" ]]; then
  #   darwin_version="$(sw_vers -productVersion)" || abort "Failure getting macOS version."
  # fi

  [[ $os == "linux" ]] && propose_joining_docker_group "$werf_join_docker_group"
  setup_trdl_bin_path "$shell" "$trdl_enable_setup_bin_path"
  install_trdl "$os" "$arch" "$trdl_tuf_repo" "$trdl_gpg_pubkey_url" "$trdl_initial_version" "$trdl_skip_signature_verification"
  add_trdl_werf_repo "$werf_tuf_repo_url" "$werf_tuf_repo_root_version" "$werf_tuf_repo_root_sha"
  enable_automatic_werf_activation "$shell" "$werf_enable_autoactivation" "$werf_autoactivate_version" "$werf_autoactivate_channel"
  finalize "$werf_autoactivate_version" "$werf_autoactivate_channel"
}

usage() {
  printf 'Usage: %s [options]

Options:
  -x          CI mode. Configures a few other options in a way recommended for CI (i.e. enables non-interactive mode). Full list of configured options can be found in the installer script itself. Default: %s. Environment variable: $WI_CI.
  -q          Run non-interactively, choosing recommended answers for all prompts. Default: %s. Allowed values regex: /%s/ Environment variable: $WI_NON_INTERACTIVE.
  -g          Skip signature verification of trdl binary. Default: %s. Allowed values regex: /%s/ Environment variable: $WI_TRDL_SKIP_SIGNATURE_VERIFICATION.
  -s shell    Shell, for which werf/trdl should be set up. Default: %s. Allowed values regex: /%s/ Environment variable: $WI_SHELL.
  -d value    Add user to the docker group? Default: %s. Allowed values regex: /%s/ Environment variable: $WI_WERF_JOIN_DOCKER_GROUP.
  -b value    Add "$HOME/bin" to $PATH for trdl? Default: %s. Allowed values regex: /%s/ Environment variable: $WI_TRDL_ENABLE_SETUP_BIN_PATH.
  -a value    Enable werf autoactivation? Default: %s. Allowed values regex: /%s/ Environment variable: $WI_WERF_ENABLE_AUTOACTIVATION.
  -i version  Initially installed version of trdl (self-updates automatically). Default: %s. Allowed values regex: /%s/ Environment variable: $WI_TRDL_INITIAL_VERSION.
  -t url      trdl TUF-repository URL. Default: %s. Allowed values regex: /%s/ Environment variable: $WI_TRDL_TUF_REPO.
  -p url      trdl GPG public key URL. Default: %s. Allowed values regex: /%s/ Environment variable: $WI_TRDL_GPG_PUBKEY_URL.
  -w url      werf TUF-repository URL. Default: %s. Allowed values regex: /%s/ Environment variable: $WI_WERF_TUF_REPO_URL.
  -r version  werf TUF-repository root version. Default: %s. Allowed values regex: /%s/ Environment variable: $WI_WERF_TUF_REPO_ROOT_VERSION.
  -y hash     werf TUF-repository root SHA-hash. Default: %s. Allowed values regex: /%s/ Environment variable: $WI_WERF_TUF_REPO_ROOT_SHA.
  -v version  Autoactivated (if enabled) werf version. Default: %s. Allowed values regex: /%s/ Environment variable: $WI_WERF_AUTOACTIVATE_VERSION.
  -c channel  Autoactivated (if enabled) werf channel. Default: %s. Allowed values regex: /%s/ Environment variable: $WI_WERF_AUTOACTIVATE_CHANNEL.
  -h          Show help.

Example:
  %s -qs bash\n' "$(basename "$0")" \
    "no" \
    "no" "$NON_INTERACTIVE_VALID_REGEX" \
    "no" "$TRDL_SKIP_SIGNATURE_VERIFICATION_VALID_REGEX" \
    "auto" "$SHELL_VALID_REGEX" \
    "auto" "$WERF_JOIN_DOCKER_GROUP_VALID_REGEX" \
    "auto" "$TRDL_ENABLE_SETUP_BIN_PATH_VALID_REGEX" \
    "auto" "$WERF_ENABLE_AUTOACTIVATION_VALID_REGEX" \
    "$DEFAULT_TRDL_INITIAL_VERSION" "$TRDL_INITIAL_VERSION_VALID_REGEX" \
    "$DEFAULT_TRDL_TUF_REPO" "$TRDL_TUF_REPO_VALID_REGEX" \
    "$DEFAULT_TRDL_GPG_PUBKEY_URL" "$TRDL_GPG_PUBKEY_URL_VALID_REGEX" \
    "$DEFAULT_WERF_TUF_REPO_URL" "$WERF_TUF_REPO_URL_VALID_REGEX" \
    "$DEFAULT_WERF_TUF_REPO_ROOT_VERSION" "$WERF_TUF_REPO_ROOT_VERSION_VALID_REGEX" \
    "$DEFAULT_WERF_TUF_REPO_ROOT_SHA" "$WERF_TUF_REPO_ROOT_SHA_VALID_REGEX" \
    "$DEFAULT_WERF_AUTOACTIVATE_VERSION" "$WERF_AUTOACTIVATE_VERSION_VALID_REGEX" \
    "$DEFAULT_WERF_AUTOACTIVATE_CHANNEL" "$WERF_AUTOACTIVATE_CHANNEL_VALID_REGEX" \
    "$(basename "$0")"
}

validate_option_by_regex() {
  declare value="$1"
  declare human_option_name="$2"
  declare valid_regex="$3"

  [[ $value =~ $valid_regex ]] || abort "Invalid $human_option_name passed: $value\nMust match regex: $valid_regex"
}

validate_bash_version() {
  declare required_bash_version="$1"

  declare current_bash_version="${BASH_VERSINFO[0]}.${BASH_VERSINFO[1]}.${BASH_VERSINFO[2]}"
  compare_versions "$current_bash_version" "$required_bash_version"
  [[ $? -gt 1 ]] && abort "Bash version must be at least \"$required_bash_version\"."
}

validate_git_version() {
  declare required_git_version="$1"

  declare current_git_version
  current_git_version="$(git --version | awk '{print $3}')" || abort "Unable to parse git version."
  compare_versions "$current_git_version" "$required_git_version"
  [[ $? -gt 1 ]] && abort "Git version must be at least \"$required_git_version\"."
}

get_arch() {
  declare arch
  arch="$(uname -m)" || abort "Can't get system architecture."

  case "$arch" in
    x86_64) arch="amd64" ;;
    arm64 | armv8* | aarch64*) arch="arm64" ;;
    i386 | i486 | i586 | i686) abort "werf is not available for x86 architecture." ;;
    arm | armv7*) abort "werf is not available for 32-bit ARM architectures." ;;
    *) abort "werf is not available for \"$arch\" architecture." ;;
  esac

  printf '%s' "$arch"
}

get_os() {
  declare os
  os="$(uname | tr '[:upper:]' '[:lower:]')" || abort "Can't detect OS."

  case "$os" in
    linux*) os="linux" ;;
    darwin*) os="darwin" ;;
    msys* | cygwin* | mingw* | nt | win*) abort "This installer does not support Windows." ;;
    *) abort "Unknown OS." ;;
  esac

  printf '%s' "$os"
}

get_shell() {
  declare result_var_name="$1"

  declare result
  result="$(printf '%s' "$SHELL" | rev | cut -d'/' -f1 | rev)" || abort "Can't determine login shell."

  declare default_shell="$result"
  declare answer
  while :; do
    case "$result" in
      bash | zsh)
        printf '[INPUT REQUIRED] Current login shell is "%s". Press ENTER to setup werf for this shell or choose another one.\n[b]ash/[z]sh/[a]bort? Default: %s.\n' "$result" "$default_shell"
        ;;
      *)
        default_shell="bash"
        printf '[INPUT REQUIRED] Current login shell is "%s". This shell is unsupported. Choose another shell or abort installation.\n[b]ash/[z]sh/[a]bort? Default: %s.\n' "$result" "$default_shell"
        ;;
    esac

    if [[ $NON_INTERACTIVE == "yes" ]]; then
      answer=""
    else
      read -r answer
    fi

    case "${answer:-$default_shell}" in
      [bB] | [bB][aA][sS][hH])
        result="bash"
        break
        ;;
      [zZ] | [zZ][sS][hH])
        result="zsh"
        break
        ;;
      [aA] | [aA][bB][oO][rR][tT])
        printf 'Aborted by the user.\n' 1>&2
        exit 0
        ;;
      *)
        printf 'Invalid choice, please retry.\n' 1>&2
        continue
        ;;
    esac
  done

  eval "$result_var_name"="$result"
}

propose_joining_docker_group() {
  declare override_join_docker_group="$1"

  [[ $override_join_docker_group == "no" ]] && return 0

  ensure_cmds_available usermod id
  if ! is_user_in_group "$USER" docker; then
    [[ $override_join_docker_group == "auto" ]] && prompt_yes_no_skip 'werf needs access to the Docker daemon. Add current user to the "docker" group? (root required)' "yes" || return 0
    run_as_root "usermod -aG docker '$USER'" || abort "Can't add user \"$USER\" to group \"docker\"."
  fi
}

setup_trdl_bin_path() {
  declare shell="$1"
  declare override_setup_trdl_bin_path="$2"

  [[ $override_setup_trdl_bin_path == "no" ]] && return 0

  case "$shell" in
    bash)
      declare active_bash_profile_file="$(get_active_bash_profile_file)"
      [[ $override_setup_trdl_bin_path == "auto" ]] && prompt_yes_no_skip "trdl is going to be installed in \"$HOME/bin/\". Add this directory to your \$PATH in \"$HOME/.bashrc\" and \"$active_bash_profile_file\"? (strongly recommended)" "yes" || return 0
      add_trdl_bin_path_setup_to_file "$HOME/.bashrc"
      add_trdl_bin_path_setup_to_file "$active_bash_profile_file"
      ;;
    zsh)
      declare active_zsh_profile_file="$(get_active_zsh_profile_file)"
      [[ $override_setup_trdl_bin_path == "auto" ]] && prompt_yes_no_skip "trdl is going to be installed in \"$HOME/bin/\". Add this directory to your \$PATH in \"$HOME/.zshrc\" and \"$active_zsh_profile_file\"? (strongly recommended)" "yes" || return 0
      add_trdl_bin_path_setup_to_file "$HOME/.zshrc"
      add_trdl_bin_path_setup_to_file "$active_zsh_profile_file"
      ;;
    *) abort "Shell \"$shell\" is not supported." ;;
  esac
}

add_trdl_bin_path_setup_to_file() {
  declare file="$1"

  declare path_append_cmd='[[ "$PATH" == *"$HOME/bin:"* ]] || export PATH="$HOME/bin:$PATH"'
  grep -qsxF -- "$path_append_cmd" "$file" && log::info "Skipping adding \"$HOME/bin/\" to \$PATH in \"$file\": already added." && return 0

  append_line_to_file "$path_append_cmd" "$file"
}

install_trdl() {
  declare os="$1"
  declare arch="$2"
  declare trdl_tuf_repo="$3"
  declare trdl_gpg_pubkey_url="$4"
  declare trdl_version="$5"
  declare skip_signature_verification="$6"

  is_trdl_installed_and_up_to_date "$trdl_version" && log::info "Skipping trdl installation: already installed in \"$HOME/bin/\"." && return 0
  log::info "Installing trdl to \"$HOME/bin/\"."

  mkdir -p "$HOME/bin" || abort "Can't create \"$HOME/bin\" directory."

  declare tmp_dir
  tmp_dir="$(mktemp -d)" || abort "Can't create temporary directory."

  download "${trdl_tuf_repo}/targets/releases/${trdl_version}/${os}-${arch}/bin/trdl" "$tmp_dir/trdl"
  download "${trdl_tuf_repo}/targets/signatures/${trdl_version}/${os}-${arch}/bin/trdl.sig" "$tmp_dir/trdl.sig"

  if [[ $skip_signature_verification == "no" ]]; then
    log::info "Verifying trdl binary signatures."
    download "$trdl_gpg_pubkey_url" "$tmp_dir/trdl.pub"
    gpg -q --import <"$tmp_dir/trdl.pub" || abort "Can't import trdl public gpg key."
    gpg -q --verify "$tmp_dir/trdl.sig" "$tmp_dir/trdl" || abort "Can't verify trdl binary with a gpg signature."
  fi

  install "$tmp_dir/trdl" "$HOME/bin" || abort "Can't install trdl binary from \"$tmp_dir/trdl\" to \"$HOME/bin\"."
  rm -rf "$tmp_dir" 2>/dev/null
}

is_trdl_installed_and_up_to_date() {
  declare required_trdl_version="$1"

  [[ -x "$HOME/bin/trdl" ]] || return 1

  declare current_trdl_version
  current_trdl_version="$("$HOME/bin/trdl" version 2>/dev/null | cut -c2-)" || return 1

  compare_versions "$current_trdl_version" "$required_trdl_version"
  test $? -lt 2
  return
}

add_trdl_werf_repo() {
  declare werf_tuf_repo_url="$1"
  declare werf_tuf_repo_root_version="$2"
  declare werf_tuf_repo_root_sha="$3"

  log::info "Adding werf repo to trdl."
  "$HOME/bin/trdl" add werf "$werf_tuf_repo_url" "$werf_tuf_repo_root_version" "$werf_tuf_repo_root_sha" || abort "Can't add \"werf\" repo to trdl."
}

enable_automatic_werf_activation() {
  declare shell="$1"
  declare override_werf_enable_autoactivation="$2"
  declare werf_autoactivate_version="$3"
  declare werf_autoactivate_channel="$4"

  [[ $override_werf_enable_autoactivation == "no" ]] && return 0

  case "$shell" in
    bash)
      declare active_bash_profile_file="$(get_active_bash_profile_file)"
      [[ $override_werf_enable_autoactivation == "auto" ]] && prompt_yes_no_skip "Add automatic werf activation to \"$HOME/.bashrc\" and \"$active_bash_profile_file\"? (recommended for interactive usage, not recommended for CI)" "yes" || return 0
      add_automatic_werf_activation_to_file "$HOME/.bashrc" "$werf_autoactivate_version" "$werf_autoactivate_channel"
      add_automatic_werf_activation_to_file "$active_bash_profile_file" "$werf_autoactivate_version" "$werf_autoactivate_channel"
      ;;
    zsh)
      declare active_zsh_profile_file="$(get_active_zsh_profile_file)"
      [[ $override_werf_enable_autoactivation == "auto" ]] && prompt_yes_no_skip "Add automatic werf activation to \"$HOME/.zshrc\" and \"$active_zsh_profile_file\"? (recommended for interactive usage, not recommended for CI)" "yes" || return 0
      add_automatic_werf_activation_to_file "$HOME/.zshrc" "$werf_autoactivate_version" "$werf_autoactivate_channel"
      add_automatic_werf_activation_to_file "$active_zsh_profile_file" "$werf_autoactivate_version" "$werf_autoactivate_channel"
      ;;
    *) abort "Shell \"$shell\" is not supported." ;;
  esac
}

add_automatic_werf_activation_to_file() {
  declare file="$1"
  declare werf_autoactivate_version="$2"
  declare werf_autoactivate_channel="$3"

  declare werf_activation_cmd="! { which werf | grep -qsE \"^$HOME/.trdl/\"; } && [[ -x \"\$HOME/bin/trdl\" ]] && source \$(\"\$HOME/bin/trdl\" use werf \"$werf_autoactivate_version\" \"$werf_autoactivate_channel\")"
  grep -qsxF -- "$werf_activation_cmd" "$file" && log::info "Skipping adding werf activation to \"$file\": already added." && return 0

  append_line_to_file "$werf_activation_cmd" "$file"
}

get_active_bash_profile_file() {
  declare result="$HOME/.bash_profile"
  for file in "$HOME/.bash_profile" "$HOME/.bash_login" "$HOME/.profile"; do
    [[ -r $file ]] && result="$file" && break
  done
  printf '%s' "$result"
}

get_active_zsh_profile_file() {
  declare result="$HOME/.zprofile"
  for file in "$HOME/.zprofile" "$HOME/.zlogin"; do
    [[ -r $file ]] && result="$file" && break
  done
  printf '%s' "$result"
}

finalize() {
  declare werf_autoactivate_version="$1"
  declare werf_autoactivate_channel="$2"

  log::info "werf installation finished successfully!"
  log::info "Open new shell session if you have enabled werf autoactivation or activate werf manually with:\n$ source \$(\"$HOME/bin/trdl\" use werf \"$werf_autoactivate_version\" \"$werf_autoactivate_channel\")"
}

append_line_to_file() {
  declare line="$1"
  declare file="$2"

  create_file "$file"

  declare append_cmd="tee -a '$file' <<< '$line' 1>/dev/null"
  [[ -w $file ]] || append_cmd="run_as_root '$append_cmd'"
  eval $append_cmd || abort "Can't append line \"$line\" to file \"$file\"."
}

create_file() {
  declare file="$1"

  [[ -e $file ]] && return 0

  declare stderr
  if ! stderr="$(touch "$file" 1>/dev/null 2>&1)" && [[ $stderr == *"Permission denied"* ]]; then
    run_as_root "touch '$file'" || abort "Unable to create file \"$file\"."
  fi
}

download() {
  declare url="$1"
  declare path="$2"

  declare download_cmd
  if is_command_exists wget; then
    download_cmd="wget -qO '$path' '$url'"
  elif is_command_exists curl; then
    download_cmd="curl -sSLo '$path' '$url'"
  else
    abort 'Neither "wget" nor "curl" available. Install and retry.'
  fi

  eval $download_cmd || abort "Unable to download \"$url\" to \"$path\"."
}

is_user_in_group() {
  declare user="$1"
  declare group="$2"

  id -nG "$user" 2>/dev/null | grep -wqs "$group"
  return
}

ensure_cmds_available() {
  declare cmd
  for cmd in "${@}"; do
    is_command_exists "$cmd" || abort "\"$cmd\" required but not available."
  done
}

run_as_root() {
  declare cmd="$1"

  if [[ $(id -u) -eq 0 ]]; then
    eval $cmd || return 1
  else
    [[ $NON_INTERACTIVE == "yes" ]] && log::warn "Non-interactive mode enabled, but current user doesn't have enough privilege, so the next command will be called with sudo (this might hang): sudo $cmd"
    ensure_cmds_available sudo
    eval sudo $cmd || return 1
  fi

  return 0
}

is_command_exists() {
  declare command="$1"
  command -v "$command" 1>/dev/null 2>&1
  return
}

# Returns 0 if answer is "yes".
# Returns 1 if answer is "skip".
# Aborts if answer is "abort".
# Returns exit code corresponding to the $def_arg if non-interactive.
prompt_yes_no_skip() {
  declare prompt_msg="$1"
  declare def_arg="$2"

  case "$def_arg" in
    [yY] | [yY][eE][sS])
      def_arg="yes"
      ;;
    [aA] | [aA][bB][oO][rR][tT])
      def_arg="abort"
      ;;
    [sS] | [sS][kK][iI][pP])
      def_arg="skip"
      ;;
  esac

  declare answer
  while :; do
    printf '[INPUT REQUIRED] %s\n[y]es/[a]bort/[s]kip? Default: %s.\n' "$prompt_msg" "$def_arg"

    if [[ $NON_INTERACTIVE == "yes" ]]; then
      answer=""
    else
      read -r answer
    fi

    case "${answer:-$def_arg}" in
      [yY] | [yY][eE][sS])
        return 0
        ;;
      [aA] | [aA][bB][oO][rR][tT])
        printf 'Aborted by the user.\n'
        exit 0
        ;;
      [sS] | [sS][kK][iI][pP])
        return 1
        ;;
      *)
        printf 'Invalid choice, please retry.\n'
        continue
        ;;
    esac
  done
}

# Returns 0 if versions equal.
# Returns 1 if $version1 is greater than $version2.
# Returns 2 if $version1 is less than $version2.
compare_versions() {
  declare version1="$1"
  declare version2="$2"

  for ver in "$version1" "$version2"; do
    [[ $ver =~ ^[.0-9]*$ ]] || abort "Invalid version passed to compare_versions(): $ver"
  done

  [[ $version1 == "$version2" ]] && return 0

  declare IFS=.
  declare -a ver1 ver2
  read -r -a ver1 <<<"$version1"
  read -r -a ver2 <<<"$version2"

  for ((i = ${#ver1[@]}; i < ${#ver2[@]}; i++)); do
    ver1[i]=0
  done

  for ((i = 0; i < ${#ver1[@]}; i++)); do
    if [[ -z ${ver2[i]} ]]; then
      ver2[i]=0
    fi
    if ((10#${ver1[i]} > 10#${ver2[i]})); then
      return 1
    fi
    if ((10#${ver1[i]} < 10#${ver2[i]})); then
      return 2
    fi
  done

  return 0
}

get_linux_distro() {
  declare distro
  if [[ -f /etc/os-release ]]; then
    . "/etc/os-release"
    distro="$NAME"
  elif type lsb_release 1>/dev/null 2>&1; then
    distro="$(lsb_release -si)"
  elif [[ -f "/etc/lsb-release" ]]; then
    . "/etc/lsb-release"
    distro="$DISTRIB_ID"
  elif [[ -f "/etc/debian_version" ]]; then
    # Older Debian/Ubuntu/etc.
    distro="debian"
  elif [[ -f "/etc/SuSe-release" ]]; then
    # Older SuSE/etc.
    distro="suse"
  elif [[ -f "/etc/redhat-release" ]]; then
    # Older Red Hat, CentOS, etc.
    distro="redhat"
  else
    abort "Unable to detect Linux distro."
  fi

  printf '%s' "$distro" | tr '[:upper:]' '[:lower:]'
}

log::info() {
  declare msg="$1"

  printf '[INFO] %b\n' "$msg"
}

log::warn() {
  declare msg="$1"

  printf '[WARNING] %b\n' "$msg" >&2
}

log::err() {
  declare msg="$1"

  printf '[ERROR] %b\n' "$msg" >&2
}

abort() {
  declare msg="$1"

  printf '[FATAL] %b\n' "$msg" >&2
  printf '[FATAL] Aborting.\n' >&2
  exit 1
}

main "$@"
