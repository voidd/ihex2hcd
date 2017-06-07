package ihex2hcd

import (
	"encoding/hex"
	"log"
)

type Parser struct {
	input   []byte
	decodedBytes []byte
	current int
	rec *Record
}

func (p *Parser) Parse() *Record {
	p.rec = new(Record)
	c, err := hex.DecodeString(string(p.input[1:]))
	if err != nil {
		p.throwError("encoding/hex: invalid byte")
	}
	p.decodedBytes = c

	p.checkInputString()

	p.rec.ByteCount = p.getByteCount()
	p.rec.Address = p.getAddress()
	p.rec.Type = RecordType(p.getType())
	p.rec.Data = p.getData()

	return p.rec
}

func (p *Parser) checkInputString() {
	if !p.checkMarker() {
		p.throwError("Invalid start of Intel HEX record! Expected \":\" got: %q", p.input[0])
	}

	if !p.allowedChars([]byte("0123456789ABCDEF"), p.input[1:]) {
		p.throwError("Not allowed characters in Intel HEX record! Expected not 0123456789ABCDEF got: %s", p.input)
	}

	if len(p.input) < 11 {
		p.throwError("Invalid length of Intel HEX record! Expected not less than 11 got: %q", len(p.input))
	}

	checksum := p.generateCheckSum()
	if p.getCheckSum() != checksum {
		p.throwError("Invalid Intel HEX record checksum! Expected %s got: %s", p.getCheckSum(), checksum)
	}
}

func (p *Parser) getCheckSum() byte {
	n := p.decodedBytes[len(p.decodedBytes) - 1:]
	return n[0]

}

func (p *Parser) generateCheckSum() (sum byte) {
	n := p.decodedBytes[:len(p.decodedBytes) - 1]
	for _, value := range n {
		sum += value
	}
	return (^sum) + 1
}

func (p *Parser) throwError(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

func (p *Parser) checkMarker() bool {
	if p.input[0] == byte(':') {
			return true
		} else {
			return false
		}
	return false
}

func (p *Parser) getByteCount() (count int) {
	if p.Next() {
		count = int(p.Value())
	}
	return

}

func (p *Parser) getAddress() (addr int64) {
	b := make([]byte, 0 , 2)
	for i := 0; i < 2 ; i++ {
		if p.Next() {
			b = append(b,p.Value())
		}
	}
	addr = int64(((b[0] & 0xFF) << 8) + (b[1] & 0xFF))
	return
}

func (p *Parser) getType() (rtype int8) {
	if p.Next() {
		rtype = int8(p.Value())
	}
	return
}

func (p *Parser) getData() (data []byte) {
	for i := 0; i < p.rec.ByteCount; i++ {
		if p.Next() {
			data = append(data, p.Value())
		}
	}
	return
}

func (p *Parser) allowedChars(c []byte, b []byte) bool {
	var count int
	for _, i := range b {
		for _, s := range c {
			if s == i {
				count++
			}
		}
	}
	if count >= len(b) {
		return true
	} else {
		return false
	}
}

func (p *Parser) Next() bool {
	p.current++
	if p.current >= len(p.decodedBytes) {
		return false
	}

	return true
}

func (p *Parser) Value() byte {
	return p.decodedBytes[p.current]
}



