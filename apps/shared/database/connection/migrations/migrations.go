package migrations

import "fmt"

type Migration struct {
	Version int
	Name    string
	Up      string
	Down    string
}

func GetMigrationsForWriteInstance() []Migration {
	return []Migration{
		{
			Version: 1,
			Name:    "create_videos_table",
			Up: `
			CREATE TABLE IF NOT EXISTS videos (
			    id SERIAL PRIMARY KEY,
			    user_id TEXT NOT NULL,
			    name TEXT NOT NULL,
			    processed_at TIMESTAMP NOT NULL DEFAULT NOW(),
			    uploaded_at TIMESTAMP NOT NULL DEFAULT NOW(),
			    path TEXT NOT NULL
			    );`,
			Down: `DROP TABLE IF EXISTS videos;`,
		},
	}
}

func GetMigrationsForReadInstance() []Migration {
	return []Migration{
		{
			Version: 1,
			Name:    "create_videos_table",
			Up: `
        CREATE TABLE IF NOT EXISTS videos (
            id           INT PRIMARY KEY,
            user_id      TEXT NOT NULL,
            name         TEXT NOT NULL,
            processed_at TIMESTAMP NOT NULL,
            uploaded_at  TIMESTAMP NOT NULL,
            path         TEXT NOT NULL
        );
        CREATE INDEX IF NOT EXISTS idx_videos_user_id ON videos (user_id);`,
			Down: `DROP TABLE IF EXISTS videos;`,
		},
		{
			Version: 2,
			Name:    "create_sync_state_table",
			Up: `
        CREATE EXTENSION IF NOT EXISTS dblink;
        CREATE EXTENSION IF NOT EXISTS pg_cron;

        CREATE TABLE IF NOT EXISTS sync_state (
            key            VARCHAR(100) PRIMARY KEY,
            last_synced_at TIMESTAMP NOT NULL DEFAULT '1970-01-01'
        );

        INSERT INTO sync_state (key) VALUES ('videos_sync')
        ON CONFLICT (key) DO NOTHING;`,
			Down: `
        DROP TABLE IF EXISTS sync_state;
        DROP EXTENSION IF EXISTS pg_cron;
        DROP EXTENSION IF EXISTS dblink;`,
		},
		{
			Version: 3,
			Name:    "create_sync_function_and_cron_job",
			Up: `
        CREATE OR REPLACE FUNCTION sync_videos()
        RETURNS void AS $$
        DECLARE
            last_sync  TIMESTAMP;
            write_conn TEXT := current_setting('app.write_db_conn');
        BEGIN
            SELECT last_synced_at INTO last_sync
            FROM sync_state WHERE key = 'videos_sync';

            INSERT INTO videos (id, user_id, name, processed_at, uploaded_at, path)
            SELECT id, user_id, name, processed_at, uploaded_at, path
            FROM dblink(write_conn,
                format(
                    'SELECT id, user_id, name, processed_at, uploaded_at, path
                     FROM videos
                     WHERE processed_at > %L',
                    last_sync
                )
            ) AS t(
                id INT, user_id TEXT, name TEXT,
                processed_at TIMESTAMP, uploaded_at TIMESTAMP, path TEXT
            )
            ON CONFLICT (id) DO UPDATE SET
                user_id      = EXCLUDED.user_id,
                name         = EXCLUDED.name,
                processed_at = EXCLUDED.processed_at,
                path         = EXCLUDED.path;

            UPDATE sync_state
            SET last_synced_at = NOW()
            WHERE key = 'videos_sync';

            RAISE LOG 'videos sync completed at %', NOW();
        END;
        $$ LANGUAGE plpgsql;

        SELECT cron.schedule('sync-videos', '30 seconds', 'SELECT sync_videos()');`,
			Down: `
        SELECT cron.unschedule('sync-videos');
        DROP FUNCTION IF EXISTS sync_videos();`,
		},
	}
}

func CreateConnectionVar(host, writeDBName, readDBName, user, password string, port int) string {
	s := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s", host, port, writeDBName, user, password)

	return fmt.Sprintf("ALTER DATABASE %s SET app.write_db_conn ='%s';", readDBName, s)
}
