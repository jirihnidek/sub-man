module sub-man

go 1.18

require (
	github.com/creasty/defaults v1.7.0
	github.com/urfave/cli/v2 v2.23.6
	gopkg.in/ini.v1 v1.67.0
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/henvic/httpretty v0.1.2 // indirect
	github.com/jirihnidek/rhsm2 v0.0.0-20231006102558-6ef3aaf7e314 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/rs/zerolog v1.30.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	golang.org/x/sys v0.12.0 // indirect
)

replace github.com/jirihnidek/rhsm2 => ../rhsm2

replace github.com/henvic/httpretty => ../forks/httpretty
