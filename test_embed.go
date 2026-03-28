package main

import (
	"embed"
	"fmt"
)

//go:embed all:web/backend/dist
var FS embed.FS

func main() {
	fmt.Println("embed test")
}
