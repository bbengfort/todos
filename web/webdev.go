// +build dev

package web

import "net/http"

// WebApp is an http filesystem that can be used to serve static assets using the todos
// server. The web application is located in the same directory as the go module or in
// the case of a compiled/bundled web app, should be bundled to this directory. When the
// go generate command is called, the file system is statically embedded into the server
// binary. This increases the size of the final binary, but prevents the need to use
// nginx with proxy routes or a multi image system.
var WebApp http.FileSystem = http.Dir("app")
