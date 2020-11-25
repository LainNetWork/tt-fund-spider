package main

import "flag"

func main() {
	var code string
	flag.StringVar(&code,"code","008282","输入基金编号")
	flag.Parse()
	FetchFundAllNetValue(code)
}
