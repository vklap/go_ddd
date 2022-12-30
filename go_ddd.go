package go_ddd

import (
	"fmt"
	_ "github.com/vklap/go_ddd/pkg/ddd"
	"strings"
)

func main() {
	fmt.Println("Hello GO")
}

func ToUpper(s string) string {
	return strings.ToUpper(s)
}
