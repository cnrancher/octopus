package testdata

import (
	"io/ioutil"

	"github.com/onsi/ginkgo"
	"github.com/pkg/errors"

	"github.com/rancher/octopus/pkg/util/converter"
)

// LoadBytes gets the content of file that located in relative `testdata` directory.
func LoadBytes(filename string) ([]byte, error) {
	var content, err = ioutil.ReadFile("testdata/" + filename)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load testdata file")
	}
	return content, nil
}

// LoadString gets the string content of file that located in relative `testdata` directory.
func LoadString(filename string) (string, error) {
	var c, err = LoadBytes(filename)
	if err != nil {
		return "", err
	}
	return converter.UnsafeBytesToString(c), nil
}

// MustLoadBytes gets the content of file that located in relative `testdata` directory,
// and prints/panics the error log if failed.
func MustLoadBytes(filename string, l ginkgo.GinkgoTInterface) []byte {
	var c, err = LoadBytes(filename)
	if err != nil {
		if l == nil {
			panic(err)
		}
		l.Fatal(err)
	}
	return c
}

// MustLoadString gets the string content of file that located in relative `testdata` directory,
// and prints/panics the error log if failed.
func MustLoadString(filename string, l ginkgo.GinkgoTInterface) string {
	var c, err = LoadString(filename)
	if err != nil {
		if l == nil {
			panic(err)
		}
		l.Fatal(err)
	}
	return c
}
