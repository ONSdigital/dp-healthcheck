module github.com/ONSdigital/dp-healthcheck

go 1.18

replace (
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d
	golang.org/x/text => golang.org/x/text v0.3.7
)

require (
	github.com/ONSdigital/log.go/v2 v2.0.9
	github.com/google/go-cmp v0.5.5
	github.com/smartystreets/goconvey v1.6.4
)

require (
	github.com/fatih/color v1.12.0 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20181017120253-0766667cb4d1 // indirect
	github.com/hokaccha/go-prettyjson v0.0.0-20210113012101-fb4e108d2519 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/smartystreets/assertions v0.0.0-20180927180507-b2de0cb4f26d // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
)
