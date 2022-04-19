package client

type TermQueryRangeFilter struct {
	From interface{} `json:"from,omitempty"`
	To   interface{} `json:"to,omitempty"`
}

func NewTermQueryRangeFilter(from, to interface{}) (t TermQueryRangeFilter) {
	if from != nil {
		t.From = from
	}
	if to != nil {
		t.To = to
	}
	return
}
