package unsafe

// Suppress skips error checking. Make sure that error does not really matter
func Suppress[T any](item T, _ error) T { return item }

// Must panics if any error happens. Make sure that error must really critical
func Must[T any](item T, err error) T {
	if err != nil {
		panic(err)
	}

	return item
}
