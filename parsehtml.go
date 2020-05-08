package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tidusant/c3m-common/c3mcommon"
	"github.com/tidusant/c3m-common/log"
	"github.com/tidusant/chadmin-repo/models"
	"golang.org/x/net/html"
)

func RenderData(htmlstr string, data models.TemplateViewData) string {
	htmlstr = strings.Replace(htmlstr, "\r\n", "\n", -1)
	// htmlstr = strings.Replace(htmlstr, "\n", "", -1)
	// htmlstr = strings.Replace(htmlstr, "\r", "", -1)

	//parse template
	isRoot := false
	var csscontent string
	var scriptcontent string
	var reg = regexp.MustCompile(`{{template "(.*?)" .}}`)
	for {
		t := reg.FindAllStringSubmatch(htmlstr, -1)
		if len(t) == 0 {
			break
		}
		for _, v := range t {
			b, err := ioutil.ReadFile(data.TemplatePath + v[1])

			if c3mcommon.CheckError(fmt.Sprintf("cannot read file %s!", v[1]), err) {
				htmltemplate := string(b)

				if v[1][:6] == "_start" {
					isRoot = true
					//				htmltemplate = strings.Replace(htmltemplate, `<script src="`, `{{scripts}}`+`<script src="`, 1)
					//css
					reg2 := regexp.MustCompile(`<link.*href="(.*?)".*>`)
					t2 := reg2.FindAllStringSubmatch(htmltemplate, -1)

					for _, v2 := range t2 {
						cssfile := strings.Replace(v2[1], "{{Templateurl}}", data.TemplatePath, -1)
						if filepath.Ext(cssfile) == ".css" {
							b, err := ioutil.ReadFile(cssfile)
							if c3mcommon.CheckError(fmt.Sprintf("cannot read file %s!", v2[1]), err) {
								csscontent += c3mcommon.MinifyCSS(b)
							}
						}
					}

					reg2 = regexp.MustCompile(`<link.*rel="stylesheet".*href="(.*?)".*>`)
					htmltemplate = reg2.ReplaceAllString(htmltemplate, "")
					reg2 = regexp.MustCompile(`<link.*href="(.*?)".*rel="stylesheet".*>`)
					htmltemplate = reg2.ReplaceAllString(htmltemplate, "")

					//javascript
					reg2 = regexp.MustCompile(`<script.*src="(.*?)".*<\/script>`)
					t2 = reg2.FindAllStringSubmatch(htmltemplate, -1)
					var jsfiles []string
					for _, v2 := range t2 {
						jsfiles = append(jsfiles, v2[1])
					}
					htmltemplate = reg2.ReplaceAllString(htmltemplate, "")
					// strings.Replace(htmltemplate, `<script.*<\/script>`, "", -1)

					for _, v3 := range jsfiles {
						scriptcontent += `<script src="` + v3 + `"></script>`
					}
					//scriptcontent += `<script src="{{Templateurl}}../scripts/jquery.js"></script>`
					scriptcontent += `<script src="{{Templateurl}}../scripts/base.js"></script>`

					b, _ := json.Marshal(data.Data)
					scriptcontent += `<script>
				var siteurl="{{siteurl}}";
				var Imageurl="{{Imageurl}}";
				var commonData=` + string(b) + `;`
					//myscript
					b, err = ioutil.ReadFile(data.TemplatePath + "js/myscript.js")
					myscript := string(b)
					if myscript != "" {
						//scan config token
						var reg2 = regexp.MustCompile(`{{(.*?)}}`)
						t2 := reg2.FindAllStringSubmatch(myscript, -1)
						for _, v2 := range t2 {
							valreplace := ""
							if _, ok := data.Configs[v2[1]]; ok {
								valreplace = data.Configs[v2[1]]
							}
							myscript = strings.Replace(myscript, `{{`+v2[1]+`}}`, valreplace, -1)
						}

						scriptcontent += myscript + `				
					$(document).ready(function(){
						
						msc.init();
					});`
					}
					scriptcontent += `</script>`
					htmltemplate = strings.Replace(htmltemplate, `</head>`, `<style>`+csscontent+`</style></head>`, 1)
					htmltemplate = strings.Replace(htmltemplate, `</head>`, scriptcontent+`</head>`, 1)
					htmltemplate = strings.Replace(htmltemplate, "{{Templateurl}}", data.Templateurl, -1)

				}

				htmlstr = strings.Replace(htmlstr, `{{template "`+v[1]+`" .}}`, htmltemplate, -1)
				//log.Debugf("parse root: %s", htmlstr)
			}
		}
	}
	if isRoot {
		//standard html
		r := bytes.NewReader([]byte(htmlstr))
		root, _ := html.Parse(r)
		htmlstr = renderHtml(root)
		htmlstr = html.UnescapeString(htmlstr)

	}

	//repeat content
	for {

		r := bytes.NewReader([]byte(htmlstr))
		root, _ := html.Parse(r)
		node, val := getNodeByAttr("ng-repeat", root)
		if val == "" {
			//no repeat content
			break
		}

		//convert html string to standard
		//log.Debugf("RenderData orgHtml %s", htmlstr)
		htmlloop := renderHtml(node)
		htmlloop = html.UnescapeString(htmlloop)
		//log.Debugf("RenderData htmlloop %s", htmlloop)
		htmlstr = strings.Replace(htmlstr, htmlloop, "{{repeat}}", 1)
		//log.Debugf("RenderData htmlstr %s", htmlstr)

		htmlrepeat := ""
		htmlloop = strings.Replace(htmlloop, ` ng-repeat="`+val+`"`, "", 1)
		//log.Debugf("RenderData val %s", val)

		args := strings.Split(val, " ")
		strs := make(map[string]string)
		if len(args) > 2 {
			var dataloop []json.RawMessage
			var err error
			argsloop := strings.Split(args[2], ".")
			querycond := ""
			if args[2] == "Carts()" {
				cartsb, _ := json.Marshal(GetCarts())

				err := json.Unmarshal(cartsb, &dataloop)
				c3mcommon.CheckError(fmt.Sprintf("pasre Carts json %s error ", args[2]), err)
			} else if len(argsloop) > 1 {

				parsers := ParseJsOperator(args[2], strs, data)

				err = json.Unmarshal([]byte(parsers), &dataloop)
				if args[2] == "Page.Langs" {
					log.Debugf("Page.Langs:%v ", parsers)
				}
				// argdata := make(map[string]json.RawMessage)
				// if len(argsloop) > 1 {
				// 	err = json.Unmarshal(data.Data[argsloop[0]], &argdata)
				// 	if c3mcommon.CheckError(fmt.Sprintf("pasre deeper json %s error ", argsloop[0]), err) {
				// 		err = json.Unmarshal(argdata["Items"], &dataloop)
				// 	}
				// }
			} else {
				//get query condition and remove
				regcond := regexp.MustCompile(`\w+\[(.+)\]`)
				tcond := regcond.FindAllStringSubmatch(args[2], -1)
				if len(tcond) > 0 {
					args[2] = strings.Replace(args[2], "["+tcond[0][1]+"]", "", -1)
					querycond = tcond[0][1]
				}
				err = json.Unmarshal(data.Data[args[2]], &dataloop)

			}
			// if val == "iimage,image in Page.Properties" {
			// 	//var t map[string]json.RawMessage
			// 	//err := json.Unmarshal(byte(data.Data), &t)
			// 	for k, _ := range data.Data["Page.Properties"] {
			// 		log.Debugf("data.Data %s: %s", k, string(data.Data["Page.Properties"][k]))
			// 	}
			// 	//log.Debugf("data.Data[args[2]] %s, %v, %v", string(data.Data[args[2]]), args[2])
			// }

			if c3mcommon.CheckError(fmt.Sprintf("pasre json %s error ", val), err) {
				//cond parse

				//
				loopindex := 0
				for _, v := range dataloop {
					if querycond != "" {
						//parse query condition
						condtype := "=="
						condargs := strings.Split(querycond, condtype)
						if len(condargs) == 1 {
							condtype = ">="
							condargs = strings.Split(querycond, condtype)
						}
						if len(condargs) == 1 {
							condtype = ">"
							condargs = strings.Split(querycond, condtype)
						}
						if len(condargs) == 1 {
							condtype = "<="
							condargs = strings.Split(querycond, condtype)
						}
						if len(condargs) == 1 {
							condtype = "<"
							condargs = strings.Split(querycond, condtype)
						}
						if len(condargs) == 1 {
							condtype = "!="
							condargs = strings.Split(querycond, condtype)
						}
						if len(condargs) > 1 {
							var propdata map[string]string
							json.Unmarshal(v, &propdata)
							condcompare := strings.Replace(condargs[1], "->", ".", -1)
							opcompare, strscompare := PrepairToParseOP(`"` + propdata[condargs[0]] + `"` + condtype + condcompare)
							rscompare := ParseJsBoolean(opcompare, strscompare, data)
							if !rscompare {
								continue
							}

						}
					}
					kv := strings.Split(args[0], ",")
					var indexstr json.RawMessage
					json.Unmarshal([]byte(`"`+strconv.Itoa(loopindex)+`"`), &indexstr)
					if len(kv) > 1 {
						data.Data[kv[0]] = indexstr
						data.Data[kv[1]] = v
					} else {
						data.Data["_index"] = indexstr
						data.Data[args[0]] = v
					}
					//log.Debugf("RenderData htmlrepeat begin %s", htmlrepeat)
					htmlrepeat += RenderData(htmlloop, data)
					loopindex++
					//log.Debugf("PrepairToParseOP ng-show _index: %v, args[0]: %v, val: %s", string(data.Data[kv[0]]), string(data.Data[kv[1]]))
					//log.Debugf("RenderData htmlrepeat end %s", htmlrepeat)

				}
			}
		}
		htmlstr = strings.Replace(htmlstr, "{{repeat}}", htmlrepeat, 1)

	}

	//show content
	for {

		r := bytes.NewReader([]byte(htmlstr))
		root, _ := html.Parse(r)

		node, val := getNodeByAttr("ng-show", root)
		if val == "" {
			break
		}

		op, strs := PrepairToParseOP(val)

		//log.Debugf("PrepairToParseOP ng-show op: %v, strs: %v, val: %s", op, strs, val)
		databool := ParseJsBoolean(op, strs, data)
		//log.Debugf("PrepairToParseOP ng-show _index: %v, args[0]: %v, val: %s", string(data.Data["i"]), string(data.Data["prod"]))
		htmlshow := renderHtml(node)
		htmlshow = html.UnescapeString(htmlshow)
		if databool {

			htmlreplace := strings.Replace(htmlshow, ` ng-show="`+val+`"`, "", 1)
			//log.Debugf("RenderData ng-show %s", val)
			//log.Debugf("RenderData htmlshow %s", htmlshow)
			//log.Debugf("RenderData htmlreplace begin %s", htmlreplace)
			htmlreplace = RenderData(htmlreplace, data)
			//log.Debugf("RenderData htmlreplace end %s", htmlreplace)
			htmlstr = strings.Replace(htmlstr, htmlshow, htmlreplace, 1)
			//log.Debugf("RenderData htmlreplace  %s", htmlstr)

		} else {
			htmlstr = strings.Replace(htmlstr, htmlshow, "", 1)
		}
	}

	//ng-src content
	for {

		r := bytes.NewReader([]byte(htmlstr))
		root, _ := html.Parse(r)

		_, val := getNodeByAttr("ng-src", root)
		if val == "" {
			break
		}

		op, strs := PrepairToParseOP(val)
		//log.Debugf("PrepairToParseOP:%s", op)
		dataval := ParseJsOperator(op, strs, data)
		//log.Debugf("dataval:%s", dataval)
		htmlstr = strings.Replace(htmlstr, ` ng-src="`+val+`"`, ` src="`+dataval+`"`, -1)
	}

	//render common data
	htmlstr = strings.Replace(htmlstr, `{{Templateurl}}`, data.Templateurl, -1)
	htmlstr = strings.Replace(htmlstr, `{{Imageurl}}`, data.Imageurl, -1)
	htmlstr = strings.Replace(htmlstr, `{{siteurl}}`, data.Siteurl, -1)
	//render data
	reg2 := regexp.MustCompile(`{{(.*?)}}`)

	for {
		t2 := reg2.FindAllStringSubmatch(htmlstr, -1)
		if len(t2) == 0 {
			break
		}

		v := t2[0]
		timestart := time.Now()
		op, strs := PrepairToParseOP(v[1])
		log.Debug("PrepairToParseOP time:%s", time.Since(timestart))
		datastr := ParseJsOperator(op, strs, data)
		log.Debug("ParseJsOperator time:%s", time.Since(timestart))
		htmlstr = strings.Replace(htmlstr, `{{`+v[1]+`}}`, datastr, -1)
		log.Debug("Replace time:%s", time.Since(timestart))
		//parse again to replaced htmlstr content's token

		//htmlstr = RenderData(htmlstr, data)
		log.Debug("RenderData time:%s", time.Since(timestart))

	}

	return htmlstr

}
func PrepairToParseOP(op string) (string, map[string]string) {
	//wrap string
	//op = html.UnescapeString(op)
	reg := regexp.MustCompile(`"(.*?)"`)
	t := reg.FindAllStringSubmatch(op, -1)
	strs := make(map[string]string)
	for _, v := range t {
		name := "string_" + strconv.Itoa(len(strs))
		strs[name] = v[1]
		op = strings.Replace(op, `"`+v[1]+`"`, "{{"+name+"}}", -1)
	}

	reg = regexp.MustCompile(`'(.*?)'`)
	t = reg.FindAllStringSubmatch(op, -1)
	for _, v := range t {
		name := "string_" + strconv.Itoa(len(strs))
		strs[name] = v[1]
		op = strings.Replace(op, `'`+v[1]+`'`, "{{"+name+"}}", -1)
	}

	//remove space
	reg2 := regexp.MustCompile(`\s+`)
	op = reg2.ReplaceAllString(op, ` `)
	reg2 = regexp.MustCompile(`\s?([\+\-\*\/,>=<!\|&])\s?`)
	op = reg2.ReplaceAllString(op, `$1`)
	reg2 = regexp.MustCompile(`\s\)`)
	op = reg2.ReplaceAllString(op, `)`)
	reg2 = regexp.MustCompile(`\(\s`)
	op = reg2.ReplaceAllString(op, `(`)
	return op, strs
}
func ParseJsOperator(op string, strs map[string]string, data models.TemplateViewData) string {
	//parse function

	// reg2 := regexp.MustCompile(`(\w+)\((.*?)\)`)
	// t4 := reg2.FindAllStringSubmatch(op, -1)
	// for _, v := range t4 {
	// 	name := "JsFunc_" + strconv.Itoa(len(strs))
	// 	strs[name] = ParseJsFunc(v[1], v[2], strs, data)
	// 	op = strings.Replace(op, v[1]+"("+v[2]+")", `{{`+name+`}}`, -1)
	// }
	op = strings.Trim(op, " ")
	for {
		reg2 := regexp.MustCompile(`(\w+)\(`)
		t3 := reg2.FindStringIndex(op)
		if len(t3) > 0 {

			fname := op[t3[0] : t3[1]-1]
			bracketc := 0
			pname := ""
			for i := t3[1]; i < len(op); i++ {
				str := op[i : i+1]
				if str == "(" {
					bracketc++
				} else if str == ")" {
					if bracketc == 0 {
						pname = op[t3[1]:i]
					} else {
						bracketc--
					}
				}
			}
			name := "JsFunc_" + strconv.Itoa(len(strs))
			strs[name] = ParseJsFunc(fname, pname, strs, data)
			op = strings.Replace(op, fname+"("+pname+")", `{{`+name+`}}`, -1)
		} else {
			break
		}
	}

	//parse ()
	for {
		reg2 := regexp.MustCompile(`\(`)
		t3 := reg2.FindStringIndex(op)
		if len(t3) > 0 {

			bracketc := 0
			pname := ""
			for i := t3[1]; i < len(op); i++ {
				str := op[i : i+1]
				if str == "(" {
					bracketc++
				} else if str == ")" {
					if bracketc == 0 {
						pname = op[t3[1]:i]
					} else {
						bracketc--
					}
				}
			}
			name := "bracket_" + strconv.Itoa(len(strs))
			strs[name] = ParseJsOperator(pname, strs, data)
			op = strings.Replace(op, "("+pname+")", `{{`+name+`}}`, -1)
		} else {
			break
		}
	}

	// if len(t4) > 0 {
	// 	for _, v := range t4 {
	// 		name := "bracket_" + strconv.Itoa(len(strs))

	// 		strv1 := v[1]
	// 		log.Debug(strv1)
	// 		strs[name] = ParseJsOperator(v[1], strs, data)
	// 		op = strings.Replace(op, "("+v[1]+")", `{{`+name+`}}`, -1)
	// 	}
	// }
	//parse string concat
	t5 := strings.Split(op, " ")

	if len(t5) > 1 {
		opstr := ""
		for _, v := range t5 {
			opstr += ParseJsOperator(v, strs, data)
		}
		return opstr
	}

	t5 = strings.Split(op, "+")
	if len(t5) > 1 {
		opnum := 0
		for _, v := range t5 {
			value, _ := strconv.Atoi(ParseJsOperator(v, strs, data))
			opnum += value
		}
		return strconv.Itoa(opnum)
	}

	t5 = strings.Split(op, "-")
	if len(t5) > 1 {
		opnum := 0
		for i, v := range t5 {
			value, _ := strconv.Atoi(ParseJsOperator(v, strs, data))
			if i == 0 {
				opnum = value
			} else {
				opnum -= value
			}
		}
		return strconv.Itoa(opnum)
	}

	t5 = strings.Split(op, "*")
	if len(t5) > 1 {
		opnum := 0
		for i, v := range t5 {
			value, _ := strconv.Atoi(ParseJsOperator(v, strs, data))
			if i == 0 {
				opnum = value
			} else {
				opnum *= value
			}
		}
		return strconv.Itoa(opnum)
	}

	t5 = strings.Split(op, "/")
	if len(t5) > 1 {
		opnum := 0
		for i, v := range t5 {
			value, _ := strconv.Atoi(ParseJsOperator(v, strs, data))
			if i == 0 {
				opnum = value
			} else {
				opnum /= value
			}
		}
		return strconv.Itoa(opnum)
	}

	//check if string
	if len(op) > 4 && op[:2] == `{{` {
		return strs[op[2:len(op)-2]]
	} else if op == "Templateurl" {
		return data.Templateurl
	} else if op == "Siteurl" {
		return data.Siteurl
	} else if op == "Imageurl" {
		return data.Imageurl
	} else {
		//check number:
		if _, err := strconv.ParseInt(op, 10, 64); err == nil {
			return op
		} else {

			//parse data
			args := strings.Split(op, ".")

			var databind map[string]json.RawMessage
			var dataarr []json.RawMessage
			var datastr json.RawMessage

			//get query condition and remove
			regcond := regexp.MustCompile(`\w+\[(.+)\]`)
			tcond := regcond.FindAllStringSubmatch(op, -1)
			if len(tcond) > 0 {
				args[0] = strings.Replace(op, "["+tcond[0][1]+"]", "", -1)
			}

			// if op == "Langs._about_url" {
			// 	test1, _ := json.Marshal(data.Data[args[0]])
			// 	log.Debugf("Page.Langs:%s-%s ", op, test1)
			// }
			//try to get value
			var err = json.Unmarshal(data.Data[args[0]], &databind)

			// if err != nil {
			// 	json.Unmarshal(data.Data[args[0]], &datajson)

			// } else

			databind = data.Data
			//loop throught params like Configs.test.a.b
			for loopi, _ := range args {

				if loopi < len(args)-1 {
					if args[loopi] == "thebest" {
						log.Debugf("the best %s %v", args[loopi], databind)
					}
					//get query condition
					regcond = regexp.MustCompile(`\w+\[(.+)\]`)
					tcond = regcond.FindAllStringSubmatch(op, -1)
					if len(tcond) > 0 {
						args[loopi] = strings.Replace(op, "["+tcond[0][1]+"]", "", -1)
					}
					err = json.Unmarshal(databind[args[loopi]], &databind)

				} else {
					//check array
					err = json.Unmarshal(databind[args[loopi]], &dataarr)
					if err != nil {
						err = json.Unmarshal(databind[args[loopi]], &datastr)
						if args[0] == "thebest" {
							log.Debugf("the best %s %v", args[loopi], databind[args[loopi]])
						}
						if err == nil {
							//try to parse string
							var str string
							err = json.Unmarshal(datastr, &str)
							if err == nil {
								//return string
								//log.Debugf("parseostest str %s : %s", op, string(datastr))
								if args[0] == "Configs" {
									return html.UnescapeString(str)
								}
								return str
							}
							//return number
							//log.Debugf("parseostest num %s : %s", op, string(datastr))
							return string(datastr)

						}
						return "0"
					}
					//if is array -> convert to json string
					b, _ := json.Marshal(dataarr)
					//log.Debugf("parseostest arr %s : %s", op, string(b))
					return string(b)
					//c3mcommon.CheckError("err parse range", err)
				}
			}

			// var strrt string
			// // if op == "i" {
			// // 	log.Debugf("PrepairToParseOP ng-show op: %v, strs: %v, val: %s", string(datajson), datajson)
			// // }

			// err = json.Unmarshal(datajson, &strrt)

			// if err == nil {
			// 	return strrt
			// } else {
			// 	//try parse int
			// 	var intrt int
			// 	err = json.Unmarshal(datajson, &intrt)
			// 	return strconv.Itoa(intrt)
			// }

			// args := strings.Split(op, ".")
			// if len(args) > 1 && data.Data[args[0]] != nil {

			// 	var databind map[string]string
			// 	err := json.Unmarshal(data.Data[args[0]], &databind)
			// 	if c3mcommon.CheckError("error parse data map[string]string - op:"+op+", args[0]:"+args[0], err) {
			// 		if args[0] == "Configs" {
			// 			return html.UnescapeString(databind[args[1]])
			// 		}
			// 		return databind[args[1]]
			// 	} else {
			// 		var databind map[string]json.RawMessage
			// 		err := json.Unmarshal(data.Data[args[0]], &databind)
			// 		if c3mcommon.CheckError("error parse data map[string]json.RawMessage", err) {
			// 			data.Data[args[0]+"_"+args[1]] = databind[args[1]]
			// 			return ParseJsOperator(args[0]+"_"+args[1], strs, data)
			// 		}
			// 	}
			// } else if data.Data[op] != nil {
			// 	var databind string

			// 	err := json.Unmarshal(data.Data[op], &databind)

			// 	if err == nil {
			// 		return databind
			// 	} else {
			// 		//try parse int
			// 		var dataint int
			// 		err = json.Unmarshal(data.Data[op], &dataint)
			// 		return strconv.Itoa(dataint)
			// 	}
			// }
		}
	}
	return op
}

