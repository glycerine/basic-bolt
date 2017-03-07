package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"time"
)

func panicOn(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	b, err := newBoltdb("my.db")
	panicOn(err)
	defer b.Close()
	err = b.writesnap([]byte("hellow"))
	panicOn(err)
	r, err := b.readsnap()
	panicOn(err)
	fmt.Printf("r = '%v'\n", string(r))
}

var chk = []byte("checkpoint")
var snap = []byte("snapshot")

type boltdb struct {
	db       *bolt.DB
	filepath string
}

func (b *boltdb) Close() {
	if b != nil && b.db != nil {
		b.db.Close()
	}
}

func newBoltdb(filepath string) (*boltdb, error) {

	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.
	db, err := bolt.Open(filepath, 0600, &bolt.Options{Timeout: 2 * time.Second})
	if err != nil {
		// probably already open by another process.
		return nil, fmt.Errorf("error opening boltdb,"+
			" in use by other process? error detail: '%v'", err)
	}

	if err != nil {
		return nil, err
	}

	return &boltdb{
		db:       db,
		filepath: filepath,
	}, nil
}

func (b *boltdb) readsnap() (res []byte, err error) {

	err = b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(chk)
		v := b.Get(snap)

		// values returned from Get() are only
		// valid while the transaction is open.
		// If you need to use a value outside of
		// the transaction then you must use
		// copy() to copy it to another byte slice.

		res = make([]byte, len(v))
		copy(res, v)
		return nil

	})
	return
}

func (b *boltdb) writesnap(state []byte) error {

	return b.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(chk)
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		err = b.Put(snap, state)
		return err
	})

}
