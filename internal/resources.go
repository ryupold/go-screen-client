package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func main() {
	compileResources(
		"resources.go",
		"icon_online.ico",
		"icon_offline.ico",
		"icon_no_server.ico",
	)
}

func compileResources(destination string, resourceFiles ...string) {
	resources, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer resources.Close()

	var resourceList = make(map[string]string)

	resources.WriteString("package main\n\n")

	for _, f := range resourceFiles {
		data, err := ioutil.ReadFile(f)
		if err != nil {
			fmt.Println("cannot read file:", f)
			panic(err)
		}
		withoutExtension := strings.Replace(path.Base(f), path.Ext(f), "", 1)
		ext := strings.Replace(path.Ext(f), ".", "", -1)
		normalizedFileName := "bin" + strings.ToUpper(ext)
		nextCharIsBig := true
		for _, c := range withoutExtension {
			if (c < 'A' || c > 'Z') && (c < 'a' || c > 'z') {
				nextCharIsBig = true
			} else if nextCharIsBig {
				normalizedFileName += strings.ToUpper(string(c))
				nextCharIsBig = false
			} else {
				normalizedFileName += strings.ToLower(string(c))
			}
		}
		for _, exist := resourceList[normalizedFileName]; exist; _, exist = resourceList[normalizedFileName] {
			normalizedFileName += "Copy"
		}
		resourceList[normalizedFileName] = f
		fmt.Printf("\t+ resource %s -> %s (%d bytes)\n", normalizedFileName, f, len(data))

		resources.WriteString(fmt.Sprintf("var %s = ", normalizedFileName))
		resources.WriteString("[]byte{")

		for i, b := range data {
			resources.WriteString(fmt.Sprintf("0x%x", b))
			if i < len(data)-1 {
				resources.WriteString(",")
			}
		}
		resources.WriteString("}\n\n")
		if err = resources.Sync(); err != nil {
			panic(err)
		}
	}
}
