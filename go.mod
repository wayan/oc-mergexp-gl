module github.com/wayan/oc-mergexp-gl

go 1.24

// replace github.com/wayan/mergeexp => ../mergeexp

require (
	github.com/go-resty/resty/v2 v2.16.5
	github.com/urfave/cli/v3 v3.3.8
	github.com/wayan/mergeexp v0.6.0
)

require golang.org/x/net v0.33.0 // indirect
