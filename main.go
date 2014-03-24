package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"github.com/kr/pretty"
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
		ret = m(indent) + "<ol class=\"tree\">\n" + m(indent + 1) + "<li>"
		name = "/"
	} else {
		ret = m(indent) + "<ol>\n" + m(indent + 1) + "<li>"
	}
	id := strings.Replace(tree.path, "/", "_", -1)
	if name == "/" {
		ret += fmt.Sprintf(`<label for="root">%s</label> <input type="checkbox" checked="checked" id="root">`, name) + "\n"
	} else {
		ret += fmt.Sprintf(`<label for="%s">%s</label> <input type="checkbox" id="%s">`, id, name, id) + "\n"
	}

	ret += m(indent + 2) + "<ol>\n"
	if len(tree.subdirs) > 0 {
		for _, s := range tree.subdirs {
			ret += m(indent + 4) + "<li>\n" + markupFromTree(s, indent + 5) + m(indent + 4) + "</li>\n"
		}
	}

	if len(tree.files) > 0 {
		for _, f := range tree.files {
			ret += m(indent + 3) + fmt.Sprintf(`<li class="file"><a href="%s%s">%s</a></li>`, tree.path, f, f) + "\n"
		}
	}
	ret += m(indent + 2) + "</ol>\n"

	ret += m(indent + 1) + "</li>\n" + m(indent) + "</ol>\n"
	return ret
}

func treeFromFiles(files map[string][]string) dir {
	root := dir{name: "", files: files[""], subdirs: []dir{}}
	keys := make([]string, len(files))
	sort.Strings(keys)
	for key, _ := range files {
		keys = append(keys, key)
	}

	for key := range files {
		var currentDir *dir = &root
		var newDir *dir
		subdirNames := strings.Split(key, "/")
		for i, d := range subdirNames {
			// Skip empty post-trailing-/ string
			if i == len(subdirNames) - 1 {
				continue
			}
			path := strings.Join(subdirNames[0:i+1], "/") + "/"
			j := subdirIndex(currentDir.subdirs, d)
			if j == -1 {
				newDir = new(dir)
				newDir.name = d
				newDir.path = path
				newDir.files = files[path]
				newDir.subdirs = []dir{}

				currentDir.subdirs = append(currentDir.subdirs, (*newDir))
				j = subdirIndex(currentDir.subdirs, d)
			}
			currentDir = &currentDir.subdirs[j]
		}
	}
	return root
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

	f := treeFromFiles(files)
	html := fmt.Sprintf(htmlHeader, root)
	html += markupFromTree(f, 1)
	html += htmlFooter
	fmt.Printf("%s\n", html)
}
