package bcg

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/url"
)

func HttpGet(url string) []byte {
	response, err := http.Get(url)
	if CheckError(err) {
		return nil
	}

	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	CheckError(err)
	return data
}
func HttpPostBody(api string, data []byte) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		DisableKeepAlives: true,
	}
	hc := &http.Client{Transport: tr}
	param := bytes.NewBuffer(data)
	body, err := hc.Post(api, "application/json; charset=utf-8", param)
	if CheckError(err) {
		return nil, err
	}
	defer body.Body.Close()
	return ioutil.ReadAll(body.Body)
}
func HttpPostForm(api string, params map[string]string) ([]byte, error) {
	data := make(url.Values)
	for key, val := range params {
		data[key] = []string{val}
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		DisableKeepAlives: true,
	}
	hc := &http.Client{Transport: tr}
	body, err := hc.PostForm(api, data)
	if CheckError(err) {
		return nil, err
	}
	defer body.Body.Close()
	return ioutil.ReadAll(body.Body)
}
func RandString(length int, base string) string {
	if base == "" {
		base = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789~@#$%^&*()_+-=,./<>?;:[]{}|"
	}
	var chars = []byte(base)
	if length == 0 {
		return ""
	}
	clen := len(chars)
	if clen < 2 || clen > 256 {
		panic("Wrong charset length for NewLenChars()")
	}
	maxrb := 255 - (256 % clen)
	b := make([]byte, length)
	r := make([]byte, length+(length/4)) // storage for random bytes.
	i := 0
	for {
		if _, err := rand.Read(r); err != nil {
			panic("Error reading random bytes: " + err.Error())
		}
		for _, rb := range r {
			c := int(rb)
			if c > maxrb {
				continue // Skip this number to avoid modulo bias.
			}
			b[i] = chars[c%clen]
			i++
			if i == length {
				return string(b)
			}
		}
	}
}
