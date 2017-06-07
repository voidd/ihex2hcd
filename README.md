# ihex2hcd

## Description
Broadcom Bluetooth firmware converter

## Usage 


``` golang
package main


import
(
	"fmt"
	"os"
	"ihex2hcd"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("list or count arguments is required")
		os.Exit(1)
	}
	inFile, _ := os.Open(os.Args[1])
	outFile, _ := os.Create(os.Args[2])

	defer inFile.Close()

	b := ihex2hcd.New(inFile)
	b.BinOutput(outFile)
	// or
	a := b.RecordOutput()
	for _, r := range a {
		fmt.Println(r.Data)
	}
	// or
	b.StringOutput()
}
