package cherry

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	MockPath = "./mock/kodu.html"
)

func readFile(path string) io.Reader {

	f, err := os.Open(path)
	if err != nil {
		fmt.Println("Failed to read the mock data")
	}

	return bufio.NewReader(f)
}

func TestParser(t *testing.T) {

	fmt.Println("Read mock file", MockPath)

	reader := readFile(MockPath)
	result, _ := ParseFromReader("Kodu&Aed", reader)

	firstItem := &Offer{
		"Päikesevari Sinu suvisesse aeda või randa",
		"/triibuline-paikesevari",
		9.99,
		29.99,
		0,
		12,
		568579,
	}
	assert.EqualValues(t, firstItem, result.Offers[0], "Päiksevarju pakkumine")

	secondItem := &Offer{
		"Liurada laste lustimiseks palaval suvepäeval",
		"/vinuulist-liurada",
		9.99,
		14.99,
		0,
		6,
		568579,
	}
	assert.EqualValues(t, secondItem, result.Offers[1], "Liurada laste")
}

// func TestMultiTagParse(t *testing.T) {
//
// 	reader := readFile(MockPath)
// 	doc, err := goquery.NewDocumentFromReader(reader)
// 	if err != nil {
// 		t.Fatal("Failed to read document", err)
// 	}
//
// 	filterScript := func(index int, s *goquery.Selection) bool {
//
// 		fmt.Println(index, s.Text())
// 		return strings.Contains(s.Text(), "timeleft_cache")
// 	}
//
// 	results := doc.Find("script").FilterFunction(filterScript).First()
//
// 	fmt.Println("Results", results.Text())
// }
