package main

import (
	"go_sca/public"
)

var (
	Proname = "demo"

	Version = "0.1"
)

func init() {
	public.Proname = Proname
}
