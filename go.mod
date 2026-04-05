module github.com/The-Mess-NZ/midipunk

go 1.24.2

require (
	github.com/The-Mess-NZ/gooey v0.0.0
	gitlab.com/gomidi/midi/v2 v2.2.19
	go.bug.st/serial v1.6.4
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/llgcode/draw2d v0.0.0-20240627062922-0ed1ff131195 // indirect
	golang.org/x/image v0.26.0 // indirect
)

require (
	github.com/creack/goselect v0.1.3 // indirect
	github.com/gorilla/websocket v1.5.3
	golang.org/x/sys v0.32.0 // indirect
)

// TODO: This probably shouldn't be here for public release.
replace github.com/The-Mess-NZ/gooey => ../gooey
