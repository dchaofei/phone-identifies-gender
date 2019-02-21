package device

import (
	"fmt"
	"github.com/zach-klippenstein/goadb"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func init() {
	initDir()
	initAllDevices()
}

var Devices = map[string]*Device{}

const (
	searchDir = "./runtime/search_xml"
	resultDir = "./runtime/result_xml"
)

func initAllDevices() {
	client, err := adb.New()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Starting server...")
	client.StartServer()

	serials, err := client.ListDeviceSerials()
	if err != nil {
		log.Fatal(err)
	}

	if len(serials) == 0 {
		log.Fatal("没有设备连接")
	}
	log.Printf("共有%d台设备连接", len(serials))

	for _, serial := range serials {
		deviceDescriptor := adb.DeviceWithSerial(serial)
		Devices[serial] = New(client.Device(deviceDescriptor))
	}
}

func initDir() {
	createDir(searchDir)
	createDir(resultDir)
}

func createDir(path string) {
	if !isExist(path) {
		os.MkdirAll(path, os.ModePerm)
	}
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

type Position struct {
	x, y int
}

type Device struct {
	*adb.Device
	inputPhonePosition *Position
	clearInputPosition *Position
}

func New(adbDevice *adb.Device) *Device {
	device := &Device{adbDevice, nil, nil}
	device.initUiControlPosition()
	return device
}

func (d Device) searchPage1XmlPath() string {
	serial, _ := d.Serial()
	return fmt.Sprintf("/sdcard/%s_search.xml", serial)
}

func (d Device) searchPage1XmlPathLocal() string {
	serial, _ := d.Serial()
	return fmt.Sprintf("%s/%s_search.xml", searchDir, serial)
}

func (d Device) resultPageXmlPath(keyword string) string {
	return fmt.Sprintf("/sdcard/%s_result.xml", keyword)
}

func (d Device) resultPageXmlPathLocal(keyword string) string {
	return fmt.Sprintf("%s/%s_result.xml", resultDir, keyword)
}

func (d *Device) initUiControlPosition() {
	if d.inputPhonePosition != nil {
		return
	}
	_, err := d.RunCommand("uiautomator", "dump", d.searchPage1XmlPath())
	if err != nil {
		log.Fatal("获取搜索页 UI失败:", err)
	}
	serial, _ := d.Serial()
	pull(d.searchPage1XmlPath(), serial, d.searchPage1XmlPathLocal())
	searchByte, err := ioutil.ReadFile(d.searchPage1XmlPathLocal())
	if err != nil {
		log.Fatal("读取搜索页 xml 失败:", err)
	}
	var buf []byte
	buf = searchByte

	d.setInputPosition(string(buf))
	d.setClearInputPosition()
}

func (d *Device) setInputPosition(xml string) {
	boundsRegexp := regexp.MustCompile(`微信号/QQ号/手机号.*?bounds="(.*?)"`)
	params := boundsRegexp.FindStringSubmatch(xml)

	if len(params) != 2 {
		log.Fatal("获取手机号输入焦点失败: 请把微信页面定到搜索页,依次点击通信录->新的朋友->上方搜索框")
	}

	paramsSlice := strings.Split(params[1], "]")

	//paramsSlice1 := strings.Split(paramsSlice[0], ",")
	//x1 := strings.Trim(paramsSlice1[0], "[]")
	//y1 := strings.Trim(paramsSlice1[1], "[]")
	paramsSlice2 := strings.Split(paramsSlice[1], ",")
	x2 := strings.Trim(paramsSlice2[0], "[]")
	y2 := strings.Trim(paramsSlice2[1], "[]")
	string2int := func(s string) int {
		i, _ := strconv.Atoi(s)
		return i
	}
	d.inputPhonePosition = &Position{x: string2int(x2), y: string2int(y2)}
	log.Println("获取到手机号输入焦点:", d.inputPhonePosition)
}

func (d *Device) setClearInputPosition() {
	d.clearInputPosition = &Position{d.inputPhonePosition.x - 20, d.inputPhonePosition.y - 20}
}

func (d Device) clickAddPosition() []string {
	return []string{strconv.Itoa(d.inputPhonePosition.x), strconv.Itoa(d.inputPhonePosition.y + 200)}
}

func (d Device) search(keyword string) {
	d.clickInput()

	_, err := d.RunCommand("input", "text", keyword)
	if err != nil {
		log.Fatal("输入手机号失败:", err)
	}
	time.Sleep(500 * time.Millisecond)
	_, err = d.RunCommand("input", append([]string{"tap"}, d.clickAddPosition()...)...)

	if err != nil {
		log.Fatal("点击搜索失败:", err)
	}
	_, err = d.RunCommand("uiautomator", "dump", d.resultPageXmlPath(keyword))
	if err != nil {
		log.Fatal("获取结果页元素失败:", err)
	}
}

func (d Device) clickInput() {
	_, err := d.RunCommand("input", "tap", strconv.Itoa(d.inputPhonePosition.x), strconv.Itoa(d.inputPhonePosition.y))
	if err != nil {
		log.Fatal("点击搜索框失败", err)
	}
}

func (d Device) clearInput() {
	_, err := d.RunCommand("input", "tap", strconv.Itoa(d.clearInputPosition.x), strconv.Itoa(d.clearInputPosition.y))
	if err != nil {
		log.Fatal("清除输入框内容失败", err)
	}
}

func (d Device) Reset(keyword string) {
	d.RunCommand("input", "keyevent", "4")
	d.clearInput()
}

func (d Device) Gender(keyword string) string {
	d.search(keyword)
	serial, _ := d.Serial()
	pull(d.resultPageXmlPath(keyword), serial, d.resultPageXmlPathLocal(keyword))
	reader, err := ioutil.ReadFile(d.resultPageXmlPathLocal(keyword))
	if err != nil {
		log.Fatalf("读取结果页%s失败:%s", d.resultPageXmlPath(keyword), err)
	}
	var buf []byte
	buf = reader

	genderRegexp := regexp.MustCompile(`content-desc="(男|女)"`)

	params := genderRegexp.FindStringSubmatch(string(buf))

	if len(params) < 2 {
		if match, _ := regexp.Match("该用户不存在", buf); match {
			d.clearInput()
		}
		// 当前页是结果不存在， 触发焦点， 在reset时以防返回到通讯录页
		return "未知"
	}

	switch params[1] {
	case "男":
		return "男"
	case "女":
		return "女"
	default:
		return "未知"
	}
}

func pull(path, deviceId, localPath string) {
	cmd := exec.Command("adb", "-s", deviceId, "pull", path, localPath)
	cmd.Run()
}
