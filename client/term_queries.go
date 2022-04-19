package client

type TermQueries []TermQuery

func NewTermQueries(queries ...TermQuery) TermQueries {
	t := TermQueries{}
	if len(queries) > 0 {
		t = queries
	}
	return t
}

func (t *TermQueries) Add(queries ...TermQuery) *TermQueries {
	*t = append(*t, queries...)
	return t
}

func (t *TermQueries) Source() (s []map[string]interface{}, err error) {
	s = make([]map[string]interface{}, 0)

	for _, q := range *t {
		if q.Key == "" {
			continue
		}
		s = append(s, map[string]interface{}{q.Key: q.Value})
	}
	return s, nil
}
