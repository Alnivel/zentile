module github.com/Alnivel/zentile

go 1.24

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/BurntSushi/xgb v0.0.0-20160522181843-27f122750802
	github.com/BurntSushi/xgbutil v0.0.0-20160919175755-f7c97cef3b4e
	github.com/mitchellh/go-homedir v1.1.0
	github.com/sirupsen/logrus v1.4.2
)

require (
	github.com/konsorten/go-windows-terminal-sequences v1.0.1 // indirect
	golang.org/x/sys v0.0.0-20190422165155-953cdadca894 // indirect
)

replace (
	github.com/BurntSushi/xgb => github.com/jezek/xgb v0.0.0-20160522181843-27f122750802
	github.com/BurntSushi/xgbutil => github.com/jezek/xgbutil v0.0.0-20160919175755-f7c97cef3b4e
)
