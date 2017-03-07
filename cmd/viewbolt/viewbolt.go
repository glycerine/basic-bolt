package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/boltdb/bolt"
)

func main() {
	fl := flag.NewFlagSet("viewbolt", flag.ExitOnError)
	var cfg ViewBoltConfig
	cfg.DefineFlags(fl)
	err := fl.Parse(os.Args[1:])
	panicOn(err)
	err = cfg.ValidateConfig()
	panicOn(err)

	cfg.Bopt = &bolt.Options{Timeout: 1 * time.Second, ReadOnly: true}
	cfg.Db, err = bolt.Open(cfg.DbPath, 0600, cfg.Bopt)
	if err != nil {
		panic(err)
	}
	if cfg.JsonOut {
		cfg.WriteJson(cfg.Db, os.Stdout)
	} else {
		cfg.Dump(cfg.Db, os.Stdout)
	}
}

func (cfg *ViewBoltConfig) Dump(db *bolt.DB, w io.Writer) {
	fmt.Fprintf(w, "\n")
	i := 0
	const indent = "    "
	err := db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, buck *bolt.Bucket) error {
			nm := string(name)

			fmt.Fprintf(w, "\n* '%v' bucket:\n", nm)

			// list the keys/values in the bucket
			buck.ForEach(func(k, v []byte) error {
				fmt.Fprintf(w, "%s%v) '%v' -> %v\n",
					indent, i, string(k), string(v))
				if v == nil {
					sub := buck.Bucket(k)
					if sub != nil {
						j := 0
						sub.ForEach(func(kk, vv []byte) error {
							fmt.Fprintf(w, "%s%v) '%v' -> %v\n",
								indent+indent, j, string(kk), string(vv))
							j++
							return nil
						})
					}
				}
				i++
				return nil
			})

			return nil // keep iterating
		})
	})
	panicOn(err)
	//	items, err := b.bucket.Items()
}

func (cfg *ViewBoltConfig) WriteJson(db *bolt.DB, w io.Writer) {
	fmt.Fprintf(w, `{`)
	defer func() { fmt.Fprintf(w, `}`) }()

	top := 0
	topcomma := ""

	err := db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, buck *bolt.Bucket) error {
			nm := string(name)

			if top == 1 {
				topcomma = ","
			}

			// user home dir
			fmt.Fprintf(w, `%s"%s":{`, topcomma, nm)
			defer func() { fmt.Fprintf(w, `}`) }()
			m := 0
			comma := ""

			// list the keys/values in the bucket
			buck.ForEach(func(k, v []byte) error {
				if m == 1 {
					comma = ","
				}
				m++
				if v != nil {

					fmt.Fprintf(os.Stderr, `we see key-val where we expect sub-bucket only! %s"%v":"%v"`, comma, string(k), string(v))
				} else {
					sub := buck.Bucket(k)
					if sub != nil {
						fmt.Fprintf(w, `%s"%v":{`,
							comma, string(k))
						defer func() { fmt.Fprintf(w, `}`) }()

						j := 0
						subcomma := ""
						sub.ForEach(func(kk, vv []byte) error {
							if j == 1 {
								subcomma = ","
							}

							fmt.Fprintf(w, `%s"%v":%v`,
								subcomma, string(kk), string(vv))
							j++
							return nil
						})
					}
				}
				//i++
				return nil
			})

			top++
			return nil // keep iterating
		})
	})
	panicOn(err)
	//	items, err := b.bucket.Items()
}

func panicOn(err error) {
	if err != nil {
		panic(err)
	}
}

func FileExists(name string) bool {
	fi, err := os.Stat(name)
	if err != nil {
		return false
	}
	if fi.IsDir() {
		return false
	}
	return true
}

func BytesToInt64(by []byte) int64 {
	return int64(binary.LittleEndian.Uint64(by))
}
