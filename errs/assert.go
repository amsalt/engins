package errs

func Assert(err error) {
	if err != nil {
		panic(err)
	}
}
