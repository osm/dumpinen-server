package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/osm/migrator"
)

// db contains the database connection and holds the database related methods.
type db struct {
	conn *sql.DB
}

// dump is a model of the dump table.
type dump struct {
	id           string
	publicID     string
	filesystemID string
	contentType  string
	insertedAt   string
	ipAddress    string
	deleteAfter  time.Time
	username     *[]byte
	password     *[]byte
	deletedAt    *string
}

// dumpAccessLog is a model of the dump_access_log table.
type dumpAccessLog struct {
	id         string
	dumpID     string
	ipAddress  string
	insertedAt string
}

// dumpInfo is the model that holds some basic info about a dump.
type dumpInfo struct {
	createdAt time.Time
	count     int
}

// newDB initializes a new database connection with the provided connection
// string.
func newDB(cs string) (*db, error) {
	var conn *sql.DB
	var err error
	if conn, err = sql.Open("postgres", cs); err != nil {
		return nil, fmt.Errorf("can't initialize database connection: %w", err)
	}

	if err = migrator.ToLatest(conn, getDatabaseRepository()); err != nil {
		return nil, err
	}

	return &db{conn}, nil
}

// insertDump inserts a new dump to the database.
func (d *db) insertDump(du *dump) error {
	query := `INSERT INTO dump (
		id,
		filesystem_id,
		public_id,
		content_type,
		ip_address,
		encrypted_username,
		encrypted_password,
		delete_after
	) VALUES (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		$8
	);`
	stmt, err := d.conn.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		newUUID(),
		du.filesystemID,
		du.publicID,
		du.contentType,
		du.ipAddress,
		du.username,
		du.password,
		du.deleteAfter,
	)
	if err != nil {
		return err
	}

	return nil
}

// getDumpByPublicID fetches the given file by the public id from the
// database.
func (d *db) getDumpByPublicID(publicID string) (*dump, error) {
	var du dump
	query := `SELECT
		id,
		content_type,
		filesystem_id,
		encrypted_username,
		encrypted_password,
		deleted_at
	FROM dump
	WHERE
		public_id = $1`
	err := d.conn.QueryRow(query, publicID).
		Scan(
			&du.id,
			&du.contentType,
			&du.filesystemID,
			&du.username,
			&du.password,
			&du.deletedAt,
		)
	if err != nil {
		return nil, err
	}

	return &du, nil
}

// getFilesystemIDsToDelete returns a slice of filesystem ids that is up for
// deletion.
func (d *db) getFilesystemIDsToDelete() ([]string, error) {
	query := `SELECT
		filesystem_id
	FROM dump
	WHERE
		deleted_at IS NULL
		AND delete_after <> '0001-01-01 01:12:12+01:12:12'
		AND now() > delete_after;`

	rows, err := d.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var filesystemIDs []string
	for rows.Next() {
		var filesystemID string
		err = rows.Scan(&filesystemID)
		if err != nil {
			return nil, err
		}

		filesystemIDs = append(filesystemIDs, filesystemID)
	}

	return filesystemIDs, nil
}

// deleteDumpByFilesystemID deletes the file by the given filesystem id.
func (d *db) deleteDumpByFilesystemID(filesystemID string) error {
	query := "UPDATE dump SET deleted_at = now() WHERE filesystem_id = $1"
	stmt, err := d.conn.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(filesystemID)
	if err != nil {
		return err
	}

	return nil
}

// insertDumpAccessLog inserts a new entry to the dump access log.
func (d *db) insertDumpAccessLog(dal *dumpAccessLog) error {
	query := `INSERT INTO dump_access_log (
		id,
		dump_id,
		ip_address
	) VALUES (
		$1,
		$2,
		$3
	);`
	stmt, err := d.conn.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(newUUID(), dal.dumpID, dal.ipAddress)
	if err != nil {
		return err
	}

	return nil
}

func (d *db) getDumpInfoByPublicID(publicID string) (*dumpInfo, error) {
	var di dumpInfo

	query := `SELECT inserted_at FROM dump WHERE public_id = $1`
	if err := d.conn.QueryRow(query, publicID).Scan(&di.createdAt); err != nil {
		return nil, err
	}

	query = `SELECT COUNT(1) FROM dump_access_log a INNER JOIN dump d ON d.id = a.dump_id WHERE d.public_id = $1`
	if err := d.conn.QueryRow(query, publicID).Scan(&di.count); err != nil {
		return nil, err
	}

	return &di, nil
}
