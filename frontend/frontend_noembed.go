//go:build !embed

package frontend

import "os"

func init() {
	FrontendFS = os.DirFS("./dist")
}
