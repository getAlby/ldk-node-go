package ldk_node

/*
#cgo LDFLAGS: -lldk_node
#cgo linux,amd64 LDFLAGS: -Wl,-rpath,${SRCDIR} -L${SRCDIR}
#cgo darwin,arm64 LDFLAGS: -Wl,-rpath,${SRCDIR} -L${SRCDIR}
#cgo darwin,amd64 LDFLAGS: -Wl,-rpath,${SRCDIR} -L${SRCDIR}
*/
import "C"
