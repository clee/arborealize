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
	files []os.FileInfo
	subdirs []dir
}

type ByDirName []dir
func (d ByDirName) Len() int { return len(d) }
func (d ByDirName) Swap(i, j int) { d[i], d[j] = d[j], d[i] }
func (d ByDirName) Less(i, j int) bool { return d[i].name < d[j].name }

type ByFileName []os.FileInfo
func (f ByFileName) Len() int { return len(f) }
func (f ByFileName) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
func (f ByFileName) Less(i, j int) bool { return f[i].Name() < f[j].Name() }

const htmlHeader = `<html>
<head>
<title>%s</title>
<link rel="stylesheet" href="//clee.github.io/arborealize/arborealize.css" />
<meta name="generator" content="http://github.com/clee/arborealize/" />
</head>
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

func human(size int64 ) string {
	suffixes := []string{"bytes", "KB", "MB", "GB", "TB", "PB"}
	suffixIndex := 0

	for size > 1024 {
		size >>= 10
		suffixIndex += 1
	}
	return fmt.Sprintf("%d %s", size, suffixes[suffixIndex])
}

func markupFromTree(tree dir, indent int) {
	name := tree.name
	if tree.name == "" {
		fmt.Printf(m(indent) + "<ol class=\"tree\">\n" + m(indent + 1) + "<li>\n")
		name = "/"
	}
	id := strings.Replace(tree.path, "/", "_", -1)
	if name == "/" {
		fmt.Printf(m(indent + 2) + `<input type="checkbox" checked="checked" id="root"><label for="root">/</label>`)
	} else {
		fmt.Printf(m(indent + 2) + `<input type="checkbox" id="` + id + `"><label for="` + id + `">` + name + `</label>`)
	}

	fmt.Printf("\n" + m(indent + 2) + "<ol>\n")

	sort.Sort(ByDirName(tree.subdirs))
	for _, s := range tree.subdirs {
		fmt.Printf(m(indent + 3) + "<li>\n")
		markupFromTree(s, indent + 2)
		fmt.Printf(m(indent + 3) + "</li>\n")
	}

	sort.Sort(ByFileName(tree.files))
	for _, f := range tree.files {
		fmt.Printf(m(indent + 3) + `<li class="file"><a href="` + tree.path + f.Name() + `">` + f.Name() + ` <span class="filesize">` + human(f.Size()) + `</span></a></li>`)
	}

	fmt.Printf("\n" + m(indent + 2) + "</ol>\n")

	if name == "/" {
		fmt.Printf(m(indent + 1) + "</li>\n" + m(indent) + "</ol>\n")
	}
}

func treeFromFiles(files map[string][]os.FileInfo) dir {
	root := dir{name: "", files: files[""], subdirs: []dir{}}

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
	files := make(map[string][]os.FileInfo)

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
			files[dir] = append(files[dir], f)
		}
		return nil
	})

	f := treeFromFiles(files)
	fmt.Printf(htmlHeader, root)
	markupFromTree(f, 1)
	fmt.Printf(htmlFooter)
}
