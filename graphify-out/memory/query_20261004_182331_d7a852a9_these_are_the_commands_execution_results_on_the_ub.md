---
type: "query"
date: "2026-10-04T18:23:31.225166+00:00"
question: "These are the commands execution results on the Ubuntu server: Go proxy and checksum downloads reset before TLS, GitHub succeeds, cached Go container image is missing."
contributor: "graphify"
outcome: "useful"
source_nodes: ["Backend", "deploy/README.md", "blue-green.sh"]
---

# Q: These are the commands execution results on the Ubuntu server: Go proxy and checksum downloads reset before TLS, GitHub succeeds, cached Go container image is missing.

## Answer

Expanded graph tokens: backend build deploy. User-provided host probes show proxy.golang.org, goproxy.io, goproxy.cn, proxy.golang.com.cn, sum.golang.org, go.dev and dl.google.com reset before TLS with HTTP 000; GitHub git refs return 200; proxy environment unset. This is consistent with host egress/TLS filtering, not a Docker-only or Go compilation defect; exact resetting device is unknown. Container probe did not run because BuildKit cache does not guarantee a daemon-local golang image tag; operator should docker pull golang:1.27.1-alpine first. Added backend-only GOPROXY ARG with official default, deployment forwarding of set HTTP/HTTPS/NO_PROXY and lowercase equivalents by argument name, no checksum/TLS bypass or runtime dependencies. Updated deploy/README.md network requirements. All 13 database-free mocked rollout tests, Bash syntax, whitespace, editor diagnostics and added prose checks pass. Actual Docker build and network repair remain operator-only Ubuntu gates. Approved outbound proxy or HTTPS allowlisting for official Go endpoints and module ZIP storage still required.

## Outcome

- Signal: useful

## Source Nodes

- Backend
- deploy/README.md
- blue-green.sh