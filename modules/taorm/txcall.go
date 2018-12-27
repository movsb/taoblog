package taorm

// TxCall calls callback within transaction.
// It automatically catches and re-throws exceptions.
func (db *DB) TxCall(callback func(tx *DB)) {
	tx, err := db.rdb.Begin()
	if err != nil {
		panic(err)
	}

	catchCall := func() (except interface{}) {
		defer func() {
			except = recover()
		}()
		callback(NewTx(tx))
		return
	}

	if except := catchCall(); except != nil {
		tx.Rollback()
		panic(except)
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		panic(err)
	}
}
