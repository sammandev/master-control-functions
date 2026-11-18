import assert from 'node:assert/strict';
import { existsSync, mkdtempSync, mkdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import test from 'node:test';

const deployment = fileURLToPath(new URL('../blue-green.sh', import.meta.url));
const bash = process.env.DEPLOY_TEST_BASH ?? (process.platform === 'win32'
  ? [
      'C:/Program Files/Git/bin/bash.exe',
      `${process.env.LOCALAPPDATA}/Programs/Git/bin/bash.exe`,
    ].find(existsSync)
  : 'bash');

function shellPath(value) {
  return value.replaceAll('\\', '/').replace(/^([A-Za-z]):/, (_, drive) => `/${drive.toLowerCase()}`);
}

function runDeployment(context, scenario, args = ['deploy', 'test-release'], buildEnvironment = {}, proxyFile) {
  assert.ok(bash, 'Bash is required for database-free rollout tests');
  const root = mkdtempSync(path.join(tmpdir(), 'webprogfndb-rollout-'));
  context.after(() => rmSync(root, { recursive: true, force: true }));
  const directory = path.join(root, 'deploy');
  const commands = path.join(root, 'bin');
  mkdirSync(path.join(directory, 'state', 'upstreams'), { recursive: true });
  mkdirSync(path.join(directory, 'state', 'assets'), { recursive: true });
  writeFileSync(path.join(directory, 'state', 'assets', 'old-hash.js'), 'old release');
  mkdirSync(commands);
  if (scenario !== 'first') {
    writeFileSync(path.join(directory, 'state', 'upstreams', 'active.conf'), 'server 127.0.0.1:24993;\n');
  }
  writeFileSync(path.join(directory, '.env'), 'DATABASE_URL=postgresql://example\nPGPASSWORD=example\n');
  if (proxyFile) {
    writeFileSync(path.join(directory, '.env.proxy'), proxyFile.content);
  }
  writeFileSync(path.join(directory, 'compose.yml'), 'services: {}\n');
  if (existsSync(deployment)) {
    writeFileSync(path.join(directory, 'blue-green.sh'), readFileSync(deployment));
  }
  const log = path.join(root, 'commands.log');
  writeFileSync(log, '');
  const docker = `#!/usr/bin/env bash
set -eu
printf '%s\\n' "docker $*" >> "$MOCK_LOG"
if [[ "$1" == build ]]; then
  printf '%s\\n' "\${http_proxy:-}" "\${https_proxy:-}" "\${no_proxy:-}" > "$MOCK_BUILD_ENV"
  if [[ "$SCENARIO" == proxy-required && -z "\${https_proxy:-}" ]]; then
    printf 'Go module download: read: connection reset by peer\\n' >&2
    exit 1
  fi
fi
if [[ "$1" == compose && "$*" == *'config --quiet'* ]]; then
  printf '%s\\n' "\${http_proxy:-}" "\${https_proxy:-}" "\${no_proxy:-}" "\${GOPROXY:-}" > "$MOCK_RUNTIME_ENV"
fi
if [[ "$1" == build && "$SCENARIO" == build-failure ]]; then exit 1; fi
if [[ "$1" == image && "$2" == inspect ]]; then
  if [[ "$SCENARIO" == build || "$SCENARIO" == build-failure || "$SCENARIO" == first || "$SCENARIO" == proxy-required ]]; then exit 1; fi
  if [[ "$*" == *--format* ]]; then
    if [[ "$*" == *com.master-control-functions.release* ]]; then
      if [[ "$SCENARIO" == mismatched-release ]]; then printf 'master-control-functions other-release\\n'; else printf 'master-control-functions test-release\\n'; fi
    else
      printf 'master-control-functions\\n'
    fi
  fi
  exit 0
fi
if [[ "$1" == inspect ]]; then
  if [[ "$*" == *State.Health.Status* ]]; then
    if [[ "$SCENARIO" == unhealthy && "$*" == *green* ]]; then printf 'unhealthy\\n'; else printf 'healthy\\n'; fi
  elif [[ "$*" == *State.Running* ]]; then
    if [[ "$SCENARIO" == first && "$*" == *edge* ]]; then exit 1; fi
    printf 'master-control-functions true\\n'
  fi
  exit 0
fi
if [[ "$1" == exec && "$*" == *'ps -o args'* ]]; then
  printf 'nginx: worker process\\n'
  exit 0
fi
if [[ "$1" == exec && "$*" == *'nginx -t'* && "$SCENARIO" == invalid ]]; then
  if grep -q 24994 "$MOCK_DEPLOY/state/upstreams/active.conf"; then exit 1; fi
fi
if [[ "$1" == exec && "$*" == *'nginx -s reload'* && "$SCENARIO" == reload-failure ]]; then
  if grep -q 24994 "$MOCK_DEPLOY/state/upstreams/active.conf"; then exit 1; fi
fi
if [[ "$1" == cp ]]; then
  printf 'new release' > "\${@: -1}/new-hash.js"
fi
exit 0
`;
  const curl = `#!/usr/bin/env bash
set -eu
printf '%s\\n' "curl $*" >> "$MOCK_LOG"
headers=''
output=''
while (( $# )); do
  case "$1" in
    --dump-header) headers="$2"; shift 2 ;;
    --output) output="$2"; shift 2 ;;
    *) shift ;;
  esac
done
port=24994
if grep -q 24993 "$MOCK_DEPLOY/state/upstreams/active.conf"; then port=24993; fi
if [[ "$SCENARIO" == wrong-route ]]; then port=24993; fi
if [[ -n "$headers" ]]; then printf 'HTTP/1.1 200 OK\\r\\nX-Release-Upstream: 127.0.0.1:%s\\r\\n\\r\\n' "$port" > "$headers"; fi
if [[ -n "$output" ]]; then printf '{"status":"ok","db":"ok"}' > "$output"; fi
`;
  for (const [name, content] of Object.entries({
    docker,
    curl,
    stat: '#!/usr/bin/env bash\nif [[ "${@: -1}" == *.env.proxy ]]; then printf "%s\\n" "$MOCK_PROXY_MODE"; else printf "600\\n"; fi\n',
    ss: '#!/usr/bin/env bash\nexit 0\n',
    flock: '#!/usr/bin/env bash\nexit 0\n',
    sleep: '#!/usr/bin/env bash\nexit 0\n',
  })) {
    writeFileSync(path.join(commands, name), content, { mode: 0o755 });
  }
  const environment = { ...process.env };
  for (const name of ['HTTP_PROXY', 'HTTPS_PROXY', 'NO_PROXY', 'http_proxy', 'https_proxy', 'no_proxy', 'GOPROXY']) {
    delete environment[name];
  }
  Object.assign(environment, buildEnvironment);
  const currentPath = Object.keys(environment).find(key => key.toLowerCase() === 'path');
  const inheritedPath = currentPath ? environment[currentPath] : '';
  if (currentPath) delete environment[currentPath];
  environment.PATH = `${commands}${path.delimiter}${inheritedPath}`;
  environment.MOCK_LOG = shellPath(log);
  environment.MOCK_DEPLOY = shellPath(directory);
  environment.MOCK_BIN = shellPath(commands);
  const buildProxy = path.join(root, 'build-proxy.txt');
  const runtimeEnvironment = path.join(root, 'runtime-env.txt');
  environment.MOCK_BUILD_ENV = shellPath(buildProxy);
  environment.MOCK_RUNTIME_ENV = shellPath(runtimeEnvironment);
  environment.MOCK_PROXY_MODE = proxyFile?.mode ?? '600';
  const startup = path.join(root, 'mock-env.sh');
  writeFileSync(startup, 'export PATH="$MOCK_BIN:$PATH"\n');
  environment.BASH_ENV = shellPath(startup);
  environment.SCENARIO = scenario;
  const result = spawnSync(bash, [shellPath(path.join(directory, 'blue-green.sh')), ...args], {
    cwd: directory,
    env: environment,
    encoding: 'utf8',
    timeout: 15_000,
  });
  assert.ifError(result.error && new Error(`${result.error.message}\n${result.stderr}\n${readFileSync(log, 'utf8')}`));
  return {
    ...result,
    commands: readFileSync(log, 'utf8'),
    buildProxy: existsSync(buildProxy) ? readFileSync(buildProxy, 'utf8').trimEnd().split('\n') : [],
    runtimeEnvironment: existsSync(runtimeEnvironment) ? readFileSync(runtimeEnvironment, 'utf8') : null,
    active: existsSync(path.join(directory, 'state', 'upstreams', 'active.conf'))
      ? readFileSync(path.join(directory, 'state', 'upstreams', 'active.conf'), 'utf8')
      : null,
    previousAsset: readFileSync(path.join(directory, 'state', 'assets', 'old-hash.js'), 'utf8'),
    candidateAsset: existsSync(path.join(directory, 'state', 'assets', 'new-hash.js')),
  };
}

test('an unhealthy candidate leaves active blue routing unchanged', context => {
  const result = runDeployment(context, 'unhealthy');
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /not healthy/);
  assert.equal(result.active, 'server 127.0.0.1:24993;\n');
  assert.doesNotMatch(result.commands, /nginx -s reload|docker stop webprogfndb-blue/);
});

