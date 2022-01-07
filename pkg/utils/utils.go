package utils

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"log"
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
