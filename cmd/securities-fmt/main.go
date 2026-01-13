package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

type Block struct {
	Header []string
	Rows   [][]string
}

func main() {
	var (
		write = flag.Bool("w", false, "writes result to file")
		file  = flag.String("file", "securities.csv", "file to format")
		check = flag.Bool("check", false, "check if file is already formatted")
	)

	flag.Parse()

	blocks, err := readBlocks(*file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ failed reading blocks: %s\n", err.Error())
		os.Exit(1)
	}

	formatted, err := formatToString(blocks)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ failed formatting string: %s\n", err.Error())
		os.Exit(1)
	}

	if *check {
		original, err := os.ReadFile(*file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ failed reading file: %s\n", err.Error())
			os.Exit(1)
		}

		if string(original) != formatted {
			fmt.Println("❌ securities.csv format check failed")
			os.Exit(1)
		}

		fmt.Println("✅ securities.csv formatted correctly")
		return
	}

	if *write {
		if err := os.WriteFile(*file, []byte(formatted), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "❌ failed writing file: %s\n", err.Error())
			os.Exit(1)
		}

		fmt.Println("✅ securities.csv formatted")
		return
	}

	// default: stampa su stdout
	fmt.Print(formatted)
}

func readBlocks(path string) ([]Block, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)

	var blocks []Block
	var cur Block

	flush := func() {
		if len(cur.Header) > 0 || len(cur.Rows) > 0 {
			blocks = append(blocks, cur)
			cur = Block{}
		}
	}

	for sc.Scan() {
		line := sc.Text()
		trim := strings.TrimSpace(line)

		switch {
		case trim == "":
			flush()

		case strings.HasPrefix(trim, "#"):
			if len(cur.Rows) > 0 {
				flush()
			}
			cur.Header = append(cur.Header, line)

		default:
			r := csv.NewReader(strings.NewReader(line))
			rec, err := r.Read()
			if err != nil || len(rec) != 3 {
				return nil, fmt.Errorf("invalid CSV line: %q", line)
			}
			cur.Rows = append(cur.Rows, rec)
		}
	}

	flush()
	return blocks, sc.Err()
}

func writeBlocks(w *bufio.Writer, blocks []Block) error {
	for i, b := range blocks {
		// Header
		for _, h := range b.Header {
			fmt.Fprintln(w, h)
		}

		sort.SliceStable(b.Rows, func(i, j int) bool {
			return b.Rows[i][0] < b.Rows[j][0]
		})

		for _, row := range b.Rows {
			if err := writeQuotedRow(w, row); err != nil {
				return err
			}
		}

		if i < len(blocks)-1 {
			fmt.Fprintln(w)
		}
	}
	return w.Flush()
}

func writeQuotedRow(w *bufio.Writer, row []string) error {
	for i, v := range row {
		// Escape CSV: le " diventano ""
		escaped := strings.ReplaceAll(v, `"`, `""`)

		if _, err := fmt.Fprintf(w, `"%s"`, escaped); err != nil {
			return err
		}

		if i < len(row)-1 {
			if err := w.WriteByte(','); err != nil {
				return err
			}
		}
	}
	return w.WriteByte('\n')
}

func formatToString(blocks []Block) (string, error) {
	var sb strings.Builder
	w := bufio.NewWriter(&sb)

	if err := writeBlocks(w, blocks); err != nil {
		return "", err
	}
	return sb.String(), nil
}
