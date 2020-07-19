package control

import (
	"fmt"
	adb "github.com/zach-klippenstein/goadb"
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

const (
	searchPageXmlName  = "%s_search.xml"
	searchPage2XmlName = "%s_search2.xml"
	resultPageXmlName  = "%s_result.xml"

	searchDir = "./runtime/search_xml/"
	resultDir = "./runtime/result_xml/"
)

// 多个手机控制
var Controls = map[string]*Control{}

func initControls() {
	fatal := func(v ...interface{}) {
		log.Fatal("initControls ", v)
	}

	client, err := adb.New()
	if err != nil {
		fatal("adb.New:", err)
	}
	log.Println("启动中...")

	if err = client.StartServer(); err != nil {
		fatal("client.StartServer:", err)
	}

	serials, err := client.ListDeviceSerials()
	if err != nil {
		fatal("client.ListDeviceSerials", err)
	}

	if len(serials) == 0 {
		fatal("没有设备连接")
	}

	log.Printf("共有%d台设备连接", len(serials))

	for _, serial := range serials {
		deviceDescriptor := adb.DeviceWithSerial(serial)
		Controls[serial], err = NewControl(client.Device(deviceDescriptor), serial)
		if err != nil {
			fatal("NewControl:", err)
		}
	}
}

// 手机控制结构体
type Control struct {
	// 等待搜索结果时间
	WaitResultDuration time.Duration
	// 等待搜索按钮弹出
	WaitSearchButtonDuration time.Duration
	*adb.Device
	// 安卓设备号
	deviceSerial string
	// 输入框坐标
	InputPosition *Position
	// 清除输入框坐标
	ClearInputPosition *Position
	// 搜索按钮坐标
	SearchButtonPosition *Position
}

func NewControl(adb *adb.Device, serial string) (*Control, error) {
	device := &Control{
		Device:       adb,
		deviceSerial: serial,
		WaitResultDuration: 200 * time.Millisecond,
		WaitSearchButtonDuration: 100 * time.Millisecond,
	}
	return device, device.initPosition()
}

func (c *Control) initPosition() error {
	if err := c.initInputPosition(); err != nil {
		return err
	}

	if err := c.Input("init"); err != nil {
		return fmt.Errorf("获取搜索页2 输入init 失败: %s", err)
	}
	xml, err := c.xml(c.localSearchPage2XmlPath(), c.phoneSearchPage2XmlPath())
	if err != nil {
		return fmt.Errorf("获取搜索页2xml失败: %s", err)
	}
	newXml := NewXml(xml)
	if err := c.initClearInputPosition(newXml); err != nil {
		return err
	}
	if err := c.initSearchButtonPosition(newXml); err != nil {
		return err
	}
	if err := c.ClearInput(); err != nil {
		return err
	}
	return nil
}

func (c *Control) initInputPosition() error {
	xml, err := c.xml(c.localSearchPageXmlPath(), c.phoneSearchPageXmlPath())
	if err != nil {
		return fmt.Errorf("获取搜索页xml失败: %s", err)
	}

	position, err := NewXml(xml).Position("com.tencent.mm:id/bhn")
	if err != nil {
		return fmt.Errorf("获取手机号输入框坐标失败: %s\n, 请把微信页面定到搜索页,依次点击通信录->新的朋友->上方的搜索框", err)
	}
	c.InputPosition = position
	log.Println("手机号输入框坐标:", c.InputPosition)
	return nil
}

// 初始化清除输入框内容坐标
func (c *Control) initClearInputPosition(xml *Xml) error {
	position, err := xml.Position("com.tencent.mm:id/asq")
	if err != nil {
		return fmt.Errorf("获取清除手机号输入框按钮坐标失败: %s", err)
	}
	c.ClearInputPosition = position
	log.Println("清除手机号输入框坐标:", c.ClearInputPosition)
	return nil
}

// 初始化搜索按钮坐标
func (c *Control) initSearchButtonPosition(xml *Xml) error {
	position, err := xml.Position("com.tencent.mm:id/ga1")
	if err != nil {
		return fmt.Errorf("获取搜索按钮坐标失败: %s", err)
	}
	c.SearchButtonPosition = position
	log.Println("搜索按钮坐标:", c.SearchButtonPosition)
	return nil
}

func (c *Control) xml(localPath, phonePath string) (string, error) {
	// 手机打印 ui 树
	_, err := c.RunCommand("uiautomator", "dump", phonePath)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("adb", "-s", c.deviceSerial, "pull", phonePath, localPath)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	bytes, err := ioutil.ReadFile(localPath)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// 手机端搜索页 xml 保存路径, 用于获取输入框坐标
func (c *Control) phoneSearchPageXmlPath() string {
	return fmt.Sprintf("/sdcard/"+searchPageXmlName, c.deviceSerial)
}

// 用于手机端搜索页的 xml 上传
func (c *Control) localSearchPageXmlPath() string {
	return fmt.Sprintf(searchDir+searchPageXmlName, c.deviceSerial)
}

// 手机端搜索页输入后 xml 保存路径, 用于获取清除和搜索按钮的坐标
func (c *Control) phoneSearchPage2XmlPath() string {
	return fmt.Sprintf("/sdcard/"+searchPage2XmlName, c.deviceSerial)
}

// 用于手机端搜索页2的 xml 上传
func (c *Control) localSearchPage2XmlPath() string {
	return fmt.Sprintf(searchDir+searchPage2XmlName, c.deviceSerial)
}

// 手机端搜索结果页 xml 保存路径
func (c *Control) phoneResultPageXmlPath(phone string) string {
	return fmt.Sprintf("/sdcard/"+resultPageXmlName, phone)
}

// 手机端搜索结果上传到本地路径
func (c *Control) localResultPageXmlPathLocal(phone string) string {
	return fmt.Sprintf(resultDir+resultPageXmlName, phone)
}

// 点击输入框
func (c *Control) ClickInput() error {
	_, err := c.RunCommand("input", "tap", strconv.Itoa(c.InputPosition.X), strconv.Itoa(c.InputPosition.Y))
	if err != nil {
		return fmt.Errorf("点击搜索输入框失败: %s", err)
	}
	return nil
}

// 清除输入框内容
func (c *Control) ClearInput() error {
	_, err := c.RunCommand("input", "tap", strconv.Itoa(c.ClearInputPosition.X), strconv.Itoa(c.ClearInputPosition.Y))
	if err != nil {
		return fmt.Errorf("清除输入框内容失败: %s", err)
	}
	return nil
}

// 点击搜索
func (c *Control) ClickSearch() error {
	_, err := c.RunCommand("input", "tap", strconv.Itoa(c.SearchButtonPosition.X), strconv.Itoa(c.SearchButtonPosition.Y))
	if err != nil {
		return fmt.Errorf("点击搜索失败: %s", err)
	}
	return nil
}

// 返回到上一页
func (c *Control) Back() error {
	_, err := c.RunCommand("input", "keyevent", "4")
	if err != nil {
		return fmt.Errorf("返回到上一页失败: %s", err)
	}
	return nil
}

func (c *Control) search(phone string) (string, error) {
	if err := c.ClickInput(); err != nil {
		return "", err
	}

	if err := c.Input(phone); err != nil {
		return "", fmt.Errorf("输入手机号失败: %s", err)
	}

	time.Sleep(c.WaitSearchButtonDuration)

	if err := c.ClickSearch(); err != nil {
		return "", err
	}

	// 等待，有可能搜索结果慢
	time.Sleep(c.WaitResultDuration)

	xml, err := c.xml(c.localResultPageXmlPathLocal(phone), c.phoneResultPageXmlPath(phone))
	if err != nil {
		return "", fmt.Errorf("获取搜索结果页 xml 失败: %s", err)
	}
	return xml, nil
}

func (c *Control) Input(keyword string) error {
	_, err := c.RunCommand("input", "text", keyword)
	return err
}

func (c *Control) Reset() error {
	if err := c.Back(); err != nil {
		return err
	}

	time.Sleep(50 * time.Millisecond)

	if err := c.ClearInput(); err != nil {
		return err
	}
	return nil
}

func (c *Control) Gender(phone string) (string, error) {
	xml, err := c.search(phone)
	if err != nil {
		return "", err
	}
	genderRegexp := regexp.MustCompile(fmt.Sprintf(`%s.*?content-desc="(男|女|未知)"`, "com.tencent.mm:id/b2c"))
	params := genderRegexp.FindStringSubmatch(xml)
	if len(params) == 2 {
		return params[1], nil
	}

	if match, _ := regexp.MatchString("该用户不存在", xml); match {
		if err := c.ClickInput(); err != nil { // 这里触发输入法，防止后续返回到通讯录页
			return "", err
		}
		return "该用户不存在", nil
	}

	if match, _ := regexp.MatchString("操作过于频繁", xml); match {
		return "", fmt.Errorf("微信提示操作过于频繁")
	}
	return "未知", nil
}

func (c *Control) Serial() string {
	return c.deviceSerial
}
