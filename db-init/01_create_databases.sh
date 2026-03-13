#!/bin/bash
set -e

# videos_write is created automatically by POSTGRES_DB env var
# we just need to create videos_read
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE video_processor_replica;
    CREATE DATABASE video_processor;
EOSQL
