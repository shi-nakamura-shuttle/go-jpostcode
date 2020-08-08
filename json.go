//go:generate statik -src=./jpostcode-data/data/json
package jpostcode

import (
	"io"
	"os"
	"reflect"

	"github.com/goccy/go-json"
	"github.com/mitchellh/mapstructure"
	"github.com/rakyll/statik/fs"
	_ "github.com/syumai/go-jpostcode/statik"
)

func searchAddressesFromJSON(postCode string) ([]*Address, error) {
	var addresses []*Address
	if len(postCode) != 7 {
		return nil, ErrInvalidArgument
	}
	firstPostCode := postCode[0:3]
	secondPostCode := postCode[3:7]
	dataFile, err := openDataFile("/" + firstPostCode + ".json")
	defer dataFile.Close()
	if err != nil {
		if err != os.ErrNotExist {
			return addresses, nil
		}
		return nil, err
	}
	var addressMap map[string]interface{}
	if err := json.NewDecoder(dataFile).Decode(&addressMap); err != nil {
		return nil, err
	}
	addressData, ok := addressMap[secondPostCode]
	if !ok {
		return addresses, nil
	}
	switch reflect.TypeOf(addressData).Kind() {
	case reflect.Slice:
		rawAddrs, ok := addressData.([]interface{})
		if !ok {
			return nil, ErrInternal
		}
		for _, rawAddr := range rawAddrs {
			addr, err := convertJSONToAddress(rawAddr)
			if err != nil {
				return nil, err
			}
			addresses = append(addresses, addr)
		}
	default:
		addr, err := convertJSONToAddress(addressData)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, addr)
	}
	return addresses, err
}

func convertJSONToAddress(input interface{}) (*Address, error) {
	var addr Address
	err := mapstructure.Decode(input, &addr)
	if err != nil {
		return nil, err
	}
	return &addr, nil
}

func openDataFile(filePath string) (io.ReadCloser, error) {
	staticFS, err := fs.New()
	if err != nil {
		return nil, err
	}
	return staticFS.Open(filePath)
}
