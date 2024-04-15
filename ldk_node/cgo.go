package ldk_node

/*
#cgo LDFLAGS: -lldk_node
#cgo linux,amd64 LDFLAGS: -Wl,-rpath,${SRCDIR}/x86_64-unknown-linux-gnu -L${SRCDIR}/x86_64-unknown-linux-gnu
#cgo darwin LDFLAGS: -Wl,-rpath,${SRCDIR}/universal-macos -L${SRCDIR}/universal-macos
#cgo windows,amd64 LDFLAGS: -Wl,-rpath,${SRCDIR}/x86_64-pc-windows-gnu -L${SRCDIR}/x86_64-pc-windows-gnu
#cgo linux,arm.6 LDFLAGS: -Wl,-rpath,${SRCDIR}/arm-unknown-linux-gnueabi -L${SRCDIR}/arm-unknown-linux-gnueabi
*/
import "C"