test('invalid Nginx configuration restores the previous upstream without reloading', context => {
  const result = runDeployment(context, 'invalid');
  assert.notEqual(result.status, 0);
  assert.equal(result.active, 'server 127.0.0.1:24993;\n');
  assert.match(result.commands, /exec webprogfndb-edge nginx -t/);
  assert.doesNotMatch(result.commands, /nginx -s reload|docker stop/);
});

test('a healthy release switches to green and retains blue and both asset versions', context => {
  const result = runDeployment(context, 'healthy');
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.active, 'server 127.0.0.1:24994;\n');
  assert.match(result.commands, /exec webprogfndb-edge nginx -t[\s\S]*nginx -s reload/);
  assert.doesNotMatch(result.commands, /docker stop|docker image rm|prune|up .*edge/);
  assert.equal(result.previousAsset, 'old release');
  assert.equal(result.candidateAsset, true);
});

test('a failed reload restores and verifies previous blue routing', context => {
  const result = runDeployment(context, 'reload-failure');
  assert.notEqual(result.status, 0);
  assert.equal(result.active, 'server 127.0.0.1:24993;\n');
  assert.match(result.stderr, /Previous routing restored/);
  assert.equal((result.commands.match(/nginx -s reload/g) ?? []).length, 2);
  assert.doesNotMatch(result.commands, /docker stop|up .*edge/);
});

