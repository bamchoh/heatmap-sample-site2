package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
)

type clickPoint struct {
	URL string  `json:"url"`
	X   float64 `json:"x"`
	Y   float64 `json:"y"`
}

var cpMap map[string][]clickPoint
var pngs map[string]string

func toBase64(basePNG string) (string, error) {
	f, err := os.Open(basePNG)
	if err != nil {
		fmt.Println("file open error.")
		return "", err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

func screenshot(uri, width string) (string, error) {
	query := url.Values{
		"url":   []string{uri},
		"width": []string{width},
	}
	uri = "https://whispering-river-48114.herokuapp.com/?" + query.Encode()
	fmt.Println(uri)
	resp, err := http.Get(uri)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer resp.Body.Close()

	var jsonData struct {
		PNG     string `json:"png"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&jsonData)
	if err != nil {
		return "", err
	}

	if jsonData.Status == "error" {
		return "", fmt.Errorf("%v", jsonData.Message)
	}

	return jsonData.PNG, nil
}

func keys(m map[string][]clickPoint) []string {
	var ary []string
	for k := range m {
		ary = append(ary, k)
	}
	sort.Strings(ary)
	return ary
}

func appIndex(c *gin.Context) {
	url, _ := c.GetQuery("url")
	urls := keys(cpMap)
	c.HTML(http.StatusOK, "index.tmpl.html", gin.H{"url": url, "urls": urls})
}

func provideClickData(c *gin.Context) {
	url, ok := c.GetQuery("url")
	if !ok || url == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("unknown url:%s, %v", url, ok),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"points": cpMap[url]})
}

func collectClickData(c *gin.Context) {
	dec := json.NewDecoder(c.Request.Body)
	var cp clickPoint
	err := dec.Decode(&cp)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	cps, ok := cpMap[cp.URL]
	if !ok {
		cps = make([]clickPoint, 0)
	}
	cpMap[cp.URL] = append(cps, cp)

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func providePNGData(c *gin.Context) {
	url, ok := c.GetQuery("url")
	if !ok || url == "" {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("unknown url:%s, %v", url, ok),
		})
		return
	}

	width, ok := c.GetQuery("width")
	if !ok {
		width = "1024"
	}

	data, ok := pngs[url]
	ok = false
	if !ok {
		var err error
		data, err = screenshot(url, width)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"status":  "error",
				"message": fmt.Sprintf("screenshot could not take:%v", err),
			})
			return
		}
		pngs[url] = data
	}
	c.JSON(http.StatusOK, gin.H{"png": data})
}

func init() {
	cpMap = make(map[string][]clickPoint)
	pngs = make(map[string]string)
}

func main() {
	fmt.Println(os.Getwd())
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	engine := gin.New()
	store := memstore.NewStore([]byte("secret"))
	engine.Use(sessions.Sessions("mysession", store))
	engine.Use(gin.Logger())
	engine.LoadHTMLGlob("templates/*.tmpl.html")
	engine.Static("/assets", "./assets")
	engine.Static("/static", "static")

	engine.GET("/", appIndex)
	engine.GET("/api/click", provideClickData)
	engine.POST("/api/click", collectClickData)
	engine.GET("/api/png", providePNGData)

	engine.Run(":" + port)
}
