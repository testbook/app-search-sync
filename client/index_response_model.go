package client

type IndexResponse struct {
	Id     string        `json:"id"`
	Errors []interface{} `json:"errors"`
}

type BulkIndexResponse []IndexResponse

func (b BulkIndexResponse) Errors() []interface{} {
	var errs []interface{}
	for _, r := range b {
		errs = append(errs, r.Errors...)
	}
	return errs
}