test('a wrong post-switch upstream rolls back instead of reporting success', context => {
  const result = runDeployment(context, 'wrong-route');
  assert.notEqual(result.status, 0);
  assert.equal(result.active, 'server 127.0.0.1:24993;\n');
  assert.match(result.stderr, /Previous routing restored/);
});

test('first deployment refuses to start containers without migration confirmation', context => {
  const result = runDeployment(context, 'first');
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /SQLite transfer must be verified/);
  assert.equal(result.active, null);
  assert.doesNotMatch(result.commands, /docker build|up -d|nginx/);
});

test('verified first deployment tests edge config before publishing blue', context => {
  const result = runDeployment(context, 'first', ['deploy', 'test-release', '--migration-verified']);
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.active, 'server 127.0.0.1:24993;\n');
  assert.match(result.commands, /run --rm --no-deps edge -t[\s\S]*up -d --no-deps edge/);
});

test('rollback selects a retained healthy slot without building or recreating edge', context => {
  const result = runDeployment(context, 'healthy', ['rollback', 'green']);
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.active, 'server 127.0.0.1:24994;\n');
  assert.doesNotMatch(result.commands, /docker build|force-recreate|up .*edge|docker stop/);
});

test('an image retagged from another release is refused before replacing a slot', context => {
  const result = runDeployment(context, 'mismatched-release');
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /release label/);
  assert.equal(result.active, 'server 127.0.0.1:24993;\n');
  assert.doesNotMatch(result.commands, /force-recreate|nginx -s reload/);
});

test('a new release is built with project and release ownership labels', context => {
  const result = runDeployment(context, 'build');
  assert.equal(result.status, 0, result.stderr);
  assert.match(result.commands, /docker build --pull[\s\S]*--label com.master-control-functions.project=master-control-functions[\s\S]*--label com.master-control-functions.release=test-release/);
  assert.equal(result.active, 'server 127.0.0.1:24994;\n');
});

test('build forwards GOPROXY as an argument and proxy settings as secrets without logging proxy values', context => {
  const buildEnvironment = {
    GOPROXY: 'https://modules.example.invalid',
    HTTP_PROXY: 'http://user:example@proxy.example.invalid:8080',
    HTTPS_PROXY: 'http://build-user:build-password@proxy.example.invalid:8080',
    NO_PROXY: 'localhost,127.0.0.1',
  };
  const result = runDeployment(context, 'build', ['deploy', 'test-release'], buildEnvironment);
  assert.equal(result.status, 0, result.stderr);
  assert.ok(result.commands.includes('--build-arg GOPROXY=https://modules.example.invalid'), 'GOPROXY must reach Docker build');
  assert.ok(result.commands.includes('--secret id=https_proxy,env=https_proxy'), 'HTTPS_PROXY must reach Docker build as a secret');
  assert.ok(result.commands.includes('--secret id=http_proxy,env=http_proxy'), 'http_proxy secret must cover the lowercase form used inside RUN');
  assert.ok(result.commands.includes('--secret id=no_proxy,env=no_proxy'), 'NO_PROXY must reach Docker build as a secret');
  for (const [name, value] of Object.entries(buildEnvironment).filter(([name]) => name !== 'GOPROXY')) {
    assert.equal(result.commands.includes(value), false, `${name} must not appear in command output`);
  }
});

