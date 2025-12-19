package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func md_to_html(input_path string, output_path string) error {
	var cmd *exec.Cmd
	var err error

	/* open source markdown */
	var in *os.File
	in, err = os.Open(input_path)
	if err != nil {
		return err
	}
	defer in.Close()

	/* create output html */
	var out *os.File
	out, err = os.Create(output_path)
	if err != nil {
		return err
	}
	defer out.Close()

	/*
		run:
		  lowdown -Thtml
		stdin  <-  markdown file
		stdout ->  html file
	*/
	cmd = exec.Command("lowdown", "-Thtml")
	cmd.Stdin = in
	cmd.Stdout = out
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func copy_file(src string, dst string) error {
	var err error

	var in *os.File
	in, err = os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	var out *os.File
	out, err = os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	var err error

	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: kew <in> <out>\n")
		os.Exit(1)
	}

	var src string = os.Args[1]
	var out string = os.Args[2]

	fmt.Printf("kew: %s -> %s\n", src, out)

	err = filepath.WalkDir(src, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		/* find relative path */
		var rel string
		rel, err = filepath.Rel(src, path)
		if err != nil {
			return err
		}

		var outpath string = filepath.Join(out, rel)

		/* mirror src */
		if entry.IsDir() {
			return os.MkdirAll(outpath, 0755)
		}

		/* convert markdown else copy */
		if strings.HasSuffix(path, ".md") {
			outpath = strings.TrimSuffix(outpath, ".md") + ".html"
			return md_to_html(path, outpath)
		}

		return copy_file(path, outpath)
	})

	if err != nil {
		fmt.Fprintln(os.Stderr, "kew:", err)
		os.Exit(1)
	}

	fmt.Println("kew: done")
}
