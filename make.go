package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
)

//run: go run %
const (
	OUT_KEEP = 0
)

var L_INFO = log.New(io.Discard, "", 0)
var L_ERROR = log.New(io.Discard, "", 0)
const (
	TRACE uint = iota 
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
	PANIC
)

var ROOT_DIR string
func main() {
	if x, err := os.Getwd(); err != nil {
		log.Fatal(err)
	} else {
		ROOT_DIR = x
	}
	{
		log_level := INFO
		if log_level <= INFO { L_INFO = log.New(os.Stderr, "", 0) }
		if log_level <= ERROR { L_ERROR = log.New(os.Stderr, "", log.Lshortfile) }
	}

	if len(os.Args) <= 1 {
		Make("cache")
		Make("build")
	}
	for _, arg := range os.Args[1:] {
		Make(arg)
	}
}

func Make(arg string) {
	//var output bytes.Buffer
	//cmd := cmd_start([]string{}, "", R("hello"), &output, "cat", "-")
	//Must1(cmd.Wait())
	//fmt.Println(output.String())

	switch arg {
	case "cache":
		cache()
	case "build":
		walk_src_files(ROOT_DIR, func (path, name string) error {
			dir := filepath.Dir(path)

			L_INFO.Printf("=== Compiling %q ===\n", filepath.Base(dir))
			target := filepath.Join(dir, "index.smd")
			fh := Must(os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644))
			cmd_start([]string{"BUILD=build"}, dir, nil, fh, "tetra", "run", "file", path)
			//cmd_start([]string{}, io.Discard, "tetra", "run", "file", )
			return fh.Close()
		})
	default:
		log.Fatalf("Unknown command: %q\n", arg)
	}
}

type Frontmatter struct {
	Title   string `json:"title"`
	Date    string `json:"date"`
	Updated string `json:"updated"`
	Author  string `json:"author"`
	Layout  string `json:"layout"`
	Tags    []string `json:"tags"`
	Draft   bool `json:"draft"`
	Series  []string `json:"series"`
	Unused  string `json:"_"`

	id        string
	links     []string
	backlinks []string
}

const EXPECTED_FILE_COUNT = 1024 // Large value to avoid realloc

func cache() {
	all_frontmatter := make([]Frontmatter, 0, EXPECTED_FILE_COUNT)

	walk_src_files(ROOT_DIR, func (path, name string) error {
		var links_buffer bytes.Buffer
		links_cmd := cmd_start([]string{}, "", nil, &links_buffer, "lychee", "--dump", "--offline", path)

		fh := Must(os.Open(path))
		reader := bufio.NewReader(fh)
		buffer := bytes.NewBuffer(make([]byte, 0, 1024))
		var prev uint8 = 0
		for {
			if char, err := reader.ReadByte(); err != nil {
				if err.Error() == "EOF" {
					break
				}
				L_ERROR.Println("Error reading from file:", err)
				break
			} else if prev == '\n' && char == '}' {
				buffer.WriteByte(char)
				break
			} else {
				buffer.WriteByte(char)
				prev = char
			}
		}

		id := filepath.Base(filepath.Dir(path))
		var frontmatter Frontmatter
		{
			decoder := json.NewDecoder(buffer)
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&frontmatter); err != nil {
				L_ERROR.Fatalf("Could not decode frontmatter of %q\n  %w\n", path, err)
			}
		}

		{
			Must1(links_cmd.Wait())

			size := bytes.Count(links_buffer.Bytes(), []byte{'\n'})
			frontmatter.links = make([]string, 0, size)
			for url := range bytes.SplitSeq(links_buffer.Bytes(), []byte("\n")) {
				if i := bytes.LastIndexByte(url, '/'); i >= 0 {
					url = url[i + 1:]
				}
				if len(url) <= 1 || url[0] == '$' {
					continue
				}
				if i := bytes.LastIndexByte(url, '#'); i >= 0 {
					continue
				}
				frontmatter.links = append(frontmatter.links, string(url))
			}
		}

		frontmatter.id = id
		all_frontmatter = append(all_frontmatter, frontmatter)
		return fh.Close()
	})

	// Make it so that series are in chronological order
	sort.Slice(all_frontmatter, func (i, j int) bool {
		return all_frontmatter[i].Date < all_frontmatter[j].Date;
	})

	backlinks := make(map[string][]string, EXPECTED_FILE_COUNT)
	series := make(map[string][]string, EXPECTED_FILE_COUNT)
	for _, frontmatter := range all_frontmatter {
		for _, target := range frontmatter.links {
			if x, ok := backlinks[target]; ok {
				backlinks[target] = append(x, frontmatter.id)
			} else {
				backlinks[target] = []string{frontmatter.id}
			}
		}

		for _, target := range frontmatter.Series {
			if x, ok := series[target]; ok {
				series[target] = append(x, frontmatter.id)
			} else {
				series[target] = []string{frontmatter.id}
			}
		}
	}

	type Cache struct {
		Backlinks map[string][]string `json:"backlinks"`
		Series    map[string][]string `json:"series"`
	}
	cache :=  Cache {
		Backlinks: backlinks,
		Series: series,
	}

	{
		L_INFO.Println("=== Creating assets/cache.ziggy ===")
		cache_path := filepath.Join(ROOT_DIR, "assets", "cache.ziggy")
		fh := Must(os.OpenFile(cache_path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644))
		encoder := json.NewEncoder(fh)
		encoder.SetIndent("", "  ")
		Must1(encoder.Encode(cache))
		Must1(fh.Close()) 
	}
}

