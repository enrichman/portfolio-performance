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
	Header []string   // commenti e separatori
	Rows   [][]string // record CSV
}

func die(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, "❌ "+msg+"\n", args...)
	os.Exit(1)
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
				return nil, fmt.Errorf("riga CSV non valida: %q", line)
			}
			cur.Rows = append(cur.Rows, rec)
		}
	}

	flush()
	return blocks, sc.Err()
}

func writeBlocks(path string, blocks []Block) error {
	tmp := path + ".tmp"

	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	for i, b := range blocks {
		// Header
		for _, h := range b.Header {
			fmt.Fprintln(w, h)
		}

		// Ordina per prima colonna
		sort.SliceStable(b.Rows, func(i, j int) bool {
			return b.Rows[i][0] < b.Rows[j][0]
		})

		// Scrive righe CSV
		for _, row := range b.Rows {
			if err := writeQuotedRow(w, row); err != nil {
				return err
			}
		}

		// Riga vuota tra blocchi (non dopo l'ultimo)
		if i < len(blocks)-1 {
			fmt.Fprintln(w)
		}
	}

	if err := w.Flush(); err != nil {
		return err
	}

	return os.Rename(tmp, path)
}

func main() {
	var (
		write = flag.Bool("w", false, "riscrive securities.csv")
		file  = flag.String("file", "securities.csv", "file da formattare")
	)
	flag.Parse()

	blocks, err := readBlocks(*file)
	if err != nil {
		die("%v", err)
	}

	if *write {
		if err := writeBlocks(*file, blocks); err != nil {
			die("%v", err)
		}
		fmt.Println("✅ file formattato")
		return
	}

	// dry-run: scrive su stdout
	if err := writeBlocks(*file+".formatted", blocks); err != nil {
		die("%v", err)
	}
	fmt.Println("ℹ️ scritto:", *file+".formatted")
	fmt.Println("Usa: diff -u", *file, *file+".formatted")
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
