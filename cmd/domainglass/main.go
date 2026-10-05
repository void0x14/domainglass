// Package main, domainglass komut satırı aracının giriş noktasıdır.
//
// Tüm mantık internal/cli paketindedir; bu dosya yalnızca süreci başlatır.
package main

import (
	"os"

	"github.com/void0x14/domainglass/internal/cli"
)

func main() {
	os.Exit(cli.Calistir(os.Args[1:], os.Stdout, os.Stderr))
}
