---
type: "query"
date: "2026-10-04T17:57:46.872914+00:00"
question: "Please check why I got this issue on the server: Protect deploy/.env with chmod 600."
contributor: "graphify"
outcome: "useful"
source_nodes: ["blue-green.sh", "deploy/README.md"]
---

# Q: Please check why I got this issue on the server: Protect deploy/.env with chmod 600.

## Answer

Expanded from graph vocab: deploy blue green secrets. Graph located deploy/blue-green.sh and deploy/README.md; source verification found exact guard at deploy/blue-green.sh L153: stat mode must equal 600. Script changes directory to deploy first. chmod +x deploy/blue-green.sh changes only script permissions, not deploy/.env. Synthetic Bash check rejects 644 and 640 and accepts 600. User should run chmod 600 deploy/.env on approved Ubuntu host, verify stat metadata only, then rerun release command. Mode 600 protects credentials with owner read and write only. No code fix or host/database access performed; actual server mode not inspected.

## Outcome

- Signal: useful

## Source Nodes

- blue-green.sh
- deploy/README.md