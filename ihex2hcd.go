package ihex2hcd

import (
	"fmt"
	"bytes"
	"io"
	"bufio"
	"log"
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
	Address   []byte
	Type      RecordType
	Data      []byte
	UpperAddr int64
}

func ParseString(input string) (*Record, error){
	p := &Parser{input: []byte(input), current: -1}
	r, err := p.Parse()
	return r, err
}

func New(r io.Reader) Hex2Bin {
	b := bytes.NewBuffer(nil)
	h := Hex2Bin{Buffer: b, r: r}
	return h
}

func (h *Hex2Bin) BinOutput(writer io.Writer) error {
	scanner := bufio.NewScanner(h.r)
	for scanner.Scan() {
		r, err := ParseString(scanner.Text())
		if err != nil {
			return err
		}
		h.records = append(h.records, r)
	}
	for _, r := range h.records {
		writer.Write(r.processRecord())
	}
	return nil
}

func (h *Hex2Bin) StringOutput() error {
	var records []*Record
	scanner := bufio.NewScanner(h.r)
	for scanner.Scan() {
		r, err := ParseString(scanner.Text())
		if err != nil {
			return err
		}
		records = append(records, r)
	}
	for _, r := range records {
		r.toString()
	}
	return nil
}

func (h *Hex2Bin) RecordOutput() ([]*Record, error) {
	scanner := bufio.NewScanner(h.r)
	for scanner.Scan() {
		r, err := ParseString(scanner.Text())
		if err != nil {
			return nil, err
		}
		h.records = append(h.records, r)
	}
	return h.records, nil
}

func (r *Record) toString() {
	switch r.Type {
	case RecordTypeData:
		fmt.Printf("[Address:%04X ByteCount:%d Data:%04X ]\n", r.Address, r.ByteCount, r.Data)

	case RecordTypeEOF:
		fmt.Sprint("EOF")

	case RecordTypeExtendedLinearAddress:
		if r.ByteCount == 2 {
			upperAddr = int64(((r.Data[0] & 0xFF) << 8) + (r.Data[1] & 0xFF))
			upperAddr <<= 16 // ELA is bits 16-31 of the segment base address (SBA), so shift left 16 bits
			fmt.Printf("[Extended Linear Address record:%04X ]\n", upperAddr)
		} else {
			log.Fatalf("Extended Linear Address record: %s", r.Data)
		}

	case RecordTypeExtendedSegmentAddress:
		if r.ByteCount == 2 {
			upperAddr = int64(((r.Data[0] & 0xFF) << 8) + (r.Data[1] & 0xFF))
			upperAddr <<= 4 // ESA is bits 4-19 of the segment base address (SBA), so shift left 4 bits
			fmt.Printf("[Extended Segment Address record:%04X ]\n", upperAddr)

		} else {
			log.Fatalf("Invalid Extended Segment Address record: %s", r.Data)
		}

	case RecordTypeStartLinearAddress:
		if r.ByteCount == 4 {
			startAddr = 0
			for i := range r.Data {
				startAddr = startAddr << 8
				startAddr |= int64(r.Data[i] & 0xFF)
				fmt.Printf("[Start Linear Addressrecord:%04X ]\n", startAddr)
			}
		} else {
			log.Fatalf("Invalid Start Linear Address record: %s", r.Data)
		}

	case RecordTypeStartSegmentAddress:
		if r.ByteCount == 4 {
			startAddr = 0
			for i := range r.Data {
				startAddr = startAddr << 8
				startAddr |= int64(r.Data[i] & 0xFF)
				fmt.Printf("[Start Segment Address :%04X ]\n", startAddr)
			}
		} else {
			log.Fatalf("Invalid Start Segment Address record: %s", r.Data)
		}
	default:
		fmt.Printf(string(r.Data))
	}
}

func (r *Record) processRecord() []byte {
	addr := int64(((r.Address[0] & 0xFF) << 8) + (r.Address[1] & 0xFF))
	offset := addr | upperAddr

	switch r.Type {
	case RecordTypeData:
		return writeData(offset, r.Address, r.Data, r.ByteCount)

	case RecordTypeEOF:
		return []byte{0x4E, 0xFC, 0x04, 0xFF, 0xFF, 0xFF, 0xFF}

	case RecordTypeExtendedLinearAddress:
		if r.ByteCount == 2 {
			upperAddr = int64(((r.Data[0] & 0xFF) << 8) + (r.Data[1] & 0xFF))
			upperAddr <<= 16 // ELA is bits 16-31 of the segment base address (SBA), so shift left 16 bits
		} else {
			log.Fatalf("Extended Linear Address record: %s", r.Data)
		}

	case RecordTypeExtendedSegmentAddress:
		if r.ByteCount == 2 {
			upperAddr = int64(((r.Data[0] & 0xFF) << 8) + (r.Data[1] & 0xFF))
			upperAddr <<= 4 // ESA is bits 4-19 of the segment base address (SBA), so shift left 4 bits
		} else {
			log.Fatalf("Invalid Extended Segment Address record: %s", r.Data)
		}

	case RecordTypeStartLinearAddress:
		if r.ByteCount == 4 {
			startAddr = 0
			for i := range r.Data {
				startAddr = startAddr << 8
				startAddr |= int64(r.Data[i] & 0xFF)
			}
		} else {
			log.Fatalf("Invalid Start Linear Address record: %s", r.Data)
		}

	case RecordTypeStartSegmentAddress:
		if r.ByteCount == 4 {
			startAddr = 0
			for i := range r.Data {
				startAddr = startAddr << 8
				startAddr |= int64(r.Data[i] & 0xFF)
			}
		} else {
			log.Fatalf("Invalid Start Segment Address record: %s", r.Data)
		}
	}
	return nil
}
func writeData(offset int64, addr []byte, data []byte, count int) []byte {
	buf := []byte{0x4c, 0xfc, byte(count + 4), byte(offset), addr[0], byte(offset >> 16), byte(offset >> 24)}
	nb := append(buf[:], data[:]...)
	return nb
}