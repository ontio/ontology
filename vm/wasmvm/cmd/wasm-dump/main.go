// Copyright 2017 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"

	"github.com/Ontology/vm/wasmvm/disasm"
	"github.com/Ontology/vm/wasmvm/wasm"
	"github.com/Ontology/vm/wasmvm/wasm/leb128"
)

// TODO: track the number of imported funcs,memories,tables and globals to adjust
// for their index offset when printing sections' content.

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: wasm-dump [options] file1.wasm [file2.wasm [...]]

ex:
 $> wasm-dump -h ./file1.wasm

options:
`,
		)
		flag.PrintDefaults()
		os.Exit(1)
	}
}

var (
	flagVerbose = flag.Bool("v", false, "enable/disable verbose mode")
	flagHeaders = flag.Bool("h", false, "print headers")
	// flagSection = flag.String("j", "", "select just one section")
	flagFull    = flag.Bool("s", false, "print raw section contents")
	flagDis     = flag.Bool("d", false, "disassemble function bodies")
	flagDetails = flag.Bool("x", false, "show section details")
)

func main() {
	log.SetPrefix("wasm-dump: ")
	log.SetFlags(0)

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	if !*flagHeaders && !*flagFull && !*flagDis && !*flagDetails {
		flag.Usage()
		flag.PrintDefaults()
		log.Printf("At least one of -d, -h, -x or -s must be given")
		os.Exit(1)
	}

	wasm.SetDebugMode(*flagVerbose)

	for i, fname := range flag.Args() {
		if i > 0 {
			fmt.Printf("\n")
		}
		process(fname)
	}
}

func process(fname string) {
	f, err := os.Open(fname)
	if err != nil {
		log.Fatalf("could not open %q: %v", fname, err)
	}
	defer f.Close()

	m, err := wasm.ReadModule(f, nil)
	if err != nil {
		log.Fatalf("could not read module: %v", err)
	}

	if *flagHeaders {
		printHeaders(f.Name(), m)
	}
	if *flagFull {
		printFull(f.Name(), m)
	}
	if *flagDis {
		printDis(f.Name(), m)
	}
	if *flagDetails {
		printDetails(f.Name(), m)
	}
}

func printHeaders(fname string, m *wasm.Module) {
	fmt.Printf("%s: module version: %#x\n\n", fname, m.Version)
	fmt.Printf("sections:\n\n")

	hdrfmt := "%9s start=0x%08x end=0x%08x (size=0x%08x) count: %d\n"
	if sec := m.Types; sec != nil {
		fmt.Printf(hdrfmt,
			sec.ID.String(),
			sec.Section.Start, sec.Section.End, sec.Section.PayloadLen,
			len(sec.Entries),
		)
	}
	if sec := m.Import; sec != nil {
		fmt.Printf(hdrfmt,
			sec.ID.String(),
			sec.Section.Start, sec.Section.End, sec.Section.PayloadLen,
			len(sec.Entries),
		)
	}
	if sec := m.Function; sec != nil {
		fmt.Printf(hdrfmt,
			sec.ID.String(),
			sec.Section.Start, sec.Section.End, sec.Section.PayloadLen,
			len(sec.Types),
		)
	}
	if sec := m.Table; sec != nil {
		fmt.Printf(hdrfmt,
			sec.ID.String(),
			sec.Section.Start, sec.Section.End, sec.Section.PayloadLen,
			len(sec.Entries),
		)
	}
	if sec := m.Memory; sec != nil {
		fmt.Printf(hdrfmt,
			sec.ID.String(),
			sec.Section.Start, sec.Section.End, sec.Section.PayloadLen,
			len(sec.Entries),
		)
	}
	if sec := m.Global; sec != nil {
		fmt.Printf(hdrfmt,
			sec.ID.String(),
			sec.Section.Start, sec.Section.End, sec.Section.PayloadLen,
			len(sec.Globals),
		)
	}
	if sec := m.Export; sec != nil {
		fmt.Printf(hdrfmt,
			sec.ID.String(),
			sec.Section.Start, sec.Section.End, sec.Section.PayloadLen,
			len(sec.Entries),
		)
	}
	if sec := m.Start; sec != nil {
		hdrfmt := "%9s start=0x%08x end=0x%08x (size=0x%08x) start: %d\n"
		fmt.Printf(hdrfmt,
			sec.ID.String(),
			sec.Section.Start, sec.Section.End, sec.Section.PayloadLen,
			sec.Index,
		)
	}
	if sec := m.Elements; sec != nil {
		fmt.Printf(hdrfmt,
			sec.ID.String(),
			sec.Section.Start, sec.Section.End, sec.Section.PayloadLen,
			len(sec.Entries),
		)
	}
	if sec := m.Code; sec != nil {
		fmt.Printf(hdrfmt,
			sec.ID.String(),
			sec.Section.Start, sec.Section.End, sec.Section.PayloadLen,
			len(sec.Bodies),
		)
	}
	if sec := m.Data; sec != nil {
		fmt.Printf(hdrfmt,
			sec.ID.String(),
			sec.Section.Start, sec.Section.End, sec.Section.PayloadLen,
			len(sec.Entries),
		)
	}
	for _, sec := range m.Other {
		fmt.Printf("%9s start=0x%08x end=0x%08x (size=0x%08x) %q\n",
			sec.ID.String(),
			sec.Start, sec.End, sec.PayloadLen,
			sec.Name,
		)
	}
}

func printFull(fname string, m *wasm.Module) {
	fmt.Printf("%s: module version: %#x\n\n", fname, m.Version)

	hdrfmt := "contents of section %s:\n"
	var sections []*wasm.Section

	if sec := m.Types; sec != nil {
		sections = append(sections, &sec.Section)
	}
	if sec := m.Import; sec != nil {
		sections = append(sections, &sec.Section)
	}
	if sec := m.Function; sec != nil {
		sections = append(sections, &sec.Section)
	}
	if sec := m.Table; sec != nil {
		sections = append(sections, &sec.Section)
	}
	if sec := m.Memory; sec != nil {
		sections = append(sections, &sec.Section)
	}
	if sec := m.Global; sec != nil {
		sections = append(sections, &sec.Section)
	}
	if sec := m.Export; sec != nil {
		sections = append(sections, &sec.Section)
	}
	if sec := m.Start; sec != nil {
		sections = append(sections, &sec.Section)
	}
	if sec := m.Elements; sec != nil {
		sections = append(sections, &sec.Section)
	}
	if sec := m.Code; sec != nil {
		sections = append(sections, &sec.Section)
	}
	if sec := m.Data; sec != nil {
		sections = append(sections, &sec.Section)
	}
	for i := range m.Other {
		sections = append(sections, &m.Other[i])
	}

	for _, sec := range sections {
		fmt.Printf(hdrfmt, sec.ID.String())
		fmt.Println(hexDump(sec.Bytes, uint(sec.Start)))
	}
}

func printDis(fname string, m *wasm.Module) {
	fmt.Printf("%s: module version: %#x\n\n", fname, m.Version)
	fmt.Printf("code disassembly:\n")
	for i := range m.Function.Types {
		f := m.GetFunction(i)
		fmt.Printf("\nfunc[%d]: %v\n", i, f.Sig)
		dis, err := disasm.Disassemble(*f, m)
		if err != nil {
			log.Fatal(err)
		}
		offset := 0
		for _, code := range dis.Code {
			n := 1
			buf := new(bytes.Buffer)
			str := new(bytes.Buffer)
			fmt.Fprintf(buf, "%02x", code.Op.Code)
			fmt.Fprintf(str, "%v", code.Op.Name)
			for _, im := range code.Immediates {
				imbuf := new(bytes.Buffer)
				binary.Write(imbuf, binary.LittleEndian, im)
				n += imbuf.Len() / 2
				for _, cc := range imbuf.Bytes() {
					fmt.Fprintf(buf, " %02x", cc)
				}
				fmt.Fprintf(str, " %v", im)
			}
			fmt.Printf(" %06x: %-26s | %s\n", offset, buf.String(), str.String())
			offset += 2 * n
		}
		fmt.Printf(" %06x: %-26s | %s\n", offset, fmt.Sprintf("%02x", 0xb), "end")
	}
}

func printDetails(fname string, m *wasm.Module) {
	fmt.Printf("%s: module version: %#x\n\n", fname, m.Version)
	fmt.Printf("section details:\n\n")

	if sec := m.Types; sec != nil {
		fmt.Printf("%v:\n", sec.ID)
		for i, f := range sec.Entries {
			fmt.Printf(" - type[%d] %v\n", i, f)
		}
	}
	if sec := m.Import; sec != nil {
		fmt.Printf("%v:\n", sec.ID)
		for i, e := range sec.Entries {
			buf := new(bytes.Buffer)
			switch typ := e.Type.(type) {
			case wasm.GlobalVarImport:
				fmt.Fprintf(buf, "%s mutable=%v",
					typ.Type.Type,
					typ.Type.Mutable,
				)
			case wasm.FuncImport:
				fmt.Fprintf(buf, "sig=%v", typ.Type)
			case wasm.MemoryImport:
				fmt.Fprintf(buf, "pages: initial=%d max=%d",
					typ.Type.Limits.Initial,
					typ.Type.Limits.Maximum,
				)
			case wasm.TableImport:
				fmt.Fprintf(buf, "elem_type=%v init=%v max=%v",
					typ.Type.ElementType,
					typ.Type.Limits.Initial,
					typ.Type.Limits.Maximum,
				)
			}
			fmt.Printf(" - %v[%d] %s <- %s.%s\n",
				e.Kind, i, buf.String(), e.ModuleName, e.FieldName,
			)
		}
	}
	if sec := m.Function; sec != nil {
		fmt.Printf("%v:\n", sec.ID)
		for i, t := range sec.Types {
			fmt.Printf(" - func[%d] sig=%d\n", i, t)
		}
	}
	if sec := m.Table; sec != nil {
		fmt.Printf("%v:\n", sec.ID)
		for i, e := range sec.Entries {
			fmt.Printf(" - table[%d] type=%v initial=%v\n", i, e.ElementType, e.Limits.Initial)
		}
	}
	if sec := m.Memory; sec != nil {
		fmt.Printf("%v:\n", sec.ID)
		for i, e := range sec.Entries {
			fmt.Printf(" - memory[%d] pages: initial=%v\n", i, e.Limits.Initial)
		}
	}
	if sec := m.Global; sec != nil {
		fmt.Printf("%v:\n", sec.ID)
		for i, g := range sec.Globals {
			// TODO(sbinet) display init infos
			fmt.Printf(" - global[%d] %v mutable=%v -- init: %#v\n", i, g.Type.Type, g.Type.Mutable, g.Init)
		}
	}
	if sec := m.Export; sec != nil {
		fmt.Printf("%v:\n", sec.ID)
		keys := make([]string, 0, len(sec.Entries))
		for n := range sec.Entries {
			keys = append(keys, n)
		}
		sort.Strings(keys)
		for _, name := range keys {
			e := sec.Entries[name]
			fmt.Printf(" - %v[%d] -> %q\n", e.Kind, e.Index, name)
		}
	}
	if sec := m.Start; sec != nil {
		fmt.Printf("%v:\n", sec.ID)
		fmt.Printf(" - start function: %d\n", sec.Index)
	}
	if sec := m.Elements; sec != nil {
		fmt.Printf("%v:\n", sec.ID)
		for i, e := range sec.Entries {
			fmt.Printf(" - segment[%d] table=%d\n", i, e.Index)
			fmt.Printf(" - init: %#v\n", e.Offset)
			for ii, elem := range e.Elems {
				fmt.Printf("  - elem[%d] = func[%d]\n", ii, elem)
			}
		}
	}
	if sec := m.Data; sec != nil {
		fmt.Printf("%v:\n", sec.ID)
		for i, e := range sec.Entries {
			fmt.Printf(" - segment[%d] size=%d - init %#v\n", i, len(e.Data), e.Offset)
			fmt.Printf("%s", hexDump(e.Data, 0))
		}
	}
	for _, sec := range m.Other {
		fmt.Printf("%v:\n", sec.ID)
		fmt.Printf(" - name: %q\n", sec.Name)
		raw := bytes.NewReader(sec.Bytes[6:])
		for {
			if raw.Len() == 0 {
				break
			}
			i, err := leb128.ReadVarUint32(raw)
			if err != nil {
				log.Fatal(err)
			}
			n, err := leb128.ReadVarUint32(raw)
			if err != nil {
				log.Fatal(err)
			}
			str := make([]byte, int(n))
			_, err = io.ReadFull(raw, str)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf(" - func[%d] %v\n", i, string(str))
		}
	}
}
