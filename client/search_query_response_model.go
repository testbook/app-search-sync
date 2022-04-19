package client

import "github.com/testbook/tbmongo/tbbson"

type MetaEngineResponse struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type MetaPageResponse struct {
	Current      int `json:"current"`
	Size         int `json:"size"`
	TotalPages   int `json:"total_pages"`
	TotalResults int `json:"total_results"`
}

type SearchQueryResponseMeta struct {
	Alerts    []interface{}      `json:"alerts"`
	Precision int                `json:"precision"`
	Engine    MetaEngineResponse `json:"engine"`
	Page      MetaPageResponse   `json:"page"`
	RequestId string             `json:"request_id"`
	Warnings  []interface{}      `json:"warnings"`
}

type SearchQueryResultMeta struct {
	Engine string  `json:"engine"`
	Id     string  `json:"id"`
	Score  float64 `json:"score"`
}

type SearchQueryRawResponse struct {
	Raw string `json:"raw"`
}

type SearchQueryResults struct {
	Meta SearchQueryResultMeta `json:"_meta"`
}

type SearchQueryResponse struct {
	Errors  []string                `json:"errors"`
	Meta    SearchQueryResponseMeta `json:"meta"`
	Results []SearchQueryResults    `json:"results"`
}

func (s *SearchQueryResponse) GetObjectIds() (oids []tbbson.ObjectID) {
	for _, sqr := range s.Results {
		oids = append(oids, tbbson.ObjectIdHex(sqr.Meta.Id))
	}
	return
}
