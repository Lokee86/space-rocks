//go:build !localpackage

package main

func runtimePath(relativePath string) string {
	return relativePath
}
