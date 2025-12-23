package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type NavNode struct {
	Name     string
	Path     string
	Files    []NavNode
	Children []NavNode
}

func build_nav(dir string, root string) (NavNode, bool) {
	var node NavNode
	node.Name = title_from_name(filepath.Base(dir))

	entries, err := os.ReadDir(dir)
	if err != nil {
		return node, false
	}

	for _, e := range entries {
		full := filepath.Join(dir, e.Name())

		if e.IsDir() {
			child, ok := build_nav(full, root)
			if ok {
				node.Children = append(node.Children, child)
			}
			continue
		}

		if strings.HasSuffix(e.Name(), ".md") {
			if e.Name() == "index.md" {
				rel_dir, _ := filepath.Rel(root, dir)
				if rel_dir == "." {
					node.Path = "index.html"
				} else {
					node.Path = rel_dir + "/index.html"
				}
				continue
			}
			rel, _ := filepath.Rel(root, full)
			html := strings.TrimSuffix(rel, ".md") + ".html"

			node.Files = append(node.Files, NavNode{
				Name: title_from_name(e.Name()),
				Path: html,
			})
		}
	}

	if len(node.Files) == 0 && len(node.Children) == 0 && node.Path == "" {
		return node, false
	}

	return node, true
}

func copy_file(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func markdown_to_html(path string) (string, error) {
	cmd := exec.Command("lowdown", "-Thtml")

	in, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer in.Close()

	var out strings.Builder
	cmd.Stdin = in
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

func render_nav(n NavNode, b *strings.Builder, cur string) {
	b.WriteString("<ul>\n")

	for _, f := range n.Files {
		sym := NavFileSymbol
		if f.Path == cur {
			sym = NavCurrentSymbol
		}
		b.WriteString(`<li><a href="` + f.Path + `">` + sym + f.Name + "</a></li>\n")
	}

	for _, c := range n.Children {
		sym := NavDirSymbol
		if c.Path == cur {
			sym = NavCurrentSymbol
		}

		if c.Path != "" {
			b.WriteString(`<li><a href="` + c.Path + `">` + c.Name + sym + `</a>`)
		} else {
			b.WriteString("<li>" + c.Name + sym)
		}

		render_nav(c, b, cur)
		b.WriteString("</li>\n")
	}

	b.WriteString("</ul>\n")
}

func replace_md_references(s string) string {
	r := strings.NewReplacer(
		/* common cases */
		".md)", ".html)",
		".md\"", ".html\"",
		".md'", ".html'",
		".md)", ".html)",
		".md#", ".html#",
		".md>", ".html>",
		".md ", ".html ",
		".md,", ".html,",
	)
	return r.Replace(s)
}

func title_from_name(name string) string {
	name = strings.TrimSuffix(name, ".md")
	name = strings.ReplaceAll(name, "-", " ")
	return name
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: kew <in> <out>\n")
		os.Exit(1)
	}

	src := os.Args[1]
	out := os.Args[2]

	/* load template */
	tmpl, err := os.ReadFile(filepath.Join(src, TemplateFile))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	/* build nav */
	rootnav, _ := build_nav(src, src)

	/* walk site */
	err = filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel(src, path)
		outpath := filepath.Join(out, rel)

		if d.IsDir() {
			return os.MkdirAll(outpath, 0755)
		}

		if strings.HasSuffix(path, ".md") {
			html, err := markdown_to_html(path)
			if err != nil {
				return err
			}
			html = replace_md_references(html)

			relhtml := strings.TrimSuffix(rel, ".md") + ".html"
			cur := relhtml
			var navbuf strings.Builder
			render_nav(rootnav, &navbuf, cur)

			page := string(tmpl)
			page = strings.Replace(page, "{{TITLE}}", SiteTitle, 1)
			page = strings.Replace(page, "{{NAV}}", navbuf.String(), 1)
			page = strings.Replace(page, "{{CONTENT}}", html, 1)
			page = strings.Replace(page, "{{FOOTER}}", FooterText, 1)

			outpath = strings.TrimSuffix(outpath, ".md") + ".html"
			return os.WriteFile(outpath, []byte(page), 0644)
		}

		return copy_file(path, outpath)
	})

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