func ParseJsBoolean(op string, strs map[string]string, data models.TemplateViewData) bool {
	//parse function

	for {
		reg2 := regexp.MustCompile(`(\w+)\(`)
		t3 := reg2.FindStringIndex(op)
		if len(t3) > 0 {

			fname := op[t3[0] : t3[1]-1]
			bracketc := 0
			pname := ""
			for i := t3[1]; i < len(op); i++ {
				str := op[i : i+1]
				if str == "(" {
					bracketc++
				} else if str == ")" {
					if bracketc == 0 {
						pname = op[t3[1]:i]
					} else {
						bracketc--
					}
				}
			}
			name := "JsFunc_" + strconv.Itoa(len(strs))
			strs[name] = ParseJsFunc(fname, pname, strs, data)
			op = strings.Replace(op, fname+"("+pname+")", `{{`+name+`}}`, -1)
		} else {
			break
		}
	}

	//parse ()
	for {
		reg2 := regexp.MustCompile(`\(`)
		t3 := reg2.FindStringIndex(op)
		if len(t3) > 0 {

			bracketc := 0
			pname := ""
			for i := t3[1]; i < len(op); i++ {
				str := op[i : i+1]
				if str == "(" {
					bracketc++
				} else if str == ")" {
					if bracketc == 0 {
						pname = op[t3[1]:i]
					} else {
						bracketc--
					}
				}
			}
			name := "bracket_" + strconv.Itoa(len(strs))
			strs[name] = ParseJsOperator(pname, strs, data)
			op = strings.Replace(op, "("+pname+")", `{{`+name+`}}`, -1)
		} else {
			break
		}
	}

	t5 := strings.Split(op, "||")
	if len(t5) > 1 {
		boolrt := false
		for _, v := range t5 {
			boolrt = boolrt || ParseJsBoolean(v, strs, data)
		}
		return boolrt
	}

	t5 = strings.Split(op, "&&")
	if len(t5) > 1 {
		boolrt := true
		for _, v := range t5 {
			boolrt = boolrt && ParseJsBoolean(v, strs, data)
		}
		return boolrt
	}

	t5 = strings.Split(op, ">=")
	if len(t5) > 1 {
		val1, _ := strconv.Atoi(ParseJsOperator(t5[0], strs, data))
		val2, _ := strconv.Atoi(ParseJsOperator(t5[1], strs, data))
		return val1 >= val2
	}

	t5 = strings.Split(op, ">")
	if len(t5) > 1 {
		val1, _ := strconv.Atoi(ParseJsOperator(t5[0], strs, data))
		val2, _ := strconv.Atoi(ParseJsOperator(t5[1], strs, data))
		return val1 > val2
	}

	t5 = strings.Split(op, "<=")
	if len(t5) > 1 {
		val1, _ := strconv.Atoi(ParseJsOperator(t5[0], strs, data))
		val2, _ := strconv.Atoi(ParseJsOperator(t5[1], strs, data))
		return val1 <= val2
	}

	t5 = strings.Split(op, "<")
	if len(t5) > 1 {

		val1, _ := strconv.Atoi(ParseJsOperator(t5[0], strs, data))
		val2, _ := strconv.Atoi(ParseJsOperator(t5[1], strs, data))

		return val1 < val2
	}

	t5 = strings.Split(op, "==")
	if len(t5) > 1 {
		val1 := ParseJsOperator(t5[0], strs, data)
		val2 := ParseJsOperator(t5[1], strs, data)
		//log.Debugf("condqueryop: %v, v1: %s-%v, v2: $s-%s, rs:%v, str:%v", op, t5[0], val1, t5[1], val2, val1 == val2, strs)
		return val1 == val2
	}

	t5 = strings.Split(op, "!=")
	if len(t5) > 1 {
		val1 := ParseJsOperator(t5[0], strs, data)

		val2 := ParseJsOperator(t5[1], strs, data)
		return val1 != val2
	}

	//check if string
	val := ParseJsOperator(op, strs, data)

	if val == "true" {
		return true
	} else if val == "false" {
		return false
	} else {
		return val != "" && val != "0"
	}

}
