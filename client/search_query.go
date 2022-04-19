package client

type Query interface {
	// Source returns the JSON-serializable query request
	Source() (interface{}, error)
}

type SearchQuery struct {
	Query
	allClauses  SearchQueries
	anyClauses  SearchQueries
	noneClauses SearchQueries
	fields      TermQueries
}

func NewSearchQuery() *SearchQuery {
	return &SearchQuery{}
}

func (q *SearchQuery) All(queries ...*SearchQuery) *SearchQuery {
	q.allClauses = append(q.allClauses, queries...)
	return q
}

func (q *SearchQuery) Any(queries ...*SearchQuery) *SearchQuery {
	q.anyClauses = append(q.anyClauses, queries...)
	return q
}

func (q *SearchQuery) None(queries ...*SearchQuery) *SearchQuery {
	q.noneClauses = append(q.noneClauses, queries...)
	return q
}

func (q *SearchQuery) Fields(queries TermQueries) *SearchQuery {
	q.fields = append(q.fields, queries...)
	return q
}

// root level query, generates `filter` object
func (q *SearchQuery) Source() (map[string]interface{}, error) {
	source := make(map[string]interface{})

	if q.allClauses != nil {
		as, err := q.allClauses.Source()
		if err != nil {
			return source, err
		}
		source["all"] = as
	}
	if q.anyClauses != nil {
		ac, err := q.anyClauses.Source()
		if err != nil {
			return source, err
		}
		source["any"] = ac
	}
	if q.noneClauses != nil {
		nc, err := q.noneClauses.Source()
		if err != nil {
			return source, err
		}
		source["none"] = nc
	}
	if q.fields != nil {
		fs, err := q.fields.Source()
		if err != nil {
			return source, err
		}
		source["fields"] = fs
	}

	return source, nil
}
