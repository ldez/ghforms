module github.com/ldez/ghforms

go 1.26.0

ignore (
	./.github
	./docs
	./internal/render/src
	./node_modules
)

require (
	github.com/fsnotify/fsnotify v1.10.1
	github.com/mattn/go-isatty v0.0.22
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2
	github.com/urfave/cli/v3 v3.10.0
	github.com/yuin/goldmark v1.8.2
	github.com/zmtcreative/gm-alert-callouts v0.8.0
	gitlab.com/greyxor/slogor v1.6.10
	golang.org/x/text v0.38.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/jeandeaual/go-locale v0.0.0-20250612000132-0ef82f21eade // indirect
	golang.org/x/sys v0.45.0 // indirect
)
