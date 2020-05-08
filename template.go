package main

import (
	"bytes"

	"github.com/tidusant/chadmin-repo/models"

	"github.com/tidusant/c3m-common/c3mcommon"
	"github.com/tidusant/c3m-common/mycrypto"

	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"strings"

	"golang.org/x/net/html"

	"github.com/spf13/viper"
)

var (
	Carts map[string]models.Product
)

func getNodeByAttr(attr string, n *html.Node) (element *html.Node, val string) {
	val = ""
	for _, a := range n.Attr {
		if a.Key == attr {
			return n, a.Val
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if element, val = getNodeByAttr(attr, c); val != "" {
			return
		}
	}
	return
}
func renderHtml(n *html.Node) string {
	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, n)
	return buf.String()
}
func renderInnerHtml(n *html.Node) string {
	htmlreturn := ""
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		htmlreturn += renderHtml(c)
	}

	return htmlreturn
}

func GetCarts() []models.Product {
	var cartarr []models.Product
	for _, v := range Carts {
		cartarr = append(cartarr, v)

	}
	return cartarr
}
func SubmitTemplate(name string, code string) string {
	//var templ models.Template

	//check template exist:
	var templfolder = wokingfolder + "/" + name
	if _, err := os.Stat("./" + templfolder); err != nil {
		if os.IsNotExist(err) {
			// file does not exist
			return "template not found!"
		} else {
			// other error
		}
	}

	// inputname := "resources/lang.txt"
	// //read file
	// b, err := ioutil.ReadFile(templfolder + "/" + inputname)
	// if err != nil {
	// 	return fmt.Sprintf("cannot read file %s!", inputname)
	// }
	// resourceslang := string(b)
	// //convert to \n line
	// resourceslang = strings.Replace(resourceslang, "\r\n", "\n", -1)
	// resourceslang = strings.Replace(resourceslang, "\r", "\n", -1)

	// inputname = "resources/config.txt"
	// //read file
	// b, err = ioutil.ReadFile(templfolder + "/" + inputname)
	// if err != nil {
	// 	return fmt.Sprintf("cannot read file %s!", inputname)
	// }
	// resourcesconfig := string(b)
	// //convert to \n line
	// resourcesconfig = strings.Replace(resourcesconfig, "\r\n", "\n", -1)
	// resourcesconfig = strings.Replace(resourcesconfig, "\r", "\n", -1)

	// inputname = "screenshot.jpg"
	// b, err = ioutil.ReadFile(templfolder + "/" + inputname)
	// //screenshotstr := base64.StdEncoding.EncodeToString(b)
	// if err != nil {
	// 	return fmt.Sprintf("cannot read file %s!", inputname)
	// }
	// if len(b) < 512 {
	// 	return c3mcommon.ReturnJsonMessage("0", "screenshot is not correct", "", "")
	// }
	// filetype := http.DetectContentType(b[:512])
	// if filetype != "image/jpeg" && filetype != "image/bmp" && filetype != "image/png" && filetype != "image/gif" && filetype != "image/webp" && filetype != "image/vnd.microsoft.icon" {

	// 	return c3mcommon.ReturnJsonMessage("0", "invalid screenshot image type", "", "")
	// }

	// // //pagehtml content
	// // pagetmpl := make(map[string]string)
	// // pagefiles, _ := ioutil.ReadDir("./" + templfolder + "/resources/pages")
	// // for _, f := range pagefiles {
	// // 	if !f.IsDir() {
	// // 		if filepath.Ext(f.Name()) == ".txt" {
	// // 			b, err := ioutil.ReadFile(templfolder + "/resources/pages/" + f.Name())
	// // 			if err != nil {
	// // 				return fmt.Sprintf("cannot read file %s!", f.Name())
	// // 			}
	// // 			// win16be := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
	// // 			// utf16bom := unicode.BOMOverride(win16be.NewDecoder())
	// // 			// unicodeReader := transform.NewReader(bytes.NewReader(b), utf16bom)
	// // 			// content, _ := ioutil.ReadAll(unicodeReader)
	// // 			// log.Debugf("Pagecontent %s - %s", content)

	// // 			strcontent := base64.StdEncoding.EncodeToString(b)
	// // 			pagetmpl[strings.Replace(f.Name(), ".txt", "", 1)] = strcontent
	// // 		}
	// // 	}
	// // }

	// // pagebytes, err := json.Marshal(pagetmpl)
	// // if err != nil {
	// // 	return fmt.Sprintf("error parsing json template !%s", err)
	// // }

	// //page resources:
	// pageresources := make(map[string]map[string]string)
	// tmpl := make(map[string]string)
	// folders, err := ioutil.ReadDir(templfolder + "/resources/pages")
	// if err == nil {
	// 	for _, d := range folders {
	// 		if !d.IsDir() {
	// 			continue
	// 		}
	// 		files, err := ioutil.ReadDir(templfolder + "/resources/pages/" + d.Name())
	// 		if err == nil {
	// 			fileresources := make(map[string]string)
	// 			for _, f := range files {
	// 				if f.Name()[len(f.Name())-4:] != ".txt" {
	// 					continue
	// 				}
	// 				b, err := ioutil.ReadFile(templfolder + "/resources/pages/" + d.Name() + "/" + f.Name())
	// 				if err != nil {
	// 					return fmt.Sprintf("cannot read file %s!", f.Name())
	// 				}
	// 				filecontent := string(b)
	// 				//convert to \n line
	// 				filecontent = strings.Replace(filecontent, "\r\n", "\n", -1)
	// 				filecontent = strings.Replace(filecontent, "\r", "\n", -1)
	// 				fileresources[strings.Replace(f.Name(), ".txt", "", 1)] = filecontent
	// 			}
	// 			pageresources[d.Name()] = fileresources
	// 			b, _ = ioutil.ReadFile(wokingfolder + "/" + name + "/resources/pages/" + d.Name() + "/" + d.Name() + ".html")
	// 			tmpl[d.Name()] = string(b)
	// 		}
	// 	}
	// }
	// //html content
	// files, _ := ioutil.ReadDir(templfolder)
	// for _, f := range files {
	// 	if !f.IsDir() {
	// 		if filepath.Ext(f.Name()) == ".html" {
	// 			b, err := ioutil.ReadFile(templfolder + "/" + f.Name())
	// 			if err != nil {
	// 				c3mcommon.CheckError(fmt.Sprintf("cannot read file %s!", f.Name()), err)
	// 				continue
	// 			}
	// 			tmpl[strings.Replace(f.Name(), `.html`, "", 1)] = string(b)
	// 		}
	// 	}
	// }

	// //pageresourcebytes, err := json.Marshal(pageresources)
	// //tmplbytes, err := json.Marshal(tmpl)
	// if err != nil {
	// 	return fmt.Sprintf("error parsing json pageresources !%s", err)
	// }

	// //css content
	// cssfiles := make(map[string]string)
	// b, err = ioutil.ReadFile(templfolder + "/index.html")
	// header := string(b)
	// RegExp := regexp.MustCompile(`<link.*?href="(.*?)"`)
	// Matches := RegExp.FindAllStringSubmatch(header, -1)
	// for i := 0; i < len(Matches); i++ {

	// 	path := strings.Replace(Matches[i][1], "{{Templateurl}}", "", -1)
	// 	if filepath.Ext(path) != ".css" {
	// 		continue
	// 	}

	// 	//read file
	// 	b, err = ioutil.ReadFile(templfolder + "/" + path)
	// 	if err != nil {
	// 		return fmt.Sprintf("cannot read file %s!", path)
	// 	}
	// 	cssfiles[path] = string(b)
	// }
	// //cssbytes, err := json.Marshal(cssfiles)
	// if err != nil {
	// 	return fmt.Sprintf("error parsing json cssstr !%s", err)
	// }
	// //script content
	// scriptfiles := make(map[string]string)
	// RegExp = regexp.MustCompile(`<script.*?src="(.*?)"`)
	// Matches = RegExp.FindAllStringSubmatch(header, -1)
	// for i := 0; i < len(Matches); i++ {
	// 	path := strings.Replace(Matches[i][1], "{{Templateurl}}", "", -1)

	// 	//read file
	// 	b, err = ioutil.ReadFile(templfolder + "/" + path)
	// 	if err != nil {
	// 		return fmt.Sprintf("cannot read file %s!", path)
	// 	}
	// 	scriptstr := string(b)
	// 	scriptfiles[path] = scriptstr
	// }
	// //scriptbytes, err := json.Marshal(scriptfiles)
	// if err != nil {
	// 	return fmt.Sprintf("error parsing json scriptfiles !%s", err)
	// }
	// //model js content
	// modelfiles := make(map[string]string)
	// folders, err = ioutil.ReadDir(wokingfolder + "/" + name + "/js/models")
	// if err == nil {
	// 	for _, f := range folders {
	// 		if f.IsDir() {
	// 			continue
	// 		}
	// 		//read file
	// 		b, err = ioutil.ReadFile(wokingfolder + "/" + name + "/js/models/" + f.Name())
	// 		if err != nil {
	// 			return fmt.Sprintf("cannot read file %s!", f.Name())
	// 		}
	// 		modelfiles[f.Name()] = string(b)
	// 	}
	// }
	// //modelbytes, err := json.Marshal(modelfiles)
	// if err != nil {
	// 	return fmt.Sprintf("error parsing json scriptfiles !%s", err)
	// }

	// //images content
	// // imagefiles := make(map[string]string)
	// // imagefolder := "./" + templfolder + "/images"

	// // imagefiles, strerr := ReadImageFolder(imagefolder, imagefolder, imagefiles)
	// // if strerr != "" {
	// // 	return fmt.Sprintf("error parsing images !%s", strerr)
	// // }
	// // imagebytes, err := json.Marshal(imagefiles)
	// // if err != nil {
	// // 	return fmt.Sprintf("error parsing images !%s", err)
	// // }

	// //fonts content
	// fontfiles := make(map[string]string)

	// files, _ = ioutil.ReadDir("./" + templfolder + "/fonts")
	// for _, f := range files {
	// 	if !f.IsDir() {
	// 		// fi, err := os.Open("./" + templfolder + "/fonts" + "/" + f.Name())
	// 		// if err != nil {
	// 		// 	panic(err)
	// 		// }
	// 		// defer fi.Close()

	// 		// // read into bufferv
	// 		// var b []byte
	// 		// buffer := bufio.NewReader(fi)
	// 		// _, err = buffer.WriteTo(b)

	// 		b, err := ioutil.ReadFile(templfolder + "/fonts" + "/" + f.Name())
	// 		base64Str := base64.StdEncoding.EncodeToString(b)
	// 		if err != nil {
	// 			return fmt.Sprintf("cannot read file %s!", f.Name())
	// 		}
	// 		fontfiles[f.Name()] = base64Str
	// 	}
	// }
	// //fontbytes, err := json.Marshal(fontfiles)

	// if err != nil {
	// 	return fmt.Sprintf("error parsing fontfiles !%s", err)
	// }

	datasubmit := make(map[string]string)
	datasubmit["Code"] = code
	//datasubmit["Content"] = string(tmplbytes)

	//datasubmit["Models"] = string(modelbytes)
	//datasubmit["Pages"] = string(pageresourcebytes)
	//datasubmit["CSS"] = string(cssbytes)
	//datasubmit["Script"] = string(scriptbytes)
	//datasubmit["Images"] = string(imagebytes)
	//datasubmit["Images"] = ""
	//datasubmit["Fonts"] = string(fontbytes)
	datasubmit["Title"] = name
	//datasubmit["Screenshot"] = screenshotstr
	//datasubmit["Configs"] = resourcesconfig
	//datasubmit["Resources"] = resourceslang

	datastr, err := json.Marshal(datasubmit)
	if err != nil {
		return fmt.Sprintf("error parsing json data !%s", err)
	}

	decodedata := mycrypto.EncodeBK(string(datastr), "submit")

	///test
	// params := mycrypto.DecodeBK(decodedata, "submit")
	// var templ models.TemplateSubmit

	// err = json.Unmarshal([]byte(params), &templ)
	// if !c3mcommon.CheckError("template parse json", err) {
	// 	log.Debugf(params)
	// }

	// var pages map[string]string
	// json.Unmarshal([]byte(templ.Pages), &pages)
	// for code, content := range pages {
	// 	content = mycrypto.Base64Decode(content)
	// 	log.Debugf("page content %s - %s", code, content)
	// }

	//====
	// data := url.Values{}
	// data.Add("data", decodedata)
	// data.Add("key", mycc.Key)
	zipfile := "./" + wokingfolder + "/" + name + ".zip"
	err = c3mcommon.Zipit(templfolder, zipfile)
	if !c3mcommon.CheckError("template error", err) {
		return err.Error()
	}
	extraParams := map[string]string{
		"key":  mycrypto.EncDat2(userkey),
		"data": decodedata,
	}

	submiturl := viper.GetString("config.buildserver") + mycrypto.EncodeBK("submit", "name")
	//rtstr := c3mcommon.RequestUrl2(submiturl, "POST", data)
	rtstr := c3mcommon.FileUploadRequest(submiturl, extraParams, "file", zipfile)
	rtstr = mycrypto.DecodeLight1(rtstr, 5)
	os.Remove(zipfile)

	var rs models.RequestResult
	//log.Debugf("response str %s", rtstr)
	err = json.Unmarshal([]byte(rtstr), &rs)

	if rs.Status != "1" {
		return fmt.Sprintf("error! %s", rs.Error)
	}
	//rptmpl.SaveTemplate(cat)
	return ""
}

func ReadImageFolder(orgfolder, folder string, imagefiles map[string]string) (map[string]string, string) {

	files, _ := ioutil.ReadDir(folder)
	for _, f := range files {
		if !f.IsDir() {
			fileext := filepath.Ext(f.Name())
			if fileext == ".jpg" || fileext == ".jpeg" || fileext == ".png" || fileext == ".gif" || fileext == ".svg" {
				b, err := ioutil.ReadFile(folder + "/" + f.Name())
				imgBase64Str := base64.StdEncoding.EncodeToString(b)

				if err != nil {
					return imagefiles, fmt.Sprintf("cannot read file %s!", folder+"/"+f.Name())
				}

				imagefiles[strings.Replace(folder, orgfolder, "", 1)+"/"+f.Name()] = imgBase64Str
			}
		} else {
			imagefiles, strerr := ReadImageFolder(orgfolder, folder+"/"+f.Name(), imagefiles)
			if strerr != "" {
				return imagefiles, strerr
			}
		}
	}
	return imagefiles, ""
}
