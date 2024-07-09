package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-gomail/gomail"
	"github.com/lestrrat-go/libxml2"
	"github.com/lestrrat-go/libxml2/xpath"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// google translate
const googleTranslateURL = "http://translate.google.com/m?q=%s&tl=%s&sl=%s"

type Announcer struct {
	Name        string   `json:"name"`
	URL         string   `json:"url"`
	Cache       string   `json:"cache"`
	Subscribers []string `json:"subscriber"`
}

type Data struct {
	Announcer []Announcer `json:"announcer"`
}

var blogContents = make(map[string]string)

const underline = "<hr>"

// generateLink 函数接收一个URL和一个字符串作为参数，返回一个带有链接的完整字符串
func generateLink(url, str string) string {
	var color string

	// 根据传入的字符串确定颜色
	switch str {
	default:
		color = "blue" // 默认颜色为黑色
	}

	// 生成带有链接的完整字符串
	link := fmt.Sprintf("<a href=\"%s\" style=\"color: %s;\">点这儿~</a>", url, color)
	return link
}

// generateLink 函数接收一个URL和一个字符串作为参数，返回一个带有链接的完整字符串
func generateDownloadLink(url, str string) string {
	var color string

	// 根据传入的字符串确定颜色
	switch str {

	default:
		color = "blue" // 默认颜色为蓝色
	}

	// 生成带有链接的完整字符串
	link := fmt.Sprintf("<a href=\"%s\" style=\"color: %s;\">点我下载~</a>", url, color)
	return link
}

func main() {

	data, err := ioutil.ReadFile("data.json")
	if err != nil {
		panic(err)
	}
	// Define structure
	var jsonData Data

	// Parse JSON data
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		panic(err)
	}

	// Access data
	for index, announcer := range jsonData.Announcer {
		name := announcer.Name
		blogListUrl := announcer.URL
		cache := announcer.Cache
		subscribers := announcer.Subscribers

		// 先获取最新的博客链接
		lastTitleUrl := GetLastTitleUrl(blogListUrl)

		// 如果和之前一样，那么说明没有发布新博客，跳过
		if lastTitleUrl == cache {
			fmt.Println(name + "没有发布新博客 pass")
			continue
		}

		// 发布了新博客

		// 先获取页面响应信息
		resp, err := getResp(lastTitleUrl)
		if err != nil {
			fmt.Println("error in getting response")
			return
		}
		title, text, translatedText, list := getBlogInfo(resp, name)
		fmt.Println(title, text, list, translatedText)
		mailText := text + underline + blogContents[name+"_t"] + translatedText + "<p></p></br>"
		//mailText += underline +  "<p>今天的博客链接:" + lastTitleUrl + " </p>"
		//mailText += "<p>今天的博客照片链接:</p>"
		mailText += underline
		mailText += "<p>今天的博客链接是: " + generateLink(lastTitleUrl, name) + " </p>"
		
		if len(list) > 0 {
		mailText += "<p>今天的博客照片链接:</p>"
		for _, img := range list {			
			mailText += "<p>" + img+ "</p>"
		}
	}

		// 群发邮件
		SendMail(subscribers, randomElement(name)+"发博客了~", mailText, name)

		// 删除对应的缓存图片
		deleteAllFiles("./data/" + name)

		jsonData.Announcer[index].Cache = lastTitleUrl

	}

	// 将修改后的 JSON 数据写入原始文件
	updatedData, err := json.MarshalIndent(jsonData, "", "    ")
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile("data.json", updatedData, 0644)
	if err != nil {
		panic(err)
	}

	fmt.Println("JSON 文件已成功更新")
}

// 获取指定路径下所有文件的名称（带后缀）
func getAllFileNames(path string) ([]string, error) {
	// 获取指定路径下的所有文件和子目录
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	// 定义用于存储文件名的切片
	var fileNames []string

	// 遍历所有文件和子目录
	for _, file := range files {
		// 判断是否为文件，是文件则将文件名加入到切片中
		if !file.IsDir() {
			fileNames = append(fileNames, file.Name())
		}
	}

	return fileNames, nil
}

