package util

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func PrintStep(tag string, step string) {
	fmt.Println(color.New().Add(color.BgBlue).Add(color.FgHiWhite).Sprintf("\n[%s] %s", tag, step))
}
