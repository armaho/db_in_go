package db_in_go

func check(cond bool) {
	if !cond {
		panic("assertion failure")
	}
}
