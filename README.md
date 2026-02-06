Argus – Local Infrastructure (Sprint 1)
Prerequisites

Docker Desktop installed and running

Docker Compose available (docker compose)

How to Run
1-Run Docker Desktop on your computer
cd infra
docker compose up -d

Check containers:

docker compose ps

Verify RabbitMQ

URL: http://localhost:15672

Verify Postgres
docker exec -it $(docker compose ps -q db) psql -U argus -d argus

SELECT 1;
\q

Stop
docker compose down
