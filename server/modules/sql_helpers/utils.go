package sql_helpers

// createSQLInMarks creates "?,?,?" string.
func createSQLInMarks(count int) string {
	s := "?"
	for i := 1; i < count; i++ {
		s += ",?"
	}
	return s
}
