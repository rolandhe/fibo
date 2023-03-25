package fibo

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var appConf map[string]any

func init() {
	profile := os.Getenv("profile")
	confFile := "./conf/app.yaml"
	if profile != "" {
		confFile = fmt.Sprintf("./conf/app-%s.yaml", profile)
	}

	log.Printf("use profile:%s, conf: %s\n", profile, confFile)

	buff, err := os.ReadFile(confFile)
	if err != nil {
		log.Println(err)
		return
	}

	if err := yaml.Unmarshal(buff, &appConf); err != nil {
		log.Println(err)
		return
	}
}

type AppInfo struct {
	Name string
	Port int
}

type zookeeperConf struct {
	servers        []string
	sessionTimeout time.Duration
	maxWorkerBits  int
	maxIdcBits     int
}

type configure struct {
	maxWorkerBits int
	maxIdcBits    int
	nameSpaces    []string
}

func GetAppInfo() *AppInfo {
	v, ok := appConf["app"]
	if !ok {
		return nil
	}

	mp := v.(map[string]any)
	return &AppInfo{
		Name: GetFromMapString(mp, "name"),
		Port: getFromMapInt(mp, "port"),
	}
}

func getZkConf() *zookeeperConf {
	v, ok := appConf["zk"]
	if !ok {
		return nil
	}

	mp := v.(map[string]any)
	serverString := GetFromMapString(mp, "servers")

	timeout := time.Duration(getFromMapInt(mp, "sessionTimeout"))
	fiboConf := getFiboConfigure()
	return &zookeeperConf{
		servers:        strToSlice(serverString),
		sessionTimeout: time.Second * timeout,
		maxWorkerBits:  fiboConf.maxWorkerBits,
		maxIdcBits:     fiboConf.maxIdcBits,
	}
}

func getFiboConfigure() *configure {
	v, ok := appConf["fibo"]
	if !ok {
		return nil
	}

	mp := v.(map[string]any)
	nsString := GetFromMapString(mp, "nameSpaces")
	nameSpaces := strToSlice(nsString)
	return &configure{
		maxWorkerBits: getFromMapInt(mp, "maxWorkerBits"),
		maxIdcBits:    getFromMapInt(mp, "maxIdcBits"),
		nameSpaces:    nameSpaces,
	}
}

func strToSlice(str string) []string {
	items := strings.Split(str, ",")
	for i := 0; i < len(items); i++ {
		v := items[i]
		items[i] = strings.TrimLeft(strings.Trim(v, " "), " ")
	}
	return items
}

func getMapInfo(conf map[string]any, key string) map[string]any {
	v, ok := conf[key]
	if !ok {
		return nil
	}
	return v.(map[string]any)
}

func convertString(raw any) string {
	v, ok := raw.(int)
	if ok {
		return strconv.Itoa(v)
	}

	return raw.(string)
}
func GetFromMapString(mpValue map[string]any, key string) string {
	v, ok := mpValue[key]
	if !ok {
		return ""
	}
	return v.(string)
}
func IsProdEnv() bool {
	profile := os.Getenv("profile")
	return profile == "prod"
}

func getFromMapInt(mpValue map[string]any, key string) int {
	v, ok := mpValue[key]
	if !ok {
		return 0
	}
	return v.(int)
}
