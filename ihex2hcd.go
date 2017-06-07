package ihex2hcd

import (
	"fmt"
	"os"
	"bytes"
	"io"
	"bufio"
)

type RecordType int8

var upperAddr, startAddr int64

const (
	RecordTypeData                   RecordType = iota
	RecordTypeEOF
	RecordTypeExtendedSegmentAddress
	RecordTypeStartSegmentAddress
	RecordTypeExtendedLinearAddress
	RecordTypeStartLinearAddress
)

type Hex2Bin struct {
	Buffer  *bytes.Buffer
	r       io.Reader
	w       io.Writer
	records []*Record
}

type Record struct {
	ByteCount int
	Address   int64
	Type      RecordType
	Data      []byte
	UpperAddr int64
}

func ParseString(input string) *Record {
	p := &Parser{input: []byte(input), current: -1}
	p.Parse()
	return p.rec
}

func New(r io.Reader) Hex2Bin {
	b := bytes.NewBuffer(nil)
	h := Hex2Bin{Buffer: b, r: r}
	return h
}

func (h *Hex2Bin) BinOutput(writer io.Writer) {
	scanner := bufio.NewScanner(h.r)
	for scanner.Scan() {
		h.records = append(h.records, ParseString(scanner.Text()))
	}
	for _, r := range h.records {
		writer.Write(r.processRecord())
	}
}

func (r *Record) processRecord() []byte {
	addr := r.Address | upperAddr

	switch r.Type {
	case RecordTypeData:
		return writeData(addr, r.Address, r.Data)

	case RecordTypeEOF:
		return []byte{0x4E, 0xFC, 0x04, 0xFF, 0xFF, 0xFF, 0xFF}

	case RecordTypeExtendedLinearAddress:
		if r.ByteCount == 2 {
			upperAddr = int64(((r.Data[0] & 0xFF) << 8) + (r.Data[1] & 0xFF))
			upperAddr <<= 16 // ELA is bits 16-31 of the segment base address (SBA), so shift left 16 bits
		} else {
			fmt.Println("Invalid Extended Linear Address record")
			os.Exit(1)
		}

	case RecordTypeExtendedSegmentAddress:
		if r.ByteCount == 2 {
			upperAddr = int64(((r.Data[0] & 0xFF) << 8) + (r.Data[1] & 0xFF))
			upperAddr <<= 4 // ESA is bits 4-19 of the segment base address (SBA), so shift left 4 bits
		} else {
			fmt.Println("Invalid Extended Segment Address record")
			os.Exit(1)
		}

	case RecordTypeStartLinearAddress:
		if r.ByteCount == 4 {
			startAddr = 0
			for i := range r.Data {
				startAddr = startAddr << 8
				startAddr |= int64(r.Data[i] & 0xFF)
			}
		} else {
			fmt.Println("Invalid Start Linear Address record")
			os.Exit(1)
		}

	case RecordTypeStartSegmentAddress:
		if r.ByteCount == 4 {
			startAddr = 0
			for i := range r.Data {
				startAddr = startAddr << 8
				startAddr |= int64(r.Data[i] & 0xFF)
			}
		} else {
			fmt.Println("Invalid Start Segment Address record")
			os.Exit(1)
		}
	}
	return nil
}
func writeData(offset int64, addr int64, data []byte) []byte {
	buf := []byte{0x4c, 0xfc, byte(addr), byte(offset), byte(offset >> 8), byte(offset >> 16), byte(offset >> 24)}
	nb := append(buf[:], data[:]...)
	return nb
}
