# bash completion for gardenctl                            -*- shell-script -*-

__debug()
{
    if [[ -n ${BASH_COMP_DEBUG_FILE} ]]; then
        echo "$*" >> "${BASH_COMP_DEBUG_FILE}"
    fi
}

# Homebrew on Macs have version 1.3 of bash-completion which doesn't include
# _init_completion. This is a very minimal version of that function.
__my_init_completion()
{
    COMPREPLY=()
    _get_comp_words_by_ref "$@" cur prev words cword
}

__index_of_word()
{
    local w word=$1
    shift
    index=0
    for w in "$@"; do
        [[ $w = "$word" ]] && return
        index=$((index+1))
    done
    index=-1
}

__contains_word()
{
    local w word=$1; shift
    for w in "$@"; do
        [[ $w = "$word" ]] && return
    done
    return 1
}

__handle_reply()
{
    __debug "${FUNCNAME[0]}"
    case $cur in
        -*)
            if [[ $(type -t compopt) = "builtin" ]]; then
                compopt -o nospace
            fi
            local allflags
            if [ ${#must_have_one_flag[@]} -ne 0 ]; then
                allflags=("${must_have_one_flag[@]}")
            else
                allflags=("${flags[*]} ${two_word_flags[*]}")
            fi
            COMPREPLY=( $(compgen -W "${allflags[*]}" -- "$cur") )
            if [[ $(type -t compopt) = "builtin" ]]; then
                [[ "${COMPREPLY[0]}" == *= ]] || compopt +o nospace
            fi

            # complete after --flag=abc
            if [[ $cur == *=* ]]; then
                if [[ $(type -t compopt) = "builtin" ]]; then
                    compopt +o nospace
                fi

                local index flag
                flag="${cur%%=*}"
                __index_of_word "${flag}" "${flags_with_completion[@]}"
                COMPREPLY=()
                if [[ ${index} -ge 0 ]]; then
                    PREFIX=""
                    cur="${cur#*=}"
                    ${flags_completion[${index}]}
                    if [ -n "${ZSH_VERSION}" ]; then
                        # zsh completion needs --flag= prefix
                        eval "COMPREPLY=( \"\${COMPREPLY[@]/#/${flag}=}\" )"
                    fi
                fi
            fi
            return 0;
            ;;
    esac

    # check if we are handling a flag with special work handling
    local index
    __index_of_word "${prev}" "${flags_with_completion[@]}"
    if [[ ${index} -ge 0 ]]; then
        ${flags_completion[${index}]}
        return
    fi

    # we are parsing a flag and don't have a special handler, no completion
    if [[ ${cur} != "${words[cword]}" ]]; then
        return
    fi

    local completions
    completions=("${commands[@]}")
    if [[ ${#must_have_one_noun[@]} -ne 0 ]]; then
        completions=("${must_have_one_noun[@]}")
    fi
    if [[ ${#must_have_one_flag[@]} -ne 0 ]]; then
        completions+=("${must_have_one_flag[@]}")
    fi
    COMPREPLY=( $(compgen -W "${completions[*]}" -- "$cur") )

    if [[ ${#COMPREPLY[@]} -eq 0 && ${#noun_aliases[@]} -gt 0 && ${#must_have_one_noun[@]} -ne 0 ]]; then
        COMPREPLY=( $(compgen -W "${noun_aliases[*]}" -- "$cur") )
    fi

    if [[ ${#COMPREPLY[@]} -eq 0 ]]; then
        declare -F __custom_func >/dev/null && __custom_func
    fi

    # available in bash-completion >= 2, not always present on macOS
    if declare -F __ltrim_colon_completions >/dev/null; then
        __ltrim_colon_completions "$cur"
    fi
}

# The arguments should be in the form "ext1|ext2|extn"
__handle_filename_extension_flag()
{
    local ext="$1"
    _filedir "@(${ext})"
}

__handle_subdirs_in_dir_flag()
{
    local dir="$1"
    pushd "${dir}" >/dev/null 2>&1 && _filedir -d && popd >/dev/null 2>&1
}

__handle_flag()
{
    __debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    # if a command required a flag, and we found it, unset must_have_one_flag()
    local flagname=${words[c]}
    local flagvalue
    # if the word contained an =
    if [[ ${words[c]} == *"="* ]]; then
        flagvalue=${flagname#*=} # take in as flagvalue after the =
        flagname=${flagname%%=*} # strip everything after the =
        flagname="${flagname}=" # but put the = back
    fi
    __debug "${FUNCNAME[0]}: looking for ${flagname}"
    if __contains_word "${flagname}" "${must_have_one_flag[@]}"; then
        must_have_one_flag=()
    fi

    # if you set a flag which only applies to this command, don't show subcommands
    if __contains_word "${flagname}" "${local_nonpersistent_flags[@]}"; then
      commands=()
    fi

    # keep flag value with flagname as flaghash
    if [ -n "${flagvalue}" ] ; then
        flaghash[${flagname}]=${flagvalue}
    elif [ -n "${words[ $((c+1)) ]}" ] ; then
        flaghash[${flagname}]=${words[ $((c+1)) ]}
    else
        flaghash[${flagname}]="true" # pad "true" for bool flag
    fi

    # skip the argument to a two word flag
    if __contains_word "${words[c]}" "${two_word_flags[@]}"; then
        c=$((c+1))
        # if we are looking for a flags value, don't show commands
        if [[ $c -eq $cword ]]; then
            commands=()
        fi
    fi

    c=$((c+1))

}

__handle_noun()
{
    __debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    if __contains_word "${words[c]}" "${must_have_one_noun[@]}"; then
        must_have_one_noun=()
    elif __contains_word "${words[c]}" "${noun_aliases[@]}"; then
        must_have_one_noun=()
    fi

    nouns+=("${words[c]}")
    c=$((c+1))
}

__handle_command()
{
    __debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    local next_command
    if [[ -n ${last_command} ]]; then
        next_command="_${last_command}_${words[c]//:/__}"
    else
        if [[ $c -eq 0 ]]; then
            next_command="_$(basename "${words[c]//:/__}")"
        else
            next_command="_${words[c]//:/__}"
        fi
    fi
    c=$((c+1))
    __debug "${FUNCNAME[0]}: looking for ${next_command}"
    declare -F "$next_command" >/dev/null && $next_command
}

__handle_word()
{
    if [[ $c -ge $cword ]]; then
        __handle_reply
        return
    fi
    __debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"
    if [[ "${words[c]}" == -* ]]; then
        __handle_flag
    elif __contains_word "${words[c]}" "${commands[@]}"; then
        __handle_command
    elif [[ $c -eq 0 ]] && __contains_word "$(basename "${words[c]}")" "${commands[@]}"; then
        __handle_command
    else
        __handle_noun
    fi
    __handle_word
}

_gardenctl_ls()
{
    last_command="gardenctl_ls"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--no-cache")
    flags+=("-n")
    flags+=("--output=")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    must_have_one_noun+=("gardens")
    must_have_one_noun+=("issues")
    must_have_one_noun+=("projects")
    must_have_one_noun+=("seeds")
    must_have_one_noun+=("shoots")
    noun_aliases=()
}

_gardenctl_target()
{
    last_command="gardenctl_target"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--garden")
    flags+=("-g")
    flags+=("--project")
    flags+=("-p")
    flags+=("--seed")
    flags+=("-s")
    flags+=("--no-cache")
    flags+=("-n")
    flags+=("--output=")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    must_have_one_noun+=("garden")
    must_have_one_noun+=("project")
    must_have_one_noun+=("seed")
    must_have_one_noun+=("shoot")
    noun_aliases=()
}

_gardenctl_drop()
{
    last_command="gardenctl_drop"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--no-cache")
    flags+=("-n")
    flags+=("--output=")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    must_have_one_noun+=("project")
    must_have_one_noun+=("seed")
    noun_aliases=()
}

_gardenctl_get()
{
    last_command="gardenctl_get"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--no-cache")
    flags+=("-n")
    flags+=("--output=")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    must_have_one_noun+=("garden")
    must_have_one_noun+=("project")
    must_have_one_noun+=("seed")
    must_have_one_noun+=("shoot")
    must_have_one_noun+=("target")
    noun_aliases=()
}

_gardenctl_download()
{
    last_command="gardenctl_download"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--no-cache")
    flags+=("-n")
    flags+=("--output=")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    must_have_one_noun+=("tf")
    noun_aliases=()
}

_gardenctl_show()
{
    last_command="gardenctl_show"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--no-cache")
    flags+=("-n")
    flags+=("--output=")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    must_have_one_noun+=("addon-manager")
    must_have_one_noun+=("alertmanager")
    must_have_one_noun+=("api")
    must_have_one_noun+=("controller-manager")
    must_have_one_noun+=("dashboard")
    must_have_one_noun+=("etcd-events")
    must_have_one_noun+=("etcd-main")
    must_have_one_noun+=("etcd-operator")
    must_have_one_noun+=("grafana")
    must_have_one_noun+=("machine-controller-manager")
    must_have_one_noun+=("operator")
    must_have_one_noun+=("prometheus")
    must_have_one_noun+=("scheduler")
    must_have_one_noun+=("tf")
    must_have_one_noun+=("ui")
    must_have_one_noun+=("vpn-seed")
    must_have_one_noun+=("vpn-shoot")
    noun_aliases=()
}

_gardenctl_logs()
{
    last_command="gardenctl_logs"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--no-cache")
    flags+=("-n")
    flags+=("--output=")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    must_have_one_noun+=("addon-manager")
    must_have_one_noun+=("alertmanager")
    must_have_one_noun+=("api")
    must_have_one_noun+=("auto-node-repair")
    must_have_one_noun+=("controller-manager")
    must_have_one_noun+=("dashboard")
    must_have_one_noun+=("etcd-events")
    must_have_one_noun+=("etcd-main")
    must_have_one_noun+=("etcd-operator")
    must_have_one_noun+=("gardener-apiserver")
    must_have_one_noun+=("gardener-controller-manager")
    must_have_one_noun+=("grafana")
    must_have_one_noun+=("prometheus")
    must_have_one_noun+=("scheduler")
    must_have_one_noun+=("tf")
    must_have_one_noun+=("ui")
    must_have_one_noun+=("vpn-seed")
    must_have_one_noun+=("vpn-shoot")
    noun_aliases=()
}

_gardenctl_register()
{
    last_command="gardenctl_register"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--all")
    flags+=("-a")
    flags+=("--no-cache")
    flags+=("-n")
    flags+=("--output=")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gardenctl_unregister()
{
    last_command="gardenctl_unregister"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--all")
    flags+=("-a")
    flags+=("--no-cache")
    flags+=("-n")
    flags+=("--output=")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gardenctl_completion()
{
    last_command="gardenctl_completion"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--help")
    flags+=("-h")
    local_nonpersistent_flags+=("--help")
    flags+=("--no-cache")
    flags+=("-n")
    flags+=("--output=")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gardenctl_shell()
{
    last_command="gardenctl_shell"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--image=")
    two_word_flags+=("-i")
    flags+=("--no-cache")
    flags+=("-n")
    flags+=("--output=")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gardenctl_kubectl()
{
    last_command="gardenctl_kubectl"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--no-cache")
    flags+=("-n")
    flags+=("--output=")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gardenctl_aws()
{
    last_command="gardenctl_aws"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--no-cache")
    flags+=("-n")
    flags+=("--output=")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gardenctl_az()
{
    last_command="gardenctl_az"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--no-cache")
    flags+=("-n")
    flags+=("--output=")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gardenctl_gcloud()
{
    last_command="gardenctl_gcloud"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--no-cache")
    flags+=("-n")
    flags+=("--output=")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gardenctl_openstack()
{
    last_command="gardenctl_openstack"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--no-cache")
    flags+=("-n")
    flags+=("--output=")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gardenctl_version()
{
    last_command="gardenctl_version"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--no-cache")
    flags+=("-n")
    flags+=("--output=")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gardenctl()
{
    last_command="gardenctl"
    commands=()
    commands+=("ls")
    commands+=("target")
    commands+=("drop")
    commands+=("get")
    commands+=("download")
    commands+=("show")
    commands+=("logs")
    commands+=("register")
    commands+=("unregister")
    commands+=("completion")
    commands+=("shell")
    commands+=("kubectl")
    commands+=("aws")
    commands+=("az")
    commands+=("gcloud")
    commands+=("openstack")
    commands+=("version")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--no-cache")
    flags+=("-n")
    flags+=("--output=")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

__start_gardenctl()
{
    local cur prev words cword
    declare -A flaghash 2>/dev/null || :
    if declare -F _init_completion >/dev/null 2>&1; then
        _init_completion -s || return
    else
        __my_init_completion -n "=" || return
    fi

    local c=0
    local flags=()
    local two_word_flags=()
    local local_nonpersistent_flags=()
    local flags_with_completion=()
    local flags_completion=()
    local commands=("gardenctl")
    local must_have_one_flag=()
    local must_have_one_noun=()
    local last_command
    local nouns=()

    __handle_word
}

if [[ $(type -t compopt) = "builtin" ]]; then
    complete -o default -F __start_gardenctl gardenctl
else
    complete -o default -o nospace -F __start_gardenctl gardenctl
fi

# ex: ts=4 sw=4 et filetype=sh
