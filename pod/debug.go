package pod

import (

	"fmt"
	"os"
)

func Debug(v interface{}) {
	fmt.Fprintf(os.Stderr, "debug: %+v\n", v)
}
