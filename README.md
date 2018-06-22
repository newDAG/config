config
======
## Installation

	go get github.com/newdag/config

## Operating instructions

Given a sample configuration file:

#test.cfg:
address = 0x1234567890

[server]

cacheSize=2232

autoStart = false

[node]

node_addr: 192.168.1.111

host: www.example.com

protocol: http://

url: %(protocol)s%(host)s

//---------------------------------------------

To read this configuration file, do:

cfg, err := config.ReadFile("test.cfg")

s, _ := cfg.GetString("node", "url") 

s1 = cfg.GetValue("node", "url", "").(string)  

// result is：  http://www.example.com

//注意GetValue后做类型转换.(type)一定要和defaultValue对应，否则就会报错

value1 := cfg.GetValue("", "address", "").(string)  

fmt.Println(value1)

// result is： 0x1234567890

value2 := cfg.GetValue("server", "XXX", "nothing").(string)

fmt.Println(value2)

// result is： nothing

value3 := cfg.GetValue("server", "cacheSize", 0).(int)

fmt.Println(value3)

// result is： 2232


Note the support for unfolding variables (such as *%(base-url)s*), which are read

from the special (reserved) section name *[DEFAULT]*.

Note that sections, options and values are all case-sensitive.

more examples in ./tests
