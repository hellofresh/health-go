#!/bin/bash
set -e

mysql -v -uroot -p${MYSQL_ROOT_PASSWORD} <<-EOSQL
USE ${MYSQL_DATABASE};

DROP TABLE IF EXISTS test;

CREATE TABLE IF NOT EXISTS test (
  id           INT NOT NULL AUTO_INCREMENT PRIMARY KEY ,
  secret       VARCHAR(256) NOT NULL,
  extra        VARCHAR(256) NOT NULL,
  redirect_uri VARCHAR(256) NOT NULL
);
EOSQL
