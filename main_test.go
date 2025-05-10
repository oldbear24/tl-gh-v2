package main

import "flag"

var testLogFile string

func init() {
	flag.StringVar(&testLogFile, "testlogfile", "", "Path to test log file")
}
