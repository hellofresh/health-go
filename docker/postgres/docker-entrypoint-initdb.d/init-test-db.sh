#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
CREATE TABLE IF NOT EXISTS test (
  id           TEXT NOT NULL PRIMARY KEY,
  secret       TEXT NOT NULL,
  extra        TEXT NOT NULL,
  redirect_uri TEXT NOT NULL
);
EOSQL
