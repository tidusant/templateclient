package main

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/tidusant/chadmin-repo/models"
)

func ParseJsFunc(name, params string, strs map[string]string, data models.TemplateViewData) string {
	if name == "test" {
		args := strings.Split(params, ",")
		return JsFunc_test(ParseJsOperator(args[0], strs, data), ParseJsOperator(args[1], strs, data))
	} else if name == "SafeHtml" {
		return JsFunc_SafeHtml(params, strs, data)
	} else if name == "InCart" {
		return JsFunc_InCart(params, strs, data)
	} else if name == "CartCount" {
		return JsFunc_CartCount()
	} else if name == "CurrencyFormat" {
		p := ParseJsOperator(params, strs, data)
		return JsFunc_CurrencyFormat(p, data.Configs["currencyFormat"])
	} else if name == "CartTotal" {
		return JsFunc_CartTotal()
	} else if name == "CheckShow" {
		args := strings.Split(params, ",")
		if len(args) > 1 {
			if ParseJsBoolean(args[0], strs, data) {
				return args[1]
			}
		}
		return ""
	} else if name == "ShipFee" {
		return JsFunc_ShipFee(data)
	} else if name == "NumFormat" {
		args := strings.Split(params, ",")
		num, _ := strconv.Atoi(args[1])
		if num <= 0 {
			num = 3
		}
		return JsFunc_NumFormat(ParseJsOperator(args[0], strs, data), num, args[2])
	} else if name == "len" {
		return JsFunc_len(params, strs, data)
	}
	return "unknowJsFunc"
}

func JsFunc_test(p1, p2 string) string {
	return p1 + "" + p2
}

func JsFunc_len(p string, strs map[string]string, data models.TemplateViewData) string {

	if len(p) > 4 && p[:2] == `{{` {
		name := p[2 : len(p)-2]
		if strs[name] != "" {
			return strconv.Itoa(len(strs[name]))
		}
	}

	val := ParseJsOperator(p, strs, data)
	var dataloop []json.RawMessage
	err := json.Unmarshal([]byte(val), &dataloop)
	if err == nil {
		return strconv.Itoa(len(dataloop))
	}
	var datastr string
	err = json.Unmarshal([]byte(val), &datastr)
	if err == nil {
		return strconv.Itoa(len(datastr))
	}
	return strconv.Itoa(len(val))
}

func JsFunc_SafeHtml(p string, strs map[string]string, data models.TemplateViewData) string {

	// if len(p) > 4 && p[:2] == `{{` {
	// 	name := p[2 : len(p)-2]
	// 	if strs[name] != "" {

	// 		strrt, _ := lzjs.DecompressFromBase64(strs[name])
	// 		return strrt
	// 	}
	// }

	val := ParseJsOperator(p, strs, data)

	// var datastr string
	// _ := json.Unmarshal(data.Data[val], &datastr)
	// if err == nil {

	// 	strrt, _ := lzjs.DecompressFromBase64(datastr)
	// 	return strrt

	// }

	return val

}

func JsFunc_NumFormat(p string, num int, sign string) string {
	if sign == "comma" {
		sign = ","
	} else if sign == "dot" {
		sign = "."
	} else if sign == "dash" {
		sign = "-"
	} else if len(sign) < num {
		sign = sign
	} else {
		return p
	}
	for i := num; i < len(p); i += num {
		p = p[:len(p)-i] + sign + p[len(p)-i:]
		i++
	}
	return p
}
func JsFunc_InCart(p string, strs map[string]string, data models.TemplateViewData) string {
	val := ParseJsOperator(p, strs, data)

	if _, ok := Carts[val]; ok {
		return "true"
	}
	return "false"
}
func JsFunc_CartCount() string {
	var count int32
	// for _, v := range Carts {
	// 	count += v.NumInCart

	// }
	return strconv.Itoa(int(count))

}
func JsFunc_CartTotal() string {
	var total int32
	// for _, v := range Carts {
	// 	total += v.Price * v.NumInCart

	// }
	return strconv.Itoa(int(total))
}
func JsFunc_CurrencyFormat(p string, format string) string {
	cc := strings.Split(format, ",")
	if len(cc) < 2 {
		cc[0] = ""
		cc[1] = "dot"
		cc[2] = "3"
	}
	num, _ := strconv.Atoi(cc[2])
	return JsFunc_NumFormat(p, num, cc[1]) + cc[0]
}
func JsFunc_ShipFee(data models.TemplateViewData) string {
	total, _ := strconv.Atoi(JsFunc_CartTotal())
	freeship, _ := strconv.Atoi(data.Configs["freeship"])

	if total > freeship {
		return "0"
	}
	if JsFunc_CartCount() == "0" {
		return "0"
	}
	return data.Configs["shipfee"]
}
