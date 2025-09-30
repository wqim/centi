package util
import (
	"net/url"
	"database/sql"
	_ "github.com/xeodou/go-sqlcipher"
	
	"centi/cryptography"
)

/*
 * a database orm for easy packet logging
 */
type DB struct {
	db		*sql.DB
	rowsLimit	uint
}

func ConnectDB( filename, password string, rowsLimit uint ) (*DB, error) {

	dbFilename := "file:" + url.QueryEscape( filename )
	dbFilename += "?_journal_mode=WAL&_key=" + url.QueryEscape( password )

	db, err := sql.Open( "sqlite3", dbFilename )
	if err != nil {
		return nil, err
	}
	final := &DB {
		db,
		rowsLimit,
	}
	// check amount of rows
	rows, err := final.Count()
	if err == nil {
		if uint(rows) > rowsLimit {
			ShredFile( filename )
			// call recursively, as it won't go on the depth more than 1
			return ConnectDB( filename, password, rowsLimit )
		}
	} // else the database was not existing before
	return final, nil
}

func(db *DB) Close() {
	db.db.Close()
}

func(db *DB) InitDB() error {
	// create the table itself.
	sqlStmt := `create table if not exists packets(id integer not null primary key autoincrement, hash text);`
	_, err := db.db.Exec( sqlStmt )
	if err != nil {
		return err
	}
	// add indexation in order to optimize database search
	_, err = db.db.Exec(`create index if not exists hashIdx on packets(hash);`)
	return err
}

// add packet into database
func(db *DB) AddPacket( data []byte ) error {
	
	hash := cryptography.Hash( data )
	_, err := db.db.Exec("insert into packets(hash) values(?);", hash)
	return err
}

// checks if packet is already sent
func(db *DB) IsInDB( data []byte ) (bool, error) {

	hash := cryptography.Hash( data )
	stmt, err := db.db.Prepare(`select * from packets where hash = ?;`)
	if err != nil {
		return false, err
	}
	defer stmt.Close()
	rows, err := stmt.Query( hash )
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		// if there is any row, packet already was resent.
		return true, nil
	}
	return false, nil
}

func(db *DB) Count() (int, error) {
	// return amount of peers in the database
	rows, err := db.db.Query(`select count(*) from packets;`)
	if err != nil {
		return -1, err
	}
	defer rows.Close()
	for rows.Next() {
		var amount int
		err = rows.Scan( &amount )
		if err != nil {
			return -1, err
		}
		return amount, nil
	}
	// this is just in case. in normal conditions,
	// the execution does not go here...
	err = rows.Err()
	if err != nil {
		return -1, err
	}
	return 0, nil
}
