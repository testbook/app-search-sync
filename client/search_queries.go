package client

type SearchQueries []*SearchQuery

// generates nested objects within `filter`, returns array of objects
func (s *SearchQueries) Source() (r []map[string]interface{}, err error) {
	r = make([]map[string]interface{}, 0)

	for _, q := range *s {
		if len(q.allClauses) > 0 {
			as, err := q.allClauses.Source()
			if err != nil {
				return r, err
			}
			if len(as) > 0 {
				r = append(r, map[string]interface{}{"all": as})
			}
		}
		if len(q.anyClauses) > 0 {
			as, err := q.anyClauses.Source()
			if err != nil {
				return r, err
			}
			if len(as) > 0 {
				r = append(r, map[string]interface{}{"any": as})
			}
		}
		if len(q.noneClauses) > 0 {
			as, err := q.noneClauses.Source()
			if err != nil {
				return r, err
			}
			if len(as) > 0 {
				r = append(r, map[string]interface{}{"none": as})
			}
		}
		if len(q.fields) > 0 {
			f, err := q.fields.Source()
			if err != nil {
				return r, err
			}
			r = append(r, f...)
		}
	}

	return
}