// 删除指定路径下的所有文件
func deleteAllFiles(path string) error {
	// 获取指定路径下的所有文件和子目录
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	// 遍历所有文件和子目录
	for _, file := range files {
		filePath := filepath.Join(path, file.Name())

		// 判断是否为文件，是文件则删除
		if !file.IsDir() {
			err := os.Remove(filePath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// 随机返回列表中的一个元素
func randomElement(k string) string {
	rand.Seed(time.Now().UnixNano()) // 使用当前时间作为随机数种子

	// 根据键 k 选择要返回的列表
	var list []string
	switch k {
	
	default:
		return ""
	}

	// 生成随机索引
	index := rand.Intn(len(list))

	// 返回随机选择的元素
	return list[index]
}

func SendMail(mailTo []string, subject string, body string, name string) error {
	//定义邮箱服务器连接信息，如果是网易邮箱 pass填密码，qq邮箱填授权码

	//mailConn := map[string]string{
	//  "user": "xxx@163.com",
	//  "pass": "your password",
	//  "host": "smtp.163.com",
	//  "port": "465",
	//}



	port, _ := strconv.Atoi(mailConn["port"]) //转换端口类型为int

	m := gomail.NewMessage()

	m.SetHeader("From", m.FormatAddress(mailConn["user"], "sample")) //这种方式可以添加别名
	//说明：如果是用网易邮箱账号发送，以下方法别名可以是中文，如果是qq企业邮箱，以下方法用中文别名，会报错，需要用上面此方法转码
	//m.SetHeader("From", "FB Sample"+"<"+mailConn["user"]+">") //这种方式可以添加别名，即“FB Sample”， 也可以直接用<code>m.SetHeader("From",mailConn["user"])</code> 读者可以自行实验下效果
	//m.SetHeader("From", mailConn["user"])
	m.SetHeader("To", mailTo...)    //发送给多个用户
	m.SetHeader("Subject", subject) //设置邮件主题
	m.SetBody("text/html", body)    //设置邮件正文

	imgs, err2 := getAllFileNames("./data/" + name)
	if err2 != nil {
		fmt.Println("error in getting file")
	}

	for _, img := range imgs {
		m.Attach("./data/" + name + "/" + img)
	}

	d := gomail.NewDialer(mailConn["host"], port, mailConn["user"], mailConn["pass"])

	err := d.DialAndSend(m)
	return err

}

func downloadImage(url, name string, index string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	fmt.Println("./data/" + name + "/" + name + index + ".jpg")
	file, err := os.Create("./data/" + name + "/" + name + index + ".jpg")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	return err
}

func GetLastTitleUrl(url string) string {
	resp, err := getResp(url)
	if err != nil {
		fmt.Println("error in get last title")
		return ""
	}

	// 先xpath解析获得必要内容
	doc, err := libxml2.ParseHTMLReader(resp.Body)
	find := xpath.NodeList(doc.Find("//ul[@class='skin-archiveList']/li//h2/a/@href"))
	if err != nil {
		fmt.Println("error in getting last title url")
		return ""
	}
	originalString := find[0].String()

	// 匹配引号内的内容
	re := regexp.MustCompile(`"(.*?)"`) // 匹配双引号内的任意字符
	matches := re.FindStringSubmatch(originalString)
	if len(matches) >= 2 {
		return "https://ameblo.jp" + matches[1] // 返回完整的最新博客链接地址
	} else {
		return ""
	}
}

func getResp(url string) (resp *http.Response, err error) {

	resp, err = http.Get(url)
	if err != nil {
		fmt.Println("error")
		return nil, err
	}
	return resp, nil
}

// 获取博客相关信息，例如标题，内容，发布时间，图片列表等
func getBlogInfo(resp *http.Response, name string) (title string, blogText string, translatedText string, imgList []string) {

	// 先xpath解析获得必要内容
	doc, err := libxml2.ParseHTMLReader(resp.Body)

	if err != nil {
		fmt.Println("error in parsing")
		return
	}

	// 释放resp资源
	err_ := resp.Body.Close()
	if err_ != nil {
		return
	}

	// 获取博客标题
	title_find := xpath.NodeList(doc.Find("//h1[@class=\"skin-entryTitle\"]/a/text()"))
	title = strings.Split(title_find[0].String(), ":")[0]
	fmt.Println(title)

	// 获取博客正文
	body_find := xpath.NodeList(doc.Find("//div[@id='entryBody']/text()"))

	var textList []string
	for _, node := range body_find {
		text := node.TextContent()
		textList = append(textList, strings.TrimSpace(text))
	}
	// 博客原文
	blogText = blogContents[name+"_title"] + title + "</p> " + underline + blogContents[name]

	for _, text := range textList {
		blogText += "<p>" + text + "</p>"
	}

	blog_text_t := ""
	for _, text := range textList {
		blog_text_t += "<p>" + text + "</p>"
	}

	// 翻译后的博客原文
	translatedText, _ = translate(blog_text_t)

	_ = translatedText

	//fmt.Println(textList)

	// 获取博客图片url
	img_find := xpath.NodeList(doc.Find("//img[@class='PhotoSwipeImage']/@src"))
	for _, node := range img_find {
		src := node.TextContent()
		imgList = append(imgList, strings.TrimSpace(src))
	}
	// 去重
	imgList = removeDuplicates(imgList)

	// 保存图片
	for i, img := range imgList {
		_ = downloadImage(img, name, strconv.Itoa(i+1))
	}

	fmt.Println(imgList)

	// 资源获取完毕, 释放doc资源
	doc.Free()

	return title, blogText, translatedText, imgList
}

// 给列表去重的方法
func removeDuplicates(strList []string) []string {
	// 创建一个 map 用于记录字符串是否已经出现过
	seen := make(map[string]bool)
	var uniqueList []string

	// 遍历原始字符串列表
	for _, str := range strList {
		// 如果字符串在 map 中不存在，则添加到 map 中，并加入到去重后的列表中
		if !seen[str] {
			seen[str] = true
			uniqueList = append(uniqueList, str)
		}
	}

	return uniqueList
}

func translate(text string) (string, error) {
	text = url.QueryEscape(text)
	url := fmt.Sprintf(googleTranslateURL, text, "zh-CN", "ja")
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	data := make([]byte, 0)
	buffer := make([]byte, 1024)
	for {
		n, err := response.Body.Read(buffer)
		if err != nil {
			break
		}
		data = append(data, buffer[:n]...)
	}

	expr := `(?s)class="(?:t0|result-container)">(.*?)<`
	re := regexp.MustCompile(expr)
	match := re.FindStringSubmatch(string(data))
	if len(match) < 2 {
		return "", fmt.Errorf("no translation found")
	}

	return htmlUnescape(match[1]), nil
}

func htmlUnescape(s string) string {
	return strings.NewReplacer(
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", "\"",
		"&apos;", "'",
	).Replace(s)
}
