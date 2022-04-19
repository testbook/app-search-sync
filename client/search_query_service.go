package client

type SearchQueryService struct {
	Term        string
	SearchQuery *SearchQuery
	PageQuery   *PageQuery
}

// https://www.elastic.co/guide/en/app-search/current/filters.html#filters-nesting-filters
func NewSearchQueryService(term string, query *SearchQuery, pageQuery *PageQuery) *SearchQueryService {
	return &SearchQueryService{
		Term:        term,
		SearchQuery: query,
		PageQuery:   pageQuery,
	}
}

func (s *SearchQueryService) Source() (r map[string]interface{}, err error) {
	r = make(map[string]interface{})
	r["query"] = s.Term

	qs, err := s.SearchQuery.Source()
	if err != nil {
		return
	}
	if qs != nil {
		r["filters"] = qs
	}

	pqs, err := s.PageQuery.Source()
	if err != nil {
		return
	}
	if pqs != nil {
		r["page"] = pqs
	}

	return
}
