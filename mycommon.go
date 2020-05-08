package main

import (
	"bytes"

	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"

	"encoding/base64"
	"encoding/json"
	"io/ioutil"

	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/tidusant/c3m-common/c3mcommon"
	"github.com/tidusant/c3m-common/log"
	"github.com/tidusant/c3m-common/mycrypto"
	"github.com/tidusant/chadmin-repo/models"

	"github.com/spf13/viper"
)

type MyCommon struct {
	Key string
}

var v *MyCommon

func (v *MyCommon) SetKey(key string) {
	v.Key = key
}

func (v *MyCommon) RequestServer(name string, data url.Values) models.RequestResult {
	timestart := time.Now()
	rt := models.RequestResult{"0", "unknow error", "", json.RawMessage{}}
	name = v.Encode(name, "name")
	if data == nil {
		data = url.Values{}
	}
	data.Add("key", v.Key)

	if strings.Index(name, "http://") == -1 && strings.Index(name, "https://") == -1 {
		name = "http://" + viper.GetString("apiserver.h") + "/" + name
	}
	if viper.GetString("config.proxy") != "" {
		proxyUrl, _ := url.Parse(viper.GetString("config.proxy"))
		http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
	}

	rsp, err := http.PostForm(name, data)
	if !c3mcommon.CheckError("request api", err) {
		rt.Message = "request server fail"
		return rt
	}
	defer rsp.Body.Close()
	log.Debug("request server time:%s", time.Since(timestart))
	rtbyte, err := ioutil.ReadAll(rsp.Body)
	c3mcommon.CheckError("request read data", err)
	//log.Debugf("string(rtbyte) %s", string(rtbyte))
	rtstr := mycrypto.DecodeLight1(string(rtbyte), 5)
	log.Debug("DecodeLight1 time:%s", time.Since(timestart))
	if rtstr == "" {
		return rt
	}

	//log.Debugf("data %s:", data)
	var rs models.RequestResult
	//log.Debugf("response str %s", rtstr)
	err = json.Unmarshal([]byte(rtstr), &rs)

	if !c3mcommon.CheckError("parse json", err) {
		rs.Status = "0"
		rs.Error = "parse json response fail"
		return rs
	}

	return rs
}

func (v *MyCommon) Encode(data string, keysalt string) string {
	//data = lzjs.CompressToBase64(data)
	data = base64.StdEncoding.EncodeToString([]byte(data))
	data = strings.Replace(data, "=", "", -1)
	//log.Debugf("keysalt: %s", keysalt)
	keysalt = base64.StdEncoding.EncodeToString([]byte(keysalt))
	keysalt = strings.Replace(keysalt, "=", "", -1)
	//log.Debugf("keysalt: %s", keysalt)
	l := 3
	data = data[:l] + keysalt + data[l:]
	//log.Println("strReturn: %s", data)
	return data
}

//Decode Old

func (v *MyCommon) Decode(code string) string {
	timestart := time.Now()
	if code == "" {
		return code
	}
	var rt string = ""
	key := code
	//key = "kZXUuYkRWzUgQk92YoNwRdh92Q3SZtFmb9Wa0NW"
	if key == rt {
		return rt
	}

	oddstr := "d"
	l := int(math.Floor(float64(len(key)-2) / 3))
	//log.Printf("len %d", key)
	//log.Printf("len %d", len(key))
	//log.Printf("1/2 len %d", l)
	num := key[l : l+2]

	key = key[:l] + key[l+2:]

	byteDecode, _ := base64.StdEncoding.DecodeString(mycrypto.Base64fix(num))
	num = string(byteDecode)
	//log.Printf("num %s", num)
	floatNum, _ := strconv.ParseFloat(num, 64)
	intNum := (int)(floatNum)
	if intNum > 0 {
		//print_r($num);print_r("\r\n");
		//get odd string
		lf := math.Ceil(float64(len(key)) / floatNum)
		oddstr = key[:int(lf)]
		ukey := strings.Replace(key, oddstr, "", 1)
		log.Printf("Decode replace:%s", time.Since(timestart))

		oddb := []byte(oddstr)
		ukeyb := []byte(ukey)
		var base64b []byte
		for i := 0; i < len(oddb); i++ {
			base64b = append(base64b, oddb[len(oddb)-1-i])

			//for ukey
			for j := 0; j < intNum-1; j++ {
				index := len(ukeyb) - 1 - j - (i * (intNum - 1))
				if index >= 0 {

					base64b = append(base64b, ukeyb[index])

				} else {
					break
				}
			}

		}
		//log.Printf("Decode loop:%s", time.Since(timestart))

		base64str := string(base64b)
		base64str = base64str[:len(base64str)-intNum]
		base64str = mycrypto.Base64fix(base64str)
		byteDecode3, _ := base64.StdEncoding.DecodeString(base64str)
		rt = string(byteDecode3)
		return rt

	}
	return rt
}

