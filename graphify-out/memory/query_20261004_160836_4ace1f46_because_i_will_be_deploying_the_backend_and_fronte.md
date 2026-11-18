---
type: "query"
date: "2026-10-04T16:08:36.763887+00:00"
question: "because I will be deploying the backend and frontend to the ubuntu server and the postgresql database is hosted on that server too, please help to update the environment files to setup the database configuration if it hosted on the ubuntu server (using localhost instead of server ip address). Please also create the nginx configurations for each backend (radom port) and frontend (random port). After that please help to update the backend and frontend README.md files on how to setup and run the projects locally and deployed in on the server."
contributor: "graphify"
outcome: "useful"
source_nodes: ["server/main.go", "PostgreSQL Cutover Procedure"]
---

# Q: because I will be deploying the backend and frontend to the ubuntu server and the postgresql database is hosted on that server too, please help to update the environment files to setup the database configuration if it hosted on the ubuntu server (using localhost instead of server ip address). Please also create the nginx configurations for each backend (radom port) and frontend (random port). After that please help to update the backend and frontend README.md files on how to setup and run the projects locally and deployed in on the server.

## Answer

Expanded graph vocabulary: postgres server static frontend; nginx vocabulary absent. Current source confirms DATABASE_URL, -addr, optional static SPA and same-origin API. Updated both env examples to quoted localhost:5432 PostgreSQL URLs with loopback-only sslmode=disable and verified-TLS alternatives. Added backend/nginx.conf on 39063 and frontend/nginx.conf on 31627, both proxying to Go at 127.0.0.1:24993; frontend serves /opt/web-program-functions/frontend/build with SPA fallback. Created backend README and replaced frontend scaffold README with local checks, operator-created API tunnel, private systemd environment, non-root service and Ubuntu deployment. Vite preview inherits the API proxy to local port 8080. Environment parsing, routing/port invariants, Markdown formatting/links, prose scan and editor checks passed. Nginx unavailable locally: actual nginx -t and port availability remain Ubuntu operator gates. No database connection, database tests, host access or deployment occurred; prior polish UI failure remains unresolved.

## Outcome

- Signal: useful

## Source Nodes

- server/main.go
- PostgreSQL Cutover Procedure