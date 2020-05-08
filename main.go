package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/tidusant/c3m-common/c3mcommon"
	"github.com/tidusant/c3m-common/inflect"
	"github.com/tidusant/c3m-common/log"
	"github.com/tidusant/c3m-common/lzjs"
	"github.com/tidusant/chadmin-repo/models"

	//"io"

	"net/http"
	//	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type PageData struct {
	Page   models.PageView
	Blocks map[string]map[string]string
}

type PageDesc struct {
	Slug        string
	Title       string
	Description string
}

type CommonData struct {
	Langs map[string]string
	Pages map[string]PageDesc
}

var (
	pagedata models.TemplateViewData
	pages    map[string](map[string]string)
	curpage  string
	curlang  string

	templates     map[string]models.Template
	loaddatadone  bool
	workingfolder string
	corejs        string
	newscatslug   map[string]models.NewsCat
	prodcatslug   map[string]ProdCat
	newsslug      map[string]models.News
	prodslug      map[string]Product
	pageslug      map[string]models.PageView
	carts         map[string]Product
	orgData       []byte
	userkey       string
	mycc          MyCommon
)

func init() {
	newscatslug = make(map[string]models.NewsCat)
	prodcatslug = make(map[string]ProdCat)
	newsslug = make(map[string]models.News)
	prodslug = make(map[string]Product)
	carts = make(map[string]Product)

	pagedata.Resources = make(map[string]string)
	pagedata.Configs = make(map[string]string)
	curlang = "vi"
	workingfolder = viper.GetString("config.workingfolder")
	//check auth
	log.Printf("check key...")

	userkey = viper.GetString("config.key")
	mycc.SetKey(userkey)
	os.Mkdir(workingfolder, os.ModePerm)

}

func main() {

	initdata()
	if !loaddatadone {
		log.Errorf("Load data fail.")
		return
	}
	var port int
	var debug bool

	//check port
	rand.Seed(time.Now().Unix())
	port = 0
	for {
		port = rand.Intn(1024-1) + int(49151) + 1
		if c3mcommon.CheckPort(port) {
			break
		}
	}

	//fmt.Println(mycrypto.Encode("abc,efc", 5))
	flag.BoolVar(&debug, "debug", false, "Indicates if debug messages should be printed in log files")
	flag.Parse()

	logLevel := log.DebugLevel
	if !debug {
		logLevel = log.InfoLevel
		gin.SetMode(gin.ReleaseMode)
	}

	log.SetOutputFile(fmt.Sprintf("portal-"+strconv.Itoa(port)), logLevel)
	defer log.CloseOutputFile()
	log.RedirectStdOut()

	log.Infof("running with port:" + strconv.Itoa(port))

	//init config

	router := gin.Default()

	router.GET("/index", func(c *gin.Context) {
		router.LoadHTMLGlob("html/*")

		if c.Request.Host == "localhost:"+strconv.Itoa(port) {
			c.Header("Access-Control-Allow-Origin", "localhost")
			c.Header("Access-Control-Allow-Headers", "access-control-allow-origin, access-control-allow-headers,access-control-allow-credentials")
			c.Header("Access-Control-Allow-Credentials", "true")
			if loadtemplate() {
				c.HTML(http.StatusOK,
					// Use the index.html template
					"index.html",
					// Pass the data that the page uses (in this case, 'title')
					gin.H{
						"title": "Home Page",
						"menu":  templates,
					})
			} else {
				c.String(http.StatusOK, "load template fail!")
			}

		} else {
			log.Debugf("Not allow " + c.Request.Host)
		}

	})

	router.GET("/submit/:name", func(c *gin.Context) {
		router.LoadHTMLGlob("html/*")
		name := c.Param("name")
		if c.Request.Host == "localhost:"+strconv.Itoa(port) {
			c.Header("Access-Control-Allow-Origin", "localhost")
			c.Header("Access-Control-Allow-Headers", "access-control-allow-origin, access-control-allow-headers,access-control-allow-credentials")
			c.Header("Access-Control-Allow-Credentials", "true")
			rs := SubmitTemplate(name, "")
			c.HTML(http.StatusOK,
				// Use the index.html template
				"redirect.html",
				// Pass the data that the page uses (in this case, 'title')
				gin.H{
					"title":   "Message",
					"message": rs,
				})
		} else {
			log.Debugf("Not allow " + c.Request.Host)
			c.Redirect(302, "/index")
		}

	})

	router.GET("/resubmit/:name/:code", func(c *gin.Context) {
		router.LoadHTMLGlob("html/*")
		name := c.Param("name")
		code := c.Param("code")
		if c.Request.Host == "localhost:"+strconv.Itoa(port) {
			c.Header("Access-Control-Allow-Origin", "localhost")
			c.Header("Access-Control-Allow-Headers", "access-control-allow-origin, access-control-allow-headers,access-control-allow-credentials")
			c.Header("Access-Control-Allow-Credentials", "true")
			rs := SubmitTemplate(name, code)
			c.HTML(http.StatusOK,
				// Use the index.html template
				"redirect.html",
				// Pass the data that the page uses (in this case, 'title')
				gin.H{
					"title":   "Message",
					"message": rs,
				})
		} else {
			log.Debugf("Not allow " + c.Request.Host)
			c.Redirect(302, "/index")
		}

	})

	router.Static("/files/", "./"+workingfolder+"/")
	router.Static("/images/", "./images/")
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "load template fail!")
	})
	router.GET("/template/:name/*path", func(c *gin.Context) {
		name := c.Param("name")

		result := "OK"
		defer func() { //catch or finally
			if err := recover(); err != nil { //catch
				router.LoadHTMLGlob("html/*")

				c.HTML(http.StatusOK,
					// Use the index.html template
					"error.html",
					// Pass the data that the page uses (in this case, 'title')
					gin.H{
						"error": fmt.Sprintf("%s", err),
					})
			}
		}()
		if c.Request.Host == "localhost:"+strconv.Itoa(port) {
			c.Header("Access-Control-Allow-Origin", "localhost")
			c.Header("Access-Control-Allow-Headers", "access-control-allow-origin, access-control-allow-headers,access-control-allow-credentials")
			c.Header("Access-Control-Allow-Credentials", "true")
			//check working dir
			result = renderPage(name, "", strconv.Itoa(port))
			//router.LoadHTMLGlob(workingfolder + "/" + name + "/*.html")

		} else {
			log.Debugf("Not allow " + c.Request.Host)
		}

		//c.String(200, result, nil)
		c.Writer.WriteHeader(http.StatusOK)
		// //Convert your cached html string to byte array
		// c.Writer.Write([]byte(result))
		c.Writer.WriteString(result)

	})
	// router.GET("/template/:name/:slug", func(c *gin.Context) {
	// 	name := c.Param("name")
	// 	slug := c.Param("slug")
	// 	result := "OK"
	// 	defer func() { //catch or finally
	// 		if err := recover(); err != nil { //catch
	// 			router.LoadHTMLGlob("html/*")

	// 			c.HTML(http.StatusOK,
	// 				// Use the index.html template
	// 				"error.html",
	// 				// Pass the data that the page uses (in this case, 'title')
	// 				gin.H{
	// 					"error": fmt.Sprintf("%s", err),
	// 				})
	// 		}
	// 	}()
	// 	if c.Request.Host == "localhost:"+strconv.Itoa(port) {
	// 		c.Header("Access-Control-Allow-Origin", "localhost")
	// 		c.Header("Access-Control-Allow-Headers", "access-control-allow-origin, access-control-allow-headers,access-control-allow-credentials")
	// 		c.Header("Access-Control-Allow-Credentials", "true")

	// 		//check working dir
	// 		result = renderPage(name, slug, strconv.Itoa(port))
	// 		//router.LoadHTMLGlob(workingfolder + "/" + name + "/*.html")

	// 	} else {
	// 		log.Debugf("Not allow " + c.Request.Host)
	// 	}

	// 	//c.String(200, result, nil)
	// 	c.Writer.WriteHeader(http.StatusOK)
	// 	//Convert your cached html string to byte array
	// 	c.Writer.Write([]byte(result))

	// })

	router.GET("/gettemplate/:code", func(c *gin.Context) {
		code := c.Param("code")
		if c.Request.Host == "localhost:"+strconv.Itoa(port) {
			c.Header("Access-Control-Allow-Origin", "localhost")
			c.Header("Access-Control-Allow-Headers", "access-control-allow-origin, access-control-allow-headers,access-control-allow-credentials")
			c.Header("Access-Control-Allow-Credentials", "true")
			message := "Done"
			if !gettemplate(code) {
				message = "Get template fail, please try again."
			}

			c.HTML(http.StatusOK,
				// Use the index.html template
				"redirect.html",
				// Pass the data that the page uses (in this case, 'title')
				gin.H{
					"title":   "Message",
					"message": message,
				})
		} else {
			log.Debugf("Not allow " + c.Request.Host)
			c.Redirect(302, "/index")
		}

	})

	router.POST("/addcart/:slug/:num", func(c *gin.Context) {
		slug := c.Param("slug")
		num, _ := strconv.Atoi(c.Param("num"))
		if num <= 0 {
			num = 1
		} else if num > 9 {
			num = 9
		}
		if c.Request.Host == "localhost:"+strconv.Itoa(port) {
			c.Header("Access-Control-Allow-Origin", "localhost")
			c.Header("Access-Control-Allow-Headers", "access-control-allow-origin, access-control-allow-headers,access-control-allow-credentials")
			c.Header("Access-Control-Allow-Credentials", "true")
			strrt := "success"
			if slug != "" {
				//check carts is contain prod:

				if _, ok := prodslug[slug]; ok {
					prod := prodslug[slug]
					//prod.NumInCart = int32(num)
					carts[slug] = prod
				} else {
					strrt = "false"
				}
			} else {
				strrt = "false"
			}
			c.String(http.StatusOK, strrt)
		} else {
			log.Debugf("Not allow " + c.Request.Host)
			c.Redirect(302, "/index")
		}
	})

	router.POST("/removecart/:slug", func(c *gin.Context) {
		slug := c.Param("slug")

		if c.Request.Host == "localhost:"+strconv.Itoa(port) {
			c.Header("Access-Control-Allow-Origin", "localhost")
			c.Header("Access-Control-Allow-Headers", "access-control-allow-origin, access-control-allow-headers,access-control-allow-credentials")
			c.Header("Access-Control-Allow-Credentials", "true")
			strrt := "success"
			if slug != "" {
				//check carts is contain prod:

				if _, ok := carts[slug]; ok {
					delete(carts, slug)
				} else {
					strrt = "false"
				}
			} else {
				strrt = "false"
			}
			c.String(http.StatusOK, strrt)
		} else {
			log.Debugf("Not allow " + c.Request.Host)
			c.Redirect(302, "/index")
		}
	})

	router.POST("/clearcart", func(c *gin.Context) {

		if c.Request.Host == "localhost:"+strconv.Itoa(port) {
			c.Header("Access-Control-Allow-Origin", "localhost")
			c.Header("Access-Control-Allow-Headers", "access-control-allow-origin, access-control-allow-headers,access-control-allow-credentials")
			c.Header("Access-Control-Allow-Credentials", "true")
			strrt := "success"
			carts = make(map[string]Product)

			c.String(http.StatusOK, strrt)
		} else {
			log.Debugf("Not allow " + c.Request.Host)
			c.Redirect(302, "/index")
		}
	})

	router.POST("/cartcount", func(c *gin.Context) {

		if c.Request.Host == "localhost:"+strconv.Itoa(port) {
			c.Header("Access-Control-Allow-Origin", "localhost")
			c.Header("Access-Control-Allow-Headers", "access-control-allow-origin, access-control-allow-headers,access-control-allow-credentials")
			c.Header("Access-Control-Allow-Credentials", "true")
			strrt := strconv.Itoa(len(carts))

			c.String(http.StatusOK, strrt)
		} else {
			log.Debugf("Not allow " + c.Request.Host)
			c.Redirect(302, "/index")
		}
	})

	go func() {
		for {
			time.Sleep(time.Millisecond * 200)

			log.Debugf("Checking if started...")
			resp, err := http.Get("http://localhost:" + strconv.Itoa(port) + "/test")
			if err != nil {
				log.Debugf("Failed:%s", err)
				continue
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				log.Debugf("Not OK:%s", resp.StatusCode)
				continue
			}

			// Reached this point: server is up and running!
			break
		}
		log.Println("SERVER UP AND RUNNING!")
		open("http://localhost:" + strconv.Itoa(port) + "/index")
	}()

	router.Run(":" + strconv.Itoa(port))
}
func renderPage(name, slug, port string) string {
	dev := viper.GetBool("config.dev")
	siteurl := `http://localhost:` + port + `/template/` + name + `/`
	Templateurl := "/files/" + name + "/"
	TemplatePath := workingfolder + "/" + name + "/"
	initjs := ""

	//load page html
	html := make(map[string]string)
	pages := make(map[string]PageData)
	pagesdesc := make(map[string]PageDesc)
	folders, err := ioutil.ReadDir(workingfolder + "/" + name + "/resources/pages")
	if err == nil {
		for _, d := range folders {
			if !d.IsDir() {
				continue
			}
			files, err := ioutil.ReadDir(workingfolder + "/" + name + "/resources/pages/" + d.Name())
			if err == nil {

				fileresources := make(map[string]map[string]string)
				for _, f := range files {
					if f.Name()[len(f.Name())-4:] != ".txt" {
						continue
					}
					blockname := strings.Replace(f.Name(), ".txt", "", 1)
					b, err := ioutil.ReadFile(workingfolder + "/" + name + "/resources/pages/" + d.Name() + "/" + f.Name())
					if err != nil {

					}
					filecontent := string(b)
					//convert to \n line
					filecontent = strings.Replace(filecontent, "\r\n", "\n", -1)
					filecontent = strings.Replace(filecontent, "\r", "\n", -1)
					lines := strings.Split(filecontent, "\n")
					datavalue := make(map[string]string)
					for _, line := range lines {
						if len(line) == 0 || line[:1] == "#" {
							continue
						}
						cfgArr := strings.Split(line, "::")
						if len(cfgArr) > 2 {
							datavalue[cfgArr[0]] = cfgArr[2]
						}
					}
					fileresources[blockname] = datavalue
				}

				b, _ := ioutil.ReadFile(workingfolder + "/" + name + "/" + d.Name() + ".html")
				html[d.Name()] = string(b)
				var PData models.PageView
				PData.Title = inflect.Camelize(d.Name())
				PData.Code = d.Name()
				PData.Lang = "en"
				PData.Templ = d.Name()
				var langlink models.LangLink
				langlink.Code = "vi"
				langlink.Flag = "vn"
				langlink.Name = "Tiếng Việt"
				langlink.Href = ""
				PData.LangLinks = append(PData.LangLinks, langlink)
				langlink.Code = "en"
				langlink.Flag = "gb"
				langlink.Name = "English"
				langlink.Href = ""
				PData.LangLinks = append(PData.LangLinks, langlink)
				PData.PageType = d.Name()
				var page PageData
				page.Page = PData
				page.Blocks = fileresources
				pages[d.Name()] = page

				var pagedesc PageDesc
				if d.Name() != "home" {
					pagedesc.Slug = d.Name() + "/"
				}
				pagedesc.Title = inflect.Camelize(d.Name())
				pagedesc.Description = ""
				pagesdesc[d.Name()] = pagedesc
			}
		}
	}
	b, _ := json.Marshal(pages)
	initjs += `var blockdatas=` + string(b) + `;`

	//resource
	b, _ = ioutil.ReadFile(workingfolder + "/" + name + "/resources/lang.txt")
	filecontent := string(b)
	//convert to \n line
	filecontent = strings.Replace(filecontent, "\r\n", "\n", -1)
	filecontent = strings.Replace(filecontent, "\r", "\n", -1)
	lines := strings.Split(filecontent, "\n")
	datavalue := make(map[string]string)
	for _, line := range lines {
		if len(line) == 0 || line[:1] == "#" {
			continue
		}
		cfgArr := strings.Split(line, "::")
		if len(cfgArr) > 2 {
			datavalue[cfgArr[0]] = cfgArr[2]
		}
	}
	var commondata CommonData
	commondata.Langs = datavalue
	commondata.Pages = pagesdesc
	// commondata := make(map[string]map[string]map[string]string)
	// data := make(map[string]map[string]string)
	// data["Langs"] = datavalue
	// commondata["en"] = data
	b, _ = json.Marshal(commondata)
	initjs += `var commondata=` + string(b) + `;`

	initjs += corejs
	// b, _ = ioutil.ReadFile(workingfolder + "/" + name + "/js/redirectinit.js")
	// initjs += string(b)
	//javascript model
	jsmodel := ""
	folders, err = ioutil.ReadDir(workingfolder + "/" + name + "/js/models")
	if err == nil {
		for _, f := range folders {
			if f.IsDir() || filepath.Ext(f.Name()) != ".js" {
				continue
			}
			b, _ = ioutil.ReadFile(workingfolder + "/" + name + "/js/models/" + f.Name())
			jsmodel += "\n" + string(b)
		}
	}
	initjs = strings.Replace(initjs, "{{models}}", jsmodel, -1)
	//layout.html
	folders, err = ioutil.ReadDir(workingfolder + "/" + name)
	if err == nil {
		for _, f := range folders {
			if f.IsDir() || filepath.Ext(f.Name()) != ".html" || f.Name() == "index.html" {
				continue
			}
			b, _ = ioutil.ReadFile(workingfolder + "/" + name + "/" + f.Name())
			html[strings.Replace(f.Name(), filepath.Ext(f.Name()), "", 1)] = string(b)
		}
	}
	b, _ = json.Marshal(html)
	initjs = strings.Replace(initjs, "{{html}}", string(b), -1)
	//configs
	b, _ = ioutil.ReadFile(workingfolder + "/" + name + "/resources/config.txt")
	filecontent = string(b)
	//convert to \n line
	filecontent = strings.Replace(filecontent, "\r\n", "\n", -1)
	filecontent = strings.Replace(filecontent, "\r", "\n", -1)
	lines = strings.Split(filecontent, "\n")
	datavalue = make(map[string]string)
	for _, line := range lines {
		if len(line) == 0 || line[:1] == "#" {
			continue
		}
		cfgArr := strings.Split(line, "::")
		if len(cfgArr) > 2 {
			datavalue[cfgArr[0]] = cfgArr[2]
		}
	}
	b, _ = json.Marshal(datavalue)
	initjs = strings.Replace(initjs, "{{configs}}", string(b), -1)

	// b, _ = ioutil.ReadFile(workingfolder + "/" + name + "/js/main.js")
	// initjs += string(b)

	initjs = strings.Replace(initjs, "{{Templateurl}}", Templateurl, -1)
	initjs = strings.Replace(initjs, "{{debug}}", "true", -1)
	initjs = strings.Replace(initjs, "{{templateurl}}", Templateurl, -1)
	initjs = strings.Replace(initjs, "{{siteurl}}", siteurl, -1)
	initjs = strings.Replace(initjs, "{{curlang}}", "en", -1)

	if !dev {
		initjs = c3mcommon.JSMinify(initjs)
	}
	//init.js
	//ioutil.WriteFile(workingfolder+"/"+name+"/js/core/init.js", []byte(initjs), 0644)
	//main html
	b, _ = ioutil.ReadFile(workingfolder + "/" + name + "/index.html")
	str := string(b)
	//b, _ = ioutil.ReadFile(workingfolder + "/" + name + "/layout.html")
	var re = regexp.MustCompile(`<body.*?>(.*?)</body>`)
	//layout := string(b)
	str = re.ReplaceAllString(str, `<body></body>`)

	//jscontroller := `<script src="` + Templateurl + `js/core/init.js"></script>`
	jscontroller := `<script>` + initjs + `</script>`
	str = strings.Replace(str, `{{jscore}}`, jscontroller, 1)

	//css
	reg2 := regexp.MustCompile(`<link.*href="(.*?)".*>`)
	t2 := reg2.FindAllStringSubmatch(str, -1)
	csscontent := ``
	for _, v2 := range t2 {
		cssfile := strings.Replace(v2[1], "{{Templateurl}}", TemplatePath, -1)
		if filepath.Ext(cssfile) == ".css" {
			b, err := ioutil.ReadFile(cssfile)
			if c3mcommon.CheckError(fmt.Sprintf("cannot read file %s!", v2[1]), err) {
				csscontent += c3mcommon.MinifyCSS(b)
			}
		}
	}

	reg2 = regexp.MustCompile(`<link.*rel="stylesheet".*href="(.*?)".*>`)
	str = reg2.ReplaceAllString(str, "")
	reg2 = regexp.MustCompile(`<link.*href="(.*?)".*rel="stylesheet".*>`)
	str = reg2.ReplaceAllString(str, "")
	str = strings.Replace(str, `</head>`, `<style>`+csscontent+`</style></head>`, 1)

	str = strings.Replace(str, "{{Templateurl}}", Templateurl, -1)

	return str

}

// open opens the specified URL in the default browser of the user.
func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}

	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func gettemplate(code string) bool {
	// //load from server:
	// data := url.Values{}
	// data.Add("data", mycrypto.EncodeBK(code, "gettemplate"))
	// rs := mycc.RequestServer("gettemplate", data)
	// if rs.Status != "1" {
	// 	log.Errorf("error auth: %s", rs.Error)
	// 	return false
	// } else {

	// 	var t Template

	// 	json.Unmarshal([]byte(rs.Data), &t)
	// 	//create files

	// 	t.Content = mycrypto.DecodeW(t.Content)
	// 	t.CSS = mycrypto.DecodeW(t.CSS)
	// 	t.Script = mycrypto.DecodeW(t.Script)
	// 	t.Images = mycrypto.DecodeW(t.Images)
	// 	t.Configs = mycrypto.DecodeW(t.Configs)
	// 	t.Langs = mycrypto.DecodeW(t.Langs)
	// 	templateFolder := "./" + workingfolder + "/" + t.Title + "/"

	// 	//remove all content and recreate data
	// 	//os.RemoveAll(templateFolder)

	// 	//recreate file
	// 	os.Mkdir(templateFolder, os.ModePerm)
	// 	fo, err := os.Create(templateFolder + "resources/config.txt")
	// 	if err != nil {
	// 		log.Errorf("error creating config file %s", err)
	// 		return false
	// 	}
	// 	if _, err := fo.Write([]byte(t.Configs)); err != nil {
	// 		log.Errorf("error write configs file %s", err)
	// 		return false
	// 	}
	// 	fo.Close()

	// 	fo, err = os.Create(templateFolder + "resources/lang.txt")
	// 	if err != nil {
	// 		log.Errorf("error creating config file %s", err)
	// 		return false
	// 	}
	// 	if _, err := fo.Write([]byte(t.Langs)); err != nil {
	// 		log.Errorf("error write configs file %s", err)
	// 		return false
	// 	}
	// 	fo.Close()

	// 	//file screenshot
	// 	err = c3mcommon.CreateImageFile(templateFolder+"screenshot.jpg", t.Screenshot)
	// 	if err != nil {
	// 		log.Errorf("error creating screenshot file %s", err)
	// 		return false
	// 	}

	// 	//html files
	// 	var htmltempls map[string]string
	// 	json.Unmarshal([]byte(t.Content), &htmltempls)
	// 	for name, html := range htmltempls {
	// 		fo, err = os.Create(templateFolder + name + ".html")
	// 		if err != nil {
	// 			log.Errorf("error creating config file %s", err)
	// 			return false
	// 		}
	// 		// close fo on exit and check for its returned error
	// 		// write a chunk
	// 		if _, err := fo.Write([]byte(html)); err != nil {
	// 			log.Errorf("error write configs file %s", err)
	// 			return false
	// 		}
	// 		fo.Close()
	// 	}

	// 	//css files
	// 	os.Mkdir(templateFolder+"/css", os.ModePerm)
	// 	var cssfiles map[string]string
	// 	json.Unmarshal([]byte(t.CSS), &cssfiles)
	// 	for name, html := range cssfiles {
	// 		fo, err = os.Create(templateFolder + name)
	// 		if err != nil {
	// 			return false
	// 		}
	// 		if _, err := fo.Write([]byte(html)); err != nil {
	// 			log.Errorf("error write cssfiles  %s", err)
	// 			return false
	// 		}
	// 		fo.Close()
	// 	}

	// 	//script files
	// 	os.Mkdir(templateFolder+"/js", os.ModePerm)
	// 	var scriptfiles map[string]string
	// 	json.Unmarshal([]byte(t.Script), &scriptfiles)
	// 	for name, html := range scriptfiles {
	// 		fo, err = os.Create(templateFolder + name)
	// 		if err != nil {
	// 			log.Errorf("error creating scriptfiles  %s", err)
	// 			return false
	// 		}
	// 		if _, err := fo.Write([]byte(html)); err != nil {
	// 			log.Errorf("error write scriptfiles  %s", err)
	// 			return false
	// 		}
	// 		fo.Close()
	// 	}

	// 	//images files
	// 	os.Mkdir(templateFolder+"/images", os.ModePerm)
	// 	var imagefiles map[string]string
	// 	json.Unmarshal([]byte(t.Images), &imagefiles)
	// 	for name, html := range imagefiles {
	// 		err = c3mcommon.CreateImageFile(templateFolder+"/images/"+name, html)
	// 		if err != nil {
	// 			log.Errorf("error createing images  %s", err)
	// 			return false
	// 		}

	// 	}
	// }
	return true
}

func loadtemplate() bool {
	//load from server:
	templates = make(map[string]models.Template, 0)
	resp := c3mcommon.RequestBuildService("loadtemplate", "POST", "key="+userkey)

	if resp.Status != "1" {
		log.Errorf("error auth: %s", resp.Error)
		return false
	} else {

		var objmapData map[string]models.Template
		json.Unmarshal([]byte(resp.Data), &objmapData)

		//get all template from local folder
		files, _ := ioutil.ReadDir("./" + workingfolder)
		for _, f := range files {
			if f.IsDir() && f.Name() != "scripts" && f.Name() != ".git" {
				if _, ok := templates[f.Name()]; !ok {
					templates[f.Name()] = models.Template{Title: f.Name()}
				}

			}
		}

		for k, t := range objmapData {
			templates[k] = t
		}

	}

	return true

}

func initdata() {

	log.Printf("load data...")
	//make request auth
	pagedata.Data = make(map[string]json.RawMessage)
	pagedata.Configs = make(map[string]string)

	resp := c3mcommon.RequestBuildService("loaddata", "POST", "key="+userkey)

	if resp.Status != "1" {
		log.Errorf("error auth: %s", resp.Error)

	} else {
		//var objmapData map[string]*json.RawMessage

		orgData = []byte(resp.Data)
		loaddata()
	}
}
func loaddata() {
	timestart := time.Now()
	log.Debugf("load data at language: %s", curlang)
	pagedata.Data = make(map[string]json.RawMessage)
	dataparse := make(map[string]json.RawMessage)
	json.Unmarshal(orgData, &dataparse)
	corejsdata := make(map[string]string)
	json.Unmarshal([]byte(dataparse["corejs"]), &corejsdata)
	corejs = corejsdata["data"]
	//load prodcat slug
	var viewprodcats []ProdCat
	var prodcats []models.ProdCat
	json.Unmarshal([]byte(dataparse["ProdCats"]), &prodcats)
	for _, val := range prodcats {
		var prodcat ProdCat
		tb, _ := lzjs.DecompressFromBase64(val.Langs[curlang].Title)
		prodcat.Title = string(tb)
		tb, _ = lzjs.DecompressFromBase64(val.Langs[curlang].Description)
		prodcat.Description = string(tb)
		tb, _ = lzjs.DecompressFromBase64(val.Langs[curlang].Content)
		prodcat.Content = string(tb)
		prodcat.Slug = val.Langs[curlang].Slug
		prodcat.Code = val.Code
		prodcatslug[prodcat.Slug] = prodcat
		viewprodcats = append(viewprodcats, prodcat)
	}
	jsonbyte, _ := json.Marshal(viewprodcats)
	pagedata.Data["ProdCats"] = json.RawMessage(string(jsonbyte))

	//load prod slug
	var prods []Product
	var dataprods []models.Product
	json.Unmarshal([]byte(dataparse["Prods"]), &dataprods)

	for _, val1 := range dataprods {
		prod := parseProductItem(val1)
		//prop price:
		maxprice := 0
		minprice := 0
		for iprop, prop := range prod.Properties {
			//init
			if iprop == 0 {
				maxprice = prop.Price
				minprice = prop.Price
			}
			if maxprice < prop.Price {
				maxprice = prop.Price
			}
			if minprice > prop.Price {
				minprice = prop.Price
			}
		}

		prod.MaxPrice = maxprice
		prod.MinPrice = minprice
		prods = append(prods, prod)
		prodslug[prod.Slug] = prod
		// //product image
		// //images files
		// for _, img := range val1.Langs[curlang].Images {
		// 	if _, err := os.Stat("/images/" + img); os.IsNotExist(err) {
		// 		// path/to/whatever does not exist
		// 		data := url.Values{}
		// 		rs := c3mcommon.RequestUrl(viper.GetString("apiserver.img")+img+"/"+mycrypto.EncodeA(val1.ShopId), "GET", data)
		// 		// err = c3mcommon.CreateImageFile("/images/"+img, html)
		// 		// if err != nil {
		// 		// 	log.Errorf("error createing images  %s", err)
		// 		// 	return false
		// 		// }
		// 		log.Debugf(rs)

		// 	}

		// }

	}
	jsonbyte, _ = json.Marshal(prods)
	pagedata.Data["Prods"] = json.RawMessage(string(jsonbyte))

	// //load news slug
	// var news []models.News
	// json.Unmarshal([]byte(pagedata.Data["News"]), &news)
	// for i, val1 := range news {
	// 	newsslug[val1.Slug] = news[i]
	// }
	// jsonbyte, _ = json.Marshal(news)
	// pagedata.Data["News"] = json.RawMessage(string(jsonbyte))

	// //load newscat slug
	// var newscats []models.NewsCat
	// json.Unmarshal([]byte(pagedata.Data["NewsCats"]), &newscats)
	// for i, val1 := range newscats {
	// 	tb, _ := lzjs.DecompressFromBase64(val1.Title)
	// 	newscats[i].Title = string(tb)
	// 	tb, _ = lzjs.DecompressFromBase64(val1.Description)
	// 	newscats[i].Description = string(tb)
	// 	tb, _ = lzjs.DecompressFromBase64(val1.Content)
	// 	newscats[i].Content = string(tb)
	// 	newscats[i].Items = news
	// 	newscatslug[val1.Slug] = newscats[i]
	// }
	// jsonbyte, _ = json.Marshal(newscats)
	// pagedata.Data["NewsCats"] = json.RawMessage(string(jsonbyte))

	log.Printf("Done")
	log.Debug("loaddata time:%s", time.Since(timestart))
	loaddatadone = true
}
func parseProductItem(orgProd models.Product) Product {
	var prod Product

	tb, _ := lzjs.DecompressFromBase64(orgProd.Langs[curlang].Title)
	prod.Name = string(tb)
	tb, _ = lzjs.DecompressFromBase64(orgProd.Langs[curlang].Description)
	prod.Description = string(tb)
	tb, _ = lzjs.DecompressFromBase64(orgProd.Langs[curlang].Content)
	prod.Content = string(tb)
	prod.Slug = orgProd.Langs[curlang].Slug
	prod.Images = orgProd.Langs[curlang].Images
	prod.Properties = orgProd.Properties
	prod.Avatar = orgProd.Langs[curlang].Avatar
	prod.CatId = orgProd.CatId
	//get cattitle
	for _, cat := range prodcatslug {
		if cat.Code == prod.CatId {
			prod.CatTitle = cat.Title
			prod.CatSlug = cat.Slug
			break
		}
	}

	return prod
}