func (v *MyCommon) CreateImageFile(path, b64content string) error {
	unbased, err := base64.StdEncoding.DecodeString(b64content)
	if err != nil {
		log.Debugf("Cannot decode b64  %s", err)
		return err
	}
	r := bytes.NewReader(unbased)

	if filepath.Ext(path) == ".png" {
		im, err := png.Decode(r)
		if err != nil {
			log.Debugf("Bad png  %s", err)
			return err
		}

		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0777)
		if err != nil {

			log.Debugf("Cannot open file  %s", err)
			return err
		}
		png.Encode(f, im)
		f.Close()
	} else if filepath.Ext(path) == ".jpg" || filepath.Ext(path) == ".jpeg" {
		im, err := jpeg.Decode(r)
		if err != nil {
			log.Debugf("Bad jpg  %s", err)
			return err
		}

		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0777)
		if err != nil {
			log.Debugf("Cannot open file  %s", err)
			return err
		}
		jpeg.Encode(f, im, nil)
		f.Close()
	} else if filepath.Ext(path) == ".gif" {
		im, err := gif.Decode(r)
		if err != nil {
			log.Debugf("Bad gif  %s", err)
			return err
		}

		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0777)
		if err != nil {

			log.Debugf("Cannot open file  %s", err)
			return err
		}
		gif.Encode(f, im, nil)
		f.Close()
	}
	return nil
}

const (
	encodePath encoding = 1 + iota
	encodeHost
	encodeUserPassword
	encodeQueryComponent
	encodeFragment
)

type encoding int
type EscapeError string

func (e EscapeError) Error() string {
	return "invalid URL escape " + strconv.Quote(string(e))
}

func ishex(c byte) bool {
	switch {
	case '0' <= c && c <= '9':
		return true
	case 'a' <= c && c <= 'f':
		return true
	case 'A' <= c && c <= 'F':
		return true
	}
	return false
}

func unhex(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}

// Return true if the specified character should be escaped when
// appearing in a URL string, according to RFC 3986.
//
// Please be informed that for now shouldEscape does not check all
// reserved characters correctly. See golang.org/issue/5684.
func shouldEscape(c byte, mode encoding) bool {
	// §2.3 Unreserved characters (alphanum)
	if 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' {
		return false
	}

	if mode == encodeHost {
		// §3.2.2 Host allows
		//  sub-delims = "!" / "$" / "&" / "'" / "(" / ")" / "*" / "+" / "," / ";" / "="
		// as part of reg-name.
		// We add : because we include :port as part of host.
		// We add [ ] because we include [ipv6]:port as part of host
		switch c {
		case '!', '$', '&', '\'', '(', ')', '*', '+', ',', ';', '=', ':', '[', ']':
			return false
		}
	}

	switch c {
	case '-', '_', '.', '~': // §2.3 Unreserved characters (mark)
		return false

	case '$', '&', '+', ',', '/', ':', ';', '=', '?', '@': // §2.2 Reserved characters (reserved)
		// Different sections of the URL allow a few of
		// the reserved characters to appear unescaped.
		switch mode {
		case encodePath: // §3.3
			// The RFC allows : @ & = + $ but saves / ; , for assigning
			// meaning to individual path segments. This package
			// only manipulates the path as a whole, so we allow those
			// last two as well. That leaves only ? to escape.
			return c == '?'

		case encodeUserPassword: // §3.2.1
			// The RFC allows ';', ':', '&', '=', '+', '$', and ',' in
			// userinfo, so we must escape only '@', '/', and '?'.
			// The parsing of userinfo treats ':' as special so we must escape
			// that too.
			return c == '@' || c == '/' || c == '?' || c == ':'

		case encodeQueryComponent: // §3.4
			// The RFC reserves (so we must escape) everything.
			return true

		case encodeFragment: // §4.1
			// The RFC text is silent but the grammar allows
			// everything, so escape nothing.
			return false
		}
	}

	// Everything else must be escaped.
	return true
}

func escape(s string, mode encoding) string {
	spaceCount, hexCount := 0, 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscape(c, mode) {
			if c == ' ' && mode == encodeQueryComponent {
				spaceCount++
			} else {
				hexCount++
			}
		}
	}

	if spaceCount == 0 && hexCount == 0 {
		return s
	}

	t := make([]byte, len(s)+2*hexCount)
	j := 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == ' ' && mode == encodeQueryComponent:
			t[j] = '+'
			j++
		case shouldEscape(c, mode):
			t[j] = '%'
			t[j+1] = "0123456789ABCDEF"[c>>4]
			t[j+2] = "0123456789ABCDEF"[c&15]
			j += 3
		default:
			t[j] = s[i]
			j++
		}
	}
	return string(t)
}

// unescape unescapes a string; the mode specifies
// which section of the URL string is being unescaped.
func unescape(s string, mode encoding) (string, error) {
	// Count %, check that they're well-formed.
	n := 0
	hasPlus := false
	for i := 0; i < len(s); {
		switch s[i] {
		case '%':
			n++
			if i+2 >= len(s) || !ishex(s[i+1]) || !ishex(s[i+2]) {
				s = s[i:]
				if len(s) > 3 {
					s = s[:3]
				}
				return "", EscapeError(s)
			}
			i += 3
		case '+':
			hasPlus = mode == encodeQueryComponent
			i++
		default:
			i++
		}
	}

	if n == 0 && !hasPlus {
		return s, nil
	}

	t := make([]byte, len(s)-2*n)
	j := 0
	for i := 0; i < len(s); {
		switch s[i] {
		case '%':
			t[j] = unhex(s[i+1])<<4 | unhex(s[i+2])
			j++
			i += 3
		case '+':
			if mode == encodeQueryComponent {
				t[j] = ' '
			} else {
				t[j] = '+'
			}
			j++
			i++
		default:
			t[j] = s[i]
			j++
			i++
		}
	}
	return string(t), nil
}

func EncodeUriComponent(rawString string) string {
	return escape(rawString, encodeFragment)
}

func DecodeUriCompontent(encoded string) (string, error) {
	return unescape(encoded, encodeQueryComponent)
}
