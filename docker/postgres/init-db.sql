SELECT 'CREATE DATABASE events_local_db'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'events_local_db')\gexec

SELECT 'CREATE DATABASE familytree_local_db'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'familytree_local_db')\gexec
