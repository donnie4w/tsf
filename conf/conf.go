/**
 * Copyright 2017 tsf Author. All Rights Reserved.
 * email: donnie4w@gmail.com
 */
package conf

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"

	"github.com/donnie4w/dom4g"
)

var CF = &confbean{}

type confbean struct {
	HttpPort int
	TcpPort  int

	Logdir  string
	LogName string

	Redis_ip   string
	Redis_port int
	Redis_pwd  string
	Redis_db   int

	Leveldb string

	Conf string
	Auth string
}

/**设置日志信息*/
func (cf *confbean) SetLog(logdir, logname string) {
	cf.Logdir, cf.LogName = logdir, logname
}
func (cf *confbean) GetLog(name string) (logdir string, logname string) {
	if cf.LogName != "" {
		logname = cf.LogName
	} else {
		logname = name
	}
	logdir = cf.Logdir
	return
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func (cf *confbean) Init() {
	xmlstr := ""
	if isExist(cf.Conf) {
		xmlconfig, err := os.Open(cf.Conf)
		if err != nil {
			panic(fmt.Sprint("xmlconfig is error:", err.Error()))
			os.Exit(0)
		}
		config, err := ioutil.ReadAll(xmlconfig)
		if err != nil {
			panic(fmt.Sprint("config is error:", err.Error()))
			os.Exit(1)
		}
		xmlstr = string(config)
	} else {
		xmlstr = tsfxml
	}

	dom, err := dom4g.LoadByXml(xmlstr)
	if err == nil {
		nodes := dom.AllNodes()
		if nodes != nil {
			fmt.Println(`======================conf start======================`)
			i := 0
			for _, node := range nodes {
				name := node.Name()
				value := node.Value
				v := reflect.ValueOf(cf).Elem().FieldByName(name)
				if v.CanSet() {
					fmt.Println("set====>", name, value)
					switch v.Type().Name() {
					case "string":
						v.Set(reflect.ValueOf(value))
					case "int":
						i, _ := strconv.Atoi(value)
						v.Set(reflect.ValueOf(i))
					default:
						fmt.Println("other type:", v.Type().Name(), ">>>", name)
					}
				} else {
					fmt.Println("no set====>", name, value)
					i++
				}
			}
			fmt.Println(`=======================conf end=======================`)
			if i > 0 {
				fmt.Println("not set number:", i)
			}
		}
	}
}

func (this *confbean) ParseFlag() {
	flag.IntVar(&this.HttpPort, "http", 8181, "http port")
	flag.IntVar(&this.TcpPort, "tcp", 7373, "tcp port")
	flag.StringVar(&this.Conf, "cf", "tsf.xml", "config file")
	flag.StringVar(&this.Auth, "auth", "tsf", "password")
}

var tsfxml = `<tsf>
	<HttpPort>8181</HttpPort>
	<TcpPort>7373</TcpPort>	
	<Logdir>./</Logdir>
	<LogName>tsf.log</LogName>
	<Auth>tsf</Auth>
</tsf>`