test('no arguments builds a uniquely tagged release and switches the healthy candidate', context => {
  const result = runDeployment(context, 'build', []);
  assert.equal(result.status, 0, result.stderr);
  assert.match(result.commands, /-t master-control-functions:\d{8}T\d{15}Z-\d+/);
  assert.equal(result.active, 'server 127.0.0.1:24994;\n');
});

test('deploy without a tag also generates a release', context => {
  const result = runDeployment(context, 'build', ['deploy']);
  assert.equal(result.status, 0, result.stderr);
  assert.match(result.commands, /-t master-control-functions:\d{8}T\d{15}Z-\d+/);
});

test('a generated tag collision fails rather than reusing an existing image', context => {
  const result = runDeployment(context, 'healthy', []);
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /Generated release tag already exists/);
  assert.equal(result.active, 'server 127.0.0.1:24993;\n');
  assert.doesNotMatch(result.commands, /docker build|force-recreate|nginx -s reload/);
});

test('argument-free first deployment still requires migration confirmation', context => {
  const result = runDeployment(context, 'first', []);
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /SQLite transfer must be verified/);
  assert.doesNotMatch(result.commands, /docker build|up -d/);
});

test('automatic release accepts explicit first-migration confirmation', context => {
  const result = runDeployment(context, 'first', ['deploy', '--migration-verified']);
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.active, 'server 127.0.0.1:24993;\n');
});

test('private proxy file is loaded automatically and overrides inherited proxy settings', context => {
  const proxy = 'http://user:example@approved.example.invalid:8080';
  const result = runDeployment(context, 'build', ['deploy', 'test-release'], {
    HTTPS_PROXY: 'http://old.example.invalid:8080',
  }, {
    content: `http_proxy='${proxy}'\nhttps_proxy='${proxy}'\nno_proxy='localhost,127.0.0.1'\nGOPROXY='https://modules.example.invalid'\n`,
  });
  assert.equal(result.status, 0, result.stderr);
  assert.deepEqual(result.buildProxy, [proxy, proxy, 'localhost,127.0.0.1']);
  assert.match(result.commands, /--secret id=https_proxy,env=https_proxy/);
  assert.match(result.commands, /--build-arg GOPROXY=https:\/\/modules.example.invalid/);
  assert.equal(`${result.commands}${result.stdout}${result.stderr}`.includes(proxy), false);
  assert.equal(result.runtimeEnvironment, '\n\n\n\n');
});

test('a proxy-required build fails safely without the profile and succeeds when it is loaded', context => {
  const missing = runDeployment(context, 'proxy-required');
  assert.notEqual(missing.status, 0);
  assert.match(missing.stderr, /connection reset by peer/);
  assert.match(missing.stderr, /check the approved proxy/);
  assert.equal(missing.active, 'server 127.0.0.1:24993;\n');
  assert.doesNotMatch(missing.commands, /force-recreate|nginx -s reload/);

  const configured = runDeployment(context, 'proxy-required', ['deploy', 'test-release'], {}, {
    content: "HTTPS_PROXY='http://approved.example.invalid:8080'\n",
  });
  assert.equal(configured.status, 0, configured.stderr);
  assert.equal(configured.buildProxy[1], 'http://approved.example.invalid:8080');
  assert.equal(configured.active, 'server 127.0.0.1:24994;\n');
});

test('an invalid shell profile fails without replacing a release', context => {
  const result = runDeployment(context, 'build', ['deploy', 'test-release'], {}, {
    content: "https_proxy='unterminated\n",
  });
  assert.notEqual(result.status, 0);
  assert.equal(result.active, 'server 127.0.0.1:24993;\n');
  assert.doesNotMatch(result.commands, /docker build|force-recreate|nginx -s reload/);
});

