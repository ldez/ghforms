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
	github.com/mattn/go-isatty v0.0.24
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.3
	github.com/urfave/cli/v3 v3.11.0
	github.com/yuin/goldmark v1.8.5
	github.com/zmtcreative/gm-alert-callouts v0.8.0
	gitlab.com/greyxor/slogor v1.7.0
	golang.org/x/text v0.41.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/jeandeaual/go-locale v0.0.0-20250612000132-0ef82f21eade // indirect
	golang.org/x/sys v0.47.0 // indirect
)