func loadtemplatedata(name string, slug string) {
	timestart := time.Now()
	var templfolder = workingfolder + "/" + name
	if _, err := os.Stat("./" + templfolder); err != nil {
		if os.IsNotExist(err) {
			log.Errorf("template folder does not exist")
			// file does not exist
			return
		} else {
			// other error
		}
	}

	//load langs
	dat, err := ioutil.ReadFile(templfolder + "/resources/lang.txt")
	log.Debug("open file time:%s", time.Since(timestart))
	if c3mcommon.CheckError("resource load", err) {
		lines := strings.Split(string(dat), "\r\n")
		for _, text := range lines {
			if text == "" {
				continue
			}
			if text[:1] == "#" {
				continue
			}
			rsc := strings.Split(text, "::")
			if len(rsc) > 1 {
				pagedata.Resources[rsc[0]] = rsc[1]
			}

		}
		b, err := json.Marshal(pagedata.Resources)
		c3mcommon.CheckError("cannot parse resources file", err)
		pagedata.Data["Langs"] = json.RawMessage(string(b))
	}

	//load page resource
	//loop to get resource
	files, err := ioutil.ReadDir(templfolder + "/resources/pages/" + slug)
	if err == nil {

		for _, f := range files {
			filersc, err := os.Open(templfolder + "/resources/pages/" + slug + "/" + f.Name())
			if c3mcommon.CheckError("resource load", err) {
				scanner := bufio.NewScanner(filersc)
				rs := make(map[string]string)
				for scanner.Scan() {
					linetext := scanner.Text()
					if strings.Trim(linetext, " ") == "" {
						continue
					}
					if linetext[:1] == "#" {
						continue
					}

					rsc := strings.Split(linetext, "::")
					if len(rsc) > 1 {
						rs[rsc[0]] = rsc[1]
					}
				}
				b, err := json.Marshal(rs)
				c3mcommon.CheckError("cannot parse resources file", err)

				pagedata.Data[strings.Replace(f.Name(), ".txt", "", 1)] = json.RawMessage(string(b))
				log.Debugf("thebestload %s %s", f.Name(), string(b))
			}

		}
	}

	//load config
	// b, err := ioutil.ReadFile(templfolder + "/resources/config.txt")
	// if c3mcommon.CheckError("resources/config.txt load error", err) {
	// win16be := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
	// utf16bom := unicode.BOMOverride(win16be.NewDecoder())
	// unicodeReader := transform.NewReader(bytes.NewReader(b), utf16bom)
	// scanner := bufio.NewScanner(unicodeReader)
	filersc, err := os.Open(templfolder + "/resources/config.txt")
	if c3mcommon.CheckError("config load", err) {
		scanner := bufio.NewScanner(filersc)

		for scanner.Scan() {
			linetext := scanner.Text()

			if linetext[:1] == "#" {
				continue
			}
			rsc := strings.Split(linetext, "::")
			if len(rsc) > 1 {
				val := strings.Replace(linetext, rsc[0]+"::"+rsc[1]+"::", "", 1)
				pagedata.Configs[rsc[0]] = val
			}
		}
		b, err := json.Marshal(pagedata.Configs)
		c3mcommon.CheckError("cannot parse config file", err)
		pagedata.Data["Configs"] = json.RawMessage(string(b))
	}

	//load page
	pagefiles, _ := ioutil.ReadDir(templfolder + "/resources/pages")
	// win16be := unicode.UTF16(unicode.BigEndian, unicode.UseBOM)
	// utf16bom := unicode.BOMOverride(win16be.NewDecoder())
	// unicodeReader := transform.NewReader(bytes.NewReader(b), utf16bom)
	//scanner := bufio.NewScanner(unicodeReader)
	pagesdata := make(map[string]models.PageView)
	pageslug = make(map[string]models.PageView)
	for _, f := range pagefiles {
		if !f.IsDir() {
			if filepath.Ext(f.Name()) == ".txt" {
				filersc, err = os.Open(templfolder + "/resources/pages/" + f.Name())
				if err != nil {
					c3mcommon.CheckError(fmt.Sprintf("cannot read resources/pages file %s!", f.Name()), err)

				}
				scanner := bufio.NewScanner(filersc)
				var page models.PageView
				for scanner.Scan() {
					linetext := scanner.Text()

					if linetext[:1] == "#" {
						continue
					}
					rsc := strings.Split(linetext, "::")
					if len(rsc) > 1 {

						if strings.ToLower(rsc[0]) == "title" {
							page.Title = rsc[1]
						}
						if strings.ToLower(rsc[0]) == "description" {
							page.Description = rsc[1]
						}
						if strings.ToLower(rsc[0]) == "content" {
							page.Content = rsc[1]
						}
						if strings.ToLower(rsc[0]) == "slug" {
							page.Slug = rsc[1]
						}

					}
				}
				page.Code = strings.Replace(f.Name(), ".txt", "", 1)
				page.Code = inflect.Parameterize(page.Code)
				pagesdata[page.Code] = page
				pageslug[page.Slug] = page

			}
		}
	}
	jsonbyte, _ := json.Marshal(pagesdata)
	pagedata.Data["Pages"] = json.RawMessage(string(jsonbyte))

}