for (const [label, proxyFile, message] of [
  ['world-readable', { content: "https_proxy='http://approved.example.invalid:8080'\n", mode: '644' }, /chmod 600/],
  ['placeholder', { content: "https_proxy='replace_me'\n" }, /placeholder/],
]) {
  test(`${label} proxy file is refused before building or changing containers`, context => {
    const result = runDeployment(context, 'build', ['deploy', 'test-release'], {}, proxyFile);
    assert.notEqual(result.status, 0);
    assert.match(result.stderr, message);
    assert.equal(result.active, 'server 127.0.0.1:24993;\n');
    assert.doesNotMatch(result.commands, /docker build|force-recreate|nginx -s reload/);
  });
}

test('help needs no Docker or runtime configuration', context => {
  const result = runDeployment(context, 'first', ['--help']);
  assert.equal(result.status, 0, result.stderr);
  assert.match(result.stdout, /Usage:/);
  assert.equal(result.commands, '');
});

test('frontend builder invokes the installed scoped Paraglide CLI without package auto-install', () => {
  const dockerfile = readFileSync(new URL('../Dockerfile', import.meta.url), 'utf8');
  assert.match(dockerfile, /node \.\/node_modules\/@inlang\/paraglide-js\/bin\/run\.js compile --project \.\/project\.inlang/);
  assert.doesNotMatch(dockerfile, /bun x paraglide-js/);
  assert.match(dockerfile, /--outdir \.\/src\/lib\/paraglide --emit-ts-declarations/);
  assert.match(dockerfile, /&& bun run check && bun run build/);
});

test('frontend builder enables Node environment proxies for compilation and Vite', () => {
  const dockerfile = readFileSync(new URL('../Dockerfile', import.meta.url), 'utf8');
  const frontend = dockerfile.split(' AS frontend')[1].split('FROM golang:')[0];
  assert.match(frontend, /^ENV NODE_USE_ENV_PROXY=1\r?$/m);
  assert.ok(frontend.indexOf('ENV NODE_USE_ENV_PROXY=1') < frontend.indexOf('bun pm migrate'));
  const runtime = dockerfile.split(' AS runtime')[1];
  assert.doesNotMatch(runtime, /NODE_USE_ENV_PROXY/);
});

test('frontend builder copies pnpm security overrides before migrating the lockfile', () => {
  const dockerfile = readFileSync(new URL('../Dockerfile', import.meta.url), 'utf8');
  const copy = 'COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./';
  assert.ok(dockerfile.includes(copy), 'dependency installation must include the override policy');
  assert.ok(dockerfile.indexOf(copy) < dockerfile.indexOf('bun pm migrate'));
  assert.match(dockerfile, /bun pm migrate && bun install --frozen-lockfile/);
});

test('Go builder matches the patched module toolchain requirement', () => {
  const dockerfile = readFileSync(new URL('../Dockerfile', import.meta.url), 'utf8');
  const module = readFileSync(new URL('../../backend/go.mod', import.meta.url), 'utf8');
  const required = module.match(/^go (\d+\.\d+\.\d+)$/m)?.[1];
  assert.ok(required, 'go.mod must specify the minimum patched Go release');
  assert.match(dockerfile, new RegExp(`^ARG GO_VERSION=${required.replaceAll('.', '\\.')}\\r?$`, 'm'));
  assert.match(dockerfile, /^FROM golang:\$\{GO_VERSION\}-alpine AS backend\r?$/m);
});

test('Go builder accepts a module proxy and mounts proxy secrets without persisting HTTP proxy arguments', () => {
  const dockerfile = readFileSync(new URL('../Dockerfile', import.meta.url), 'utf8');
  assert.match(dockerfile, /FROM golang:[^\r\n]+ AS backend\r?\nARG GOPROXY=https:\/\/proxy\.golang\.org,direct\r?\n/);
  assert.doesNotMatch(dockerfile, /(?:ARG|ENV)\s+(?:HTTP_PROXY|HTTPS_PROXY|NO_PROXY|http_proxy|https_proxy|no_proxy)\b/);
  assert.doesNotMatch(dockerfile, /GOSUMDB\s*=\s*off/);
  const secretMounts = dockerfile.match(/--mount=type=secret,id=https_proxy,env=https_proxy/g) ?? [];
  assert.ok(secretMounts.length >= 3, 'every network-reaching RUN must mount the proxy secrets');
});

test('a failed image build leaves active routing and release containers unchanged', context => {
  const result = runDeployment(context, 'build-failure');
  assert.notEqual(result.status, 0);
  assert.equal(result.active, 'server 127.0.0.1:24993;\n');
  assert.match(result.commands, /docker build/);
  assert.doesNotMatch(result.commands, /force-recreate|nginx -s reload|docker stop/);
  assert.equal(result.candidateAsset, false);
});
