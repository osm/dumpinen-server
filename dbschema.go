package main

import (
	"github.com/osm/migrator/repository"
)

// getDatabaseRepository returns the migrations for the database.
func getDatabaseRepository() repository.Source {
	return repository.FromMemory(map[int]string{
		1: `
			CREATE TABLE migration (
				version int NOT NULL PRIMARY KEY
			);
		`,
		2: `
			CREATE TABLE dump (
				id uuid NOT NULL PRIMARY KEY,
				public_id text NOT NULL,
				filesystem_id uuid NOT NULL,
				content_type text NOT NULL,
				ip_address text NOT NULL,
				delete_after timestamptz NOT NULL,
				encrypted_username bytea DEFAULT NULL,
				encrypted_password bytea DEFAULT NULL,
				deleted_at timestamptz DEFAULT NULL,
				inserted_at timestamptz  DEFAULT transaction_timestamp() NOT NULL
			);
			CREATE INDEX dump_delete_after ON dump(delete_after);
			CREATE INDEX dump_public_id_idx ON dump(public_id);
			CREATE UNIQUE INDEX dump_filesystem_id_uniq_idx ON dump(filesystem_id);
			CREATE UNIQUE INDEX dump_public_id_uniq_idx ON dump(public_id);

			CREATE TABLE dump_access_log (
				id uuid NOT NULL PRIMARY KEY,
				dump_id uuid NOT NULL REFERENCES dump(id),
				ip_address text NOT NULL,
				inserted_at timestamptz  DEFAULT transaction_timestamp() NOT NULL
			);
		`,
	})
}
