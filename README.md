# arborealize

This is a utility that generates an HTML5 + CSS3 page with recursive directory contents from a directory you specify on the command line.

## usage

```
$ arborealize > index.html
# or:
$ arborealize -root /path/to/folder > index.html
```

This will take the recursive contents of the current directory and generate an index.html file with their contents. If you use a different `-root` argument, unless you place the index.html in that same root folder, the links will all be broken (because the generated links are all relative links).

## example

You can see some example output at http://clee.github.io/arborealize/ :)
