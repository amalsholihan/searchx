package searchx

func (ks *Searchx) Get() *Searchx {
	ks.SetRawQuery()
	ks.Parse()
	ks.ParseSelectMapping()
	ks.ProcessSearch()
	ks.ProcessUnion()
	ks.ParseCountQuery()
	ks.ParseCurrentPageQuery(1, 15)

	return ks
}
