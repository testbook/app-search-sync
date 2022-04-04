package main

import (
	"fmt"
	"strconv"
)

type resumeStrategy int

const (
	timestampResumeStrategy resumeStrategy = iota
	tokenResumeStrategy
)

func (arg *resumeStrategy) String() string {
	return fmt.Sprintf("%d", *arg)
}

func (arg *resumeStrategy) Set(value string) (err error) {
	var i int
	if i, err = strconv.Atoi(value); err != nil {
		return
	}
	rs := resumeStrategy(i)
	*arg = rs
	return
}
