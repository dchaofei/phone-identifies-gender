package control

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Xml struct {
	xml string
}

func NewXml(xml string) *Xml {
	return &Xml{xml: xml}
}

func (x *Xml) Position(resourceId string) (*Position, error) {
	boundsRegexp := regexp.MustCompile(fmt.Sprintf(`%s.*?bounds="(.*?)"`, resourceId))
	params := boundsRegexp.FindStringSubmatch(x.xml)

	if len(params) != 2 {
		return nil, fmt.Errorf("没有匹配到 %s", resourceId)
	}

	s := strings.Split(params[1], "]")
	slice1 := strings.Split(s[0], ",")
	slice2 := strings.Split(s[1], ",")
	x1 := s2i(strings.Trim(slice1[0], "[]"))
	y1 := s2i(strings.Trim(slice1[1], "[]"))
	x2 := s2i(strings.Trim(slice2[0], "[]"))
	y2 := s2i(strings.Trim(slice2[1], "[]"))
	return &Position{
		X: (x2-x1)/2 + x1,
		Y: (y2-y1)/2 + y1,
	}, nil
}

func s2i(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
