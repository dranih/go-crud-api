package utils

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
)

func GzCompress(input string) (string, error) {
	var output bytes.Buffer
	gzwriter := gzip.NewWriter(&output)
	_, err := gzwriter.Write([]byte(input))
	if err != nil {
		log.Printf("Error compress : %s", err)
		return "", err
	}
	gzwriter.Close()
	return output.String(), nil
}

func GzUncompress(input string) (string, error) {
	reader := bytes.NewReader([]byte(input))
	gzreader, err := gzip.NewReader(reader)
	if err != nil {
		log.Printf("Error uncompress : %s", err)
		return "", err
	}
	output, err := ioutil.ReadAll(gzreader)
	if err != nil {
		log.Printf("Error uncompress : %s", err)
		return "", err
	}
	return string(output), nil
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func NumberFormat(number float64, dec int, decPoint, thousandsSep string) string {
	neg := false
	if number < 0 {
		number = -number
		neg = true
	}
	//dec := int(decimals)
	// Will round off
	str := fmt.Sprintf("%."+strconv.Itoa(dec)+"F", number)
	prefix, suffix := "", ""
	if dec > 0 {
		prefix = str[:len(str)-(dec+1)]
		suffix = str[len(str)-dec:]
	} else {
		prefix = str
	}
	sep := []byte(thousandsSep)
	n, l1, l2 := 0, len(prefix), len(sep)
	// thousands sep num
	c := (l1 - 1) / 3
	tmp := make([]byte, l2*c+l1)
	pos := len(tmp) - 1
	for i := l1 - 1; i >= 0; i, n, pos = i-1, n+1, pos-1 {
		if l2 > 0 && n > 0 && n%3 == 0 {
			for j := range sep {
				tmp[pos] = sep[l2-j-1]
				pos--
			}
		}
		tmp[pos] = prefix[i]
	}
	s := string(tmp)
	if dec > 0 {
		s += decPoint + suffix
	}
	if neg {
		s = "-" + s
	}

	return s
}
