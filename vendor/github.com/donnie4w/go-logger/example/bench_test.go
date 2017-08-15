package example

import (
	"fmt"
	"strconv"
	"testing"
	//	"time"

	"github.com/donnie4w/go-logger/logger"
)

func _init() {

	//指定是否控制台打印，默认为true
	logger.SetConsole(true)
	//指定日志文件备份方式为文件大小的方式
	//第一个参数为日志文件存放目录
	//第二个参数为日志文件命名
	//第三个参数为备份文件最大数量
	//第四个参数为备份文件大小
	//第五个参数为文件大小的单位
	logger.SetRollingFile(`C:\Users\Thinkpad\Desktop\logtest`, "test-b.log", 20, 1, logger.MB)

	//指定日志文件备份方式为日期的方式
	//第一个参数为日志文件存放目录
	//第二个参数为日志文件命名
	//logger.SetRollingDaily("d:/logtest", "test.log")

	//指定日志级别  ALL，DEBUG，INFO，WARN，ERROR，FATAL，OFF 级别由低到高
	//一般习惯是测试阶段为debug，生成环境为info以上
	logger.SetLevel(logger.DEBUG)

}

func log(i int) {
	logger.Debug("Debug>>>>>>>>>>>>>>>>>>>>>>", strconv.Itoa(i))
	logger.Info("Info>>>>>>>>>>>>>>>>>>>>>>>>>", strconv.Itoa(i))
	logger.Warn("Warn>>>>>>>>>>>>>>>>>>>>>>>>>", strconv.Itoa(i))
	logger.Error("Error>>>>>>>>>>>>>>>>>>>>>>>>>", strconv.Itoa(i))
	logger.Fatal("Fatal>>>>>>>>>>>>>>>>>>>>>>>>>", strconv.Itoa(i))
}

func division(a, b int32) {
	fmt.Println(a, b)
}

func Benchmark_log(b *testing.B) {
	b.N = 10000
	fmt.Println("Benchmark_log()")
	//b.StopTimer()
	//b.StartTimer()
	for i := 0; i < b.N; i++ { //use b.N for looping
		division(4, int32(i))
		go log(i)
		//time.Sleep(10 * time.Millisecond)
	}

}
