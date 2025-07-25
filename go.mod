module github.com/getsavvyinc/upgrade-cli

go 1.21.6

require (
	github.com/hashicorp/go-version v1.6.0
	github.com/stretchr/testify v1.8.4
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

retract v0.7.0 // missing fallback for arm64 -> all
