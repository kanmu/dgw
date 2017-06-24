package dgwexample

//go:generate dgw postgres://dgw_test@localhost/dgw_test?sslmode=disable --typemap=./typemap.toml --schema=public --package=dgwexample --output=defaultstruct.go
//go:generate dgw postgres://dgw_test@localhost/dgw_test?sslmode=disable --typemap=./typemap.toml --schema=public --package=dgwexample --output=customstruct.go --template=./custom.tmpl
