package general
import (
	"io/ioutil"
	"net/http"
	"bytes"
)

/*
 * general http api client for making requests to specified URIs with specified methods and headers.
 */
func HTTPRequest( url, method string, data []byte, headers map[string]string ) ([]byte, error) {
	req, err := http.NewRequest( method, url, bytes.NewBuffer(data) )
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	cli := &http.Client{}
	resp, err := cli.Do( req )
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	defer resp.Body.Close()
	newData, err := ioutil.ReadAll( resp.Body )
	if err != nil {
		return nil, err
	}
	return newData, nil
}
