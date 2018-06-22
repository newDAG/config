package main

import (
	"fmt"

	"github.com/newdag/config"
)

func main() {

	cfg, err := config.ReadFile("./test.cfg")
	if err != nil {
		panic("can not found config file!")
	}
	//测试GetValue(section, option, defaultValue)
	//注意.(type)的转换类型一定要和defaultValue对应，否则就会报错

	value1 := cfg.GetValue("", "address", "").(string)
	fmt.Println(value1)

	value2 := cfg.GetValue("server", "XXX", "nothing").(string)
	fmt.Println(value2)

	value3 := cfg.GetValue("server", "cacheSize", 0).(int)
	fmt.Println(value3)

	value4 := cfg.GetValue("server", "autoStart", true).(bool)
	fmt.Println(value4)

	value5 := cfg.GetValue("node", "node_addr", "127.0.0.1").(string)
	fmt.Println(value5)
	url, _ := cfg.GetString("node", "url")
	fmt.Println(url)

}
