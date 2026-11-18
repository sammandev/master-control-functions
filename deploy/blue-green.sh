#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

DEPLOY_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
REPO_DIR=$(dirname -- "$DEPLOY_DIR")
cd -- "$DEPLOY_DIR"

fail() {
    printf '%s\n' "$*" >&2
    exit 1
}

compose() {
    docker compose --project-name master-control-functions \
        --project-directory "$DEPLOY_DIR" -f "$DEPLOY_DIR/compose.yml" "$@"
}

active_color() {
    if [[ ! -f state/upstreams/active.conf ]]; then
        printf 'none\n'
    elif grep -Fxq 'server 127.0.0.1:24993;' state/upstreams/active.conf; then
        printf 'blue\n'
    elif grep -Fxq 'server 127.0.0.1:24994;' state/upstreams/active.conf; then
        printf 'green\n'
    else
        fail 'Unrecognized active upstream; restore its last verified configuration first.'
    fi
}

slot_port() {
    case "$1" in
        blue) printf '24993\n' ;;
        green) printf '24994\n' ;;
        *) fail 'Invalid release slot.' ;;
    esac
}

wait_for_health() {
    local color="$1" status attempt
    for ((attempt = 0; attempt < 90; attempt++)); do
        status=$(docker inspect --format \
            '{{if .State.Health}}{{.State.Health.Status}}{{else}}missing{{end}}' \
            "webprogfndb-$color") || fail "Candidate $color could not be inspected."
        case "$status" in
            healthy) return 0 ;;
            unhealthy|missing) fail "Candidate $color is not healthy; active routing unchanged." ;;
        esac
        sleep 1
    done
    fail "Candidate $color is not healthy; active routing unchanged."
}

owned_running() {
    local container="$1" details
    details=$(docker inspect --format \
        '{{index .Config.Labels "com.master-control-functions.project"}} {{.State.Running}}' \
        "$container" 2>/dev/null) || return 1
    [[ "$details" == 'master-control-functions '* ]] || fail "Container name belongs to another project: $container"
    [[ "$details" == 'master-control-functions true' ]]
}

require_free_port() {
    local listeners
    listeners=$(ss -H -ltn "sport = :$1") || fail 'Could not inspect port availability.'
    [[ -z "$listeners" ]] || fail "Port $1 is occupied; leave its service unchanged."
}

wait_for_drain() {
    local attempt workers
    for ((attempt = 0; attempt < 90; attempt++)); do
        workers=$(docker exec webprogfndb-edge ps -o args) || fail 'Could not inspect draining edge workers.'
        if ! grep -Fq 'worker process is shutting down' <<< "$workers"; then return 0; fi
        sleep 1
    done
    fail 'Old Nginx workers still draining; no release container was replaced. Retry later.'
}

verify_route() {
    local color="$1" attempt port expected_header header
    port=$(slot_port "$color")
    expected_header="X-Release-Upstream: 127.0.0.1:$port"
    for ((attempt = 0; attempt < 20; attempt++)); do
        if curl --silent --show-error --fail --noproxy 127.0.0.1 \
            --connect-timeout 2 --max-time 5 --dump-header "$probe_headers" \
            --output /dev/null http://127.0.0.1:31627/health; then
            while IFS= read -r header; do
                header=${header%$'\r'}
                if [[ "${header,,}" == "${expected_header,,}" ]]; then
                    return 0
                fi
            done < "$probe_headers"
        fi
        sleep 1
    done
    return 1
}

finish() {
    local status=$?
    trap - EXIT
    if [[ "$switch_pending" == true && "$status" != 0 ]]; then
        if [[ "$current" != none ]]; then
            cp state/upstreams/previous.conf state/upstreams/restore.conf
            chmod 644 state/upstreams/restore.conf
            mv -f state/upstreams/restore.conf state/upstreams/active.conf
            if [[ "$reload_attempted" == true ]]; then
                if docker exec webprogfndb-edge nginx -t \
                    && docker exec webprogfndb-edge nginx -s reload \
                    && verify_route "$current"; then
                    printf 'Previous routing restored; both application slots retained.\n' >&2
                else
                    printf 'CRITICAL: previous file restored but live rollback could not be verified; inspect edge privately.\n' >&2
                fi
            fi
        else
            if owned_running webprogfndb-edge; then docker stop webprogfndb-edge; fi
            rm -f state/upstreams/active.conf
            printf 'Initial publication failed; no release marked active.\n' >&2
        fi
    fi
    rm -rf -- "$asset_stage"
    rm -f -- "$probe_headers"
    exit "$status"
}

