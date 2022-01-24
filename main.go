package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	www  = flag.String("www", "www", "set target directory")
	src  = flag.String("src", "src/www", "set source directory")
	tmpl = flag.String("tmpl", "src/tmpl", "set template directory")
)

func main() {
	flag.Parse()

	fmt.Printf("generating website\n")

	err := process(*www, *src, *tmpl)
	if err != nil {
		log.Fatal(err)
	}
}

func process(www, src, tmpl string) error {
	infos, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, info := range infos {
		name := info.Name()
		src := filepath.Join(src, name)

		if info.IsDir() {
			www := filepath.Join(www, name)
			tmpl := filepath.Join(tmpl, name)

			err = process(www, src, tmpl)
			if err != nil {
				return err
			}
			continue
		}

		ext := filepath.Ext(name)
		switch ext {
		case ".md":
			fmt.Printf(" - %q\n", src)
		}
	}

	return nil
}
