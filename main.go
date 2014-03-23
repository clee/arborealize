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
<link rel="stylesheet" href="//netdna.bootstrapcdn.com/bootstrap/3.1.1/css/bootstrap.min.css"/>
<link rel="stylesheet" href="//clee.github.io/arborealize/arborealize.css"/>
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

func m(indent int) string {
	return strings.Repeat(" ", indent)
}

func markupFromTree(tree dir, indent int) (ret string) {
	name := tree.name
	if tree.name == "" {
		ret = m(indent) + "<ol class=\"tree\">\n<li>\n"
		name = "/"
	} else {
		ret = m(indent) + "<ol>\n<li>\n"
	}
	id := strings.Replace(tree.path, "/", "_", -1)
	ret += m(indent + 1) + fmt.Sprintf(`<label for="%s">%s</label> <input type="checkbox" id="%s" />`, id, name, id)
	if len(tree.subdirs) > 0 {
		for _, s := range tree.subdirs {
			ret += markupFromTree(s, indent + 1)
		}
	}
	if len(tree.files) > 0 {
		ret += m(indent + 1) + "<ol>\n"
		for _, f := range tree.files {
			ret += m(indent + 2) + fmt.Sprintf("<li class=\"file\"><a href=\"%s%s\">%s</a></li>\n", tree.path, f, f)
		}
		ret += m(indent + 1) + "</ol>\n"
	}
	ret += m(indent) + "</li>\n</ol>\n"
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
	html += markupFromTree(f, 1)
	html += htmlFooter
	fmt.Printf("%s\n", html)
}
