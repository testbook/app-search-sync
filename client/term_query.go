package client

type TermQuery struct {
	Key   string
	Value interface{} // make value specific to type for date, location etc
}

func NewTermQuery(key string, value interface{}) TermQuery {
	return TermQuery{
		Key:   key,
		Value: value,
	}
}
