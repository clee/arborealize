package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type dir struct {
	name string
	path string
	files []string
	subdirs []dir
}

const htmlHeader = `<html>
<head>
<title>%s</title>
<style type="text/css">
ol.tree { padding: 0 0 0 30px; width: 300px; }
li { position: relative; margin-left: -15px; list-style: none; }
li.file { margin-left: -1px !important; }
li.file a { background: url(document.png) 0 0 no-repeat; color: #fff; padding-left: 21px; text-decoration: none; display: block; }
li input { position: absolute; left: 0; margin-left: 0; opacity: 0; z-index: 2; cursor: pointer; height: 1em; width: 1em; top: 0; }
li input + ol { background: url(toggle-small-expand.png) 40px 0 no-repeat; margin: -0.938em 0 0 -44px; /* 15px */ height: 1em; }
li input + ol > li { display: none; margin-left: -14px !important; padding-left: 1px; }
li label { background: url(folder-horizontal.png) 15px 1px no-repeat; cursor: pointer; display: block; padding-left: 37px; }
li input:checked + ol { background: url(toggle-small.png) 40px 5px no-repeat; margin: -1.25em 0 0 -44px; padding: 1.563em 0 0 80px; height: auto; }
li input:checked + ol > li { display: block; margin: 0 0 0.125em; }
li input:checked + ol > li:last-child { margin: 0 0 0.063em; }
</style>
<body>
`
const htmlFooter = `</body>
</html>`

func subdirIndex(subdirs []dir, name string) int {
	for i, d := range subdirs {
		if d.name == name {
			return i
		}
	}
	return -1
}

func markupFromTree(tree dir) (ret string) {
	name := tree.name
	if tree.name == "" {
		ret = "<ol class=\"tree\">\n<li>\n"
		name = "/"
	} else {
		ret = "<ol>\n<li>\n"
	}
	ret += fmt.Sprintf("<label for=\"%s\">%s</label>\n", name, name)
	if len(tree.subdirs) > 0 {
		for _, s := range tree.subdirs {
			ret += markupFromTree(s)
		}
	}
	if len(tree.files) > 0 {
		ret += "<ol>\n"
		for _, f := range tree.files {
			ret += fmt.Sprintf("<li class=\"file\"><a href=\"%s%s\">%s</a></li>\n", tree.path, f, f)
		}
		ret += "</ol>\n"
	}
	ret += "</li>\n</ol>\n"
	return
}

func treeFromFiles(files map[string][]string) (dir, error) {
	root := dir{name: "", files: files[""], subdirs: []dir{}}
	keys := make([]string, len(files))
	for key, _ := range files {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		var currentDir *dir = &root
		var newDir *dir
		subdirNames := strings.Split(key, "/")
		for i, d := range subdirNames {
			// Skip empty post-trailing-/ string
			if i == len(subdirNames) -1 {
				continue
			}
			if j := subdirIndex(currentDir.subdirs, d); j == -1 {
				newDir = new(dir)
				newDir.name = d
				newDir.path = strings.Join(subdirNames[0:i+1], "/") + "/"
				newDir.files = files[newDir.path]
				currentDir.subdirs = append(currentDir.subdirs, (*newDir))
			} else {
				newDir = &currentDir.subdirs[j]
			}
			currentDir = newDir
		}
	}
	return root, nil
}

func main() {
	var root string
	var err error
	files := make(map[string][]string)

	if root, err = os.Getwd(); err != nil {
		fmt.Printf("can't get the current path\n")
		panic(err)
	}

	flag.StringVar(&root, "root", root, "filesystem root for scan")
	flag.Parse()
	if !strings.HasSuffix(root, "/") {
		root = root + "/"
	}

	filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		path = strings.TrimPrefix(path, root)
		dir := strings.TrimSuffix(path, f.Name())
		if f.Mode().IsRegular() {
			files[dir] = append(files[dir], f.Name())
		}
		return nil
	})

	f, err := treeFromFiles(files)
	html := fmt.Sprintf(htmlHeader, root)
	html += markupFromTree(f)
	html += htmlFooter
	fmt.Printf("%s\n", html)
}