//#run: BUILD=test <archive2/src2.md %
//frontmatter="$( <&0 jq --slurp --raw-input '.
//  | .[0:index("\n}\n") + 2]
//  | fromjson
//' )" || exit "$?"
//
//case "${BUILD:-local}"
//in local)
//  printf %s\\n "${frontmatter}" | jq --raw-output '[
//    "# \(.title)",
//    "Author: \(.author)",
//    ""
//  ] | join("\n")'
//;; build)
//  printf %s\\n "${frontmatter}" | jq --raw-output '[
//    "---",
//    ".title  = \(.title | tojson)",
//    ".date   = @date(\(.date | tojson))",
//    ".author = \( (.author // "Yueleshia") | tojson)",
//    ".layout = \(.layout | tojson)",
//    ".tags   = \(.tags | tojson)",
//    ".draft  = \(.draft)",
//    ".custom = {",
//    "  series = \((.series // null) | tojson)",
//    "}",
//    "---",
//    ""
//  ] | join("\n")'
//;; test)
//  printf %s\\n "${frontmatter}" >&2
//esac

func walk_src_files(root string, process func(string, string) error ) {
	cache_dir := filepath.Join(root, "content", "blog")

	err := filepath.WalkDir(cache_dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // Return the error if there is one
		}
		if d.IsDir() {
			return nil
		}
		base := filepath.Base(path) 
		if base == "src.md" {
			if err := process(path, base); err != nil {
				L_ERROR.Println(err)
			}
		}
		filepath.Join(path, "index.smd")

		// Print the path and whether it's a directory or a file
		return nil // Return nil to continue walking
	})
	if err != nil {
		L_ERROR.Fatal(err)
	}
}


func Must[T any](x T, err error) T {
	if err != nil {
		fmt.Println(err.Error)
		os.Exit(1)
	}
	return x
}
func Must1(err error) {
	if err != nil {
		fmt.Println(err.Error)
		os.Exit(1)
	}
}

func find_go_root() (string, error) {
	var dir string
	if x, err := os.Getwd(); err != nil {
		return "", err
	} else {
		dir = x
	}

	// Unlikely we will be 1000 folders deep
	for i := 0; i < 1000; i += 1 {
		_, err := os.Stat(filepath.Join(dir, "make.go"))
		if os.IsNotExist(err) {
			dir = filepath.Dir(dir)
		} else {
			return dir, nil
		}
	}
	return "", fmt.Errorf("Could not find directory with make.go")
	
}

func R[T ~[]byte | ~string](stdin T) io.Reader {
	return bytes.NewReader([]byte(stdin))
}

func cmd_start(env_vars []string, working_dir string, stdin io.Reader, stdout io.Writer, name string, args ...string)  *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Dir = working_dir
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(cmd.Environ(), env_vars...)
	err := cmd.Start()
	if err != nil {
		L_ERROR.Fatal(err)
	}
	return cmd
}