automatic_release=false
command_name="${1:-deploy}"
if (( $# )); then shift; fi
case "$command_name" in
    -h|--help)
        printf 'Usage: ./deploy/blue-green.sh [deploy [RELEASE] [--migration-verified]|rollback blue|rollback green|status]\n'
        printf 'No arguments deploys current source with a unique UTC release tag. No source is fetched automatically.\n'
        exit 0
        ;;
    status)
        (( $# == 0 )) || fail 'status takes no arguments.'
        printf 'Configured slot: %s\n' "$(active_color)"
        docker ps --all --filter label=com.master-control-functions.project=master-control-functions \
            --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}'
        exit 0
        ;;
    deploy)
        (( $# <= 2 )) || fail 'deploy accepts an optional release tag and --migration-verified.'
        release=''
        confirmation=''
        for argument in "$@"; do
            case "$argument" in
                --migration-verified)
                    [[ -z "$confirmation" ]] || fail 'Duplicate migration confirmation.'
                    confirmation="$argument"
                    ;;
                *)
                    [[ -z "$release" ]] || fail 'Unknown deployment option.'
                    release="$argument"
                    ;;
            esac
        done
        if [[ -z "$release" ]]; then
            release="$(date -u +%Y%m%dT%H%M%S%NZ)-$$"
            automatic_release=true
        fi
        [[ "$release" =~ ^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,127}$ ]] || fail 'Invalid Docker release tag.'
        ;;
    rollback)
        (( $# == 1 )) || fail 'rollback requires blue or green.'
        slot_port "$1" >/dev/null
        confirmation=''
        ;;
    *) fail 'Usage: ./deploy/blue-green.sh [deploy [RELEASE] [--migration-verified]|rollback blue|rollback green|status]' ;;
esac

for dependency in docker curl flock stat ss; do
    command -v "$dependency" >/dev/null || fail "Required command missing: $dependency"
done
docker compose version >/dev/null
[[ -f .env ]] || fail 'Create deploy/.env from .env.example on Ubuntu first.'
[[ "$(stat -c '%a' .env)" == 600 ]] || fail 'Protect deploy/.env with chmod 600.'
! grep -q replace_me .env || fail 'Replace the placeholder password privately on Ubuntu.'
mkdir -p state/upstreams state/assets
chmod 755 state/upstreams state/assets
exec 9>state/deploy.lock
flock -n 9 || fail 'Another deployment is running.'

current=$(active_color)
if [[ "$current" == none && "$confirmation" != --migration-verified ]]; then
    fail 'Initial SQLite transfer must be verified before deployment; use --migration-verified only after verification.'
fi
edge_running=false
if owned_running webprogfndb-edge; then edge_running=true; fi
if [[ "$current" == none ]]; then
    [[ "$edge_running" == false ]] || fail 'Edge exists without active state; restore verified state first.'
    require_free_port 31627
else
    [[ "$edge_running" == true ]] || fail 'Existing edge is stopped; recover it before another release.'
    verify_route_headers=$(mktemp state/preflight.XXXXXX)
    probe_headers="$verify_route_headers"
    if ! verify_route "$current"; then
        rm -f "$verify_route_headers"
        fail 'Configured and live routing disagree; recover before replacing any slot.'
    fi
    rm -f "$verify_route_headers"
    wait_for_drain
fi

if [[ "$command_name" == deploy ]]; then
    target=blue
    if [[ "$current" == blue ]]; then target=green; fi
    if ! owned_running "webprogfndb-$target"; then require_free_port "$(slot_port "$target")"; fi
    export APP_IMAGE="master-control-functions:$release"
    printf 'Release image: %s. Candidate slot: %s.\n' "$APP_IMAGE" "$target"
    if docker image inspect "$APP_IMAGE" >/dev/null 2>&1; then
        [[ "$automatic_release" == false ]] || fail 'Generated release tag already exists; retry to generate a new tag.'
        labels=$(docker image inspect --format \
            '{{index .Config.Labels "com.master-control-functions.project"}} {{index .Config.Labels "com.master-control-functions.release"}}' \
            "$APP_IMAGE")
        [[ "$labels" == "master-control-functions $release" ]] || fail 'Project or release label does not match this image tag.'
        printf 'Using existing immutable release image %s.\n' "$APP_IMAGE"
    else (
        if [[ -f .env.proxy ]]; then
            [[ "$(stat -c '%a' .env.proxy)" == 600 ]] || fail 'Protect deploy/.env.proxy with chmod 600.'
            ! grep -q replace_me .env.proxy || fail 'Replace the proxy placeholder privately in deploy/.env.proxy.'
            # A private profile replaces inherited proxy settings only for this build.
            unset HTTP_PROXY HTTPS_PROXY NO_PROXY http_proxy https_proxy no_proxy
            set -a
            . "$DEPLOY_DIR/.env.proxy"
            set +a
            printf 'Loaded private build settings from deploy/.env.proxy.\n'
        fi
        # GOPROXY is credential-free and travels as a build argument; proxy
        # settings travel as build secrets so their values never appear on the
        # command line, in image history, or in build cache metadata. Docker
        # receives the values through its own environment (assignment prefix,
        # not arguments); the Dockerfile mounts them as optional RUN secrets.
        build_args=()
        secret_args=()
        if [[ -n "${GOPROXY:-}" ]]; then
            build_args+=(--build-arg "GOPROXY=$GOPROXY")
        fi
        if [[ -n "${HTTP_PROXY:-${http_proxy:-}}" ]]; then
            secret_args+=(--secret 'id=http_proxy,env=http_proxy')
        fi
        if [[ -n "${HTTPS_PROXY:-${https_proxy:-}}" ]]; then
            secret_args+=(--secret 'id=https_proxy,env=https_proxy')
        fi
        if [[ -n "${NO_PROXY:-${no_proxy:-}}" ]]; then
            secret_args+=(--secret 'id=no_proxy,env=no_proxy')
        fi
        http_proxy="${HTTP_PROXY:-${http_proxy:-}}" https_proxy="${HTTPS_PROXY:-${https_proxy:-}}" \
            no_proxy="${NO_PROXY:-${no_proxy:-}}" \
            docker build --pull "${build_args[@]}" "${secret_args[@]}" \
                -f "$DEPLOY_DIR/Dockerfile" -t "$APP_IMAGE" \
                --label com.master-control-functions.project=master-control-functions \
                --label "com.master-control-functions.release=$release" "$REPO_DIR" \
            || fail 'Image build failed; active routing unchanged. For Go download resets, check the approved proxy in deploy/.env.proxy and its BuildKit reachability.'
    )
    fi
    compose config --quiet
    compose up -d --no-deps --force-recreate "app-$target"
else
    target="$1"
    owned_running "webprogfndb-$target" || docker start "webprogfndb-$target" >/dev/null
fi
wait_for_health "$target"
asset_stage=$(mktemp -d state/assets-stage.XXXXXX)
probe_headers=$(mktemp state/probe.XXXXXX)
switch_pending=false
reload_attempted=false
trap finish EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
docker cp "webprogfndb-$target:/app/web/_app/immutable/." "$asset_stage/"
chmod -R a+rX "$asset_stage"
cp -a "$asset_stage/." state/assets/

if [[ "$current" != none ]]; then
    cp state/upstreams/active.conf state/upstreams/previous.conf
fi
printf 'server 127.0.0.1:%s;\n' "$(slot_port "$target")" > state/upstreams/candidate.conf
chmod 644 state/upstreams/candidate.conf
switch_pending=true
mv -f state/upstreams/candidate.conf state/upstreams/active.conf
if [[ "$edge_running" == true ]]; then
    docker exec webprogfndb-edge nginx -t
    reload_attempted=true
    docker exec webprogfndb-edge nginx -s reload
else
    compose run --rm --no-deps edge -t
    compose up -d --no-deps edge
fi
verify_route "$target" || fail 'Live traffic did not reach the candidate; restoring previous routing.'
switch_pending=false
printf 'Active slot: %s. Old slot retained for rollback; no image pruning performed.\n' "$target"