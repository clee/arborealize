package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type dir struct {
	Name string
	Path string
	Files []os.FileInfo
	Subdirs []dir
}

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
		if d.Name == name {
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

func markupFromTree(tree dir, indent int) (ret string) {
	name := tree.Name
	if tree.Name == "" {
		ret = m(indent) + "<ol class=\"tree\">\n" + m(indent + 1) + "<li>"
		name = "/"
	} else {
		ret = m(indent) + "<ol>\n" + m(indent + 1) + "<li>"
	}
	id := strings.Replace(tree.Path, "/", "_", -1)
	if name == "/" {
		ret += fmt.Sprintf(`<input type="checkbox" checked="checked" id="root"><label for="root">%s</label>`, name) + "\n"
	} else {
		ret += fmt.Sprintf(`<input type="checkbox" id="%s"><label for="%s">%s</label>`, id, id, name) + "\n"
	}

	ret += m(indent + 2) + "<ol>\n"

	for _, s := range tree.Subdirs {
		ret += m(indent + 3) + "<li>\n" + markupFromTree(s, indent + 4) + m(indent + 3) + "</li>\n"
	}

	for _, f := range tree.Files {
		ret += m(indent + 3) + fmt.Sprintf(`<li class="file"><a href="%s%s">%s <span class="filesize">%s</span></a></li>`, tree.Path, f.Name(), f.Name(), human(f.Size())) + "\n"
	}

	ret += m(indent + 2) + "</ol>\n"

	ret += m(indent + 1) + "</li>\n" + m(indent) + "</ol>\n"
	return ret
}

func markupTemplate(tree dir) string {
	const treeTemplate = `{{if .Name}}
<ol><li><input type="checkbox" id="{{html .Name}}"><label for="{{html .Name}}">{{.Name}}</label>
{{else}}
	<ol class="tree"><li><input type="checkbox" checked="checked" id="root"><label for="root">/</label>
{{end}}
<ol>
{{range .Subdirs}}
	{{markupTemplate .}}
{{end}}
{{with $self := .}}
{{range .Files}}<li class="file"><a href="{{$self.Path}}{{.Name}}">{{.Name}} <span class="filesize">{{human .Size}}</span></a></li>
{{end}}
{{end}}
</ol>
</li>
</ol>`

	var doc bytes.Buffer
	funcMap := template.FuncMap { "human": human, "markupTemplate": markupTemplate }
	t := template.Must(template.New("tree template").Funcs(funcMap).Parse(treeTemplate))
	if e := t.Execute(&doc, tree); e != nil {
		fmt.Printf("what the shit")
		panic(e)
	}

	return doc.String()
}

func treeFromFiles(files map[string][]os.FileInfo) dir {
	root := dir{Name: "", Files: files[""], Subdirs: []dir{}}

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
			j := subdirIndex(currentDir.Subdirs, d)
			if j == -1 {
				newDir = new(dir)
				newDir.Name = d
				newDir.Path = path
				newDir.Files = files[path]
				newDir.Subdirs = []dir{}

				currentDir.Subdirs = append(currentDir.Subdirs, (*newDir))
				j = subdirIndex(currentDir.Subdirs, d)
			}
			currentDir = &currentDir.Subdirs[j]
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
	html := fmt.Sprintf(htmlHeader, root)
	html += markupTemplate(f)
	html += htmlFooter
	fmt.Printf("%s\n", html)
}
