package pointers

import "time"

func String(s string) *string                       { return &s }
func Int(i int) *int                                { return &i }
func Time(t time.Time, e error) (*time.Time, error) { return &t, e }
