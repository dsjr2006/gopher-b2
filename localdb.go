package gopherb2

import (
	"github.com/boltdb/bolt"
	"github.com/uber-go/zap"
	"time"
)

type boltDB struct {
	*bolt.DB
}

// openDB opens a database.
func openDB(file string) (*boltDB, error) {
	db, err := bolt.Open(file, 0600, &bolt.Options{Timeout: 2 * time.Second})
	if err != nil {
		logger.Warn("Could not open boltdb file",
			zap.Error(err),
		)
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("checksums"))
		return err
	})
	return &boltDB{DB: db}, err
}

/*
boltdb, err := bolt.Open("gopherb2.db", 0600, nil)
if err != nil {
    logger.Fatal("Could not open boltdb",
        zap.Error(err),
    )
}
defer boltdb.Close()
*/
