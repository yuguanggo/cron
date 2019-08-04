package worker

import (
	"io/ioutil"
	"encoding/json"
	"fmt"
)

type Config struct {
	EtcdEndpoints []string `json:"etcdEndpoints"`
	EtcdDialTimeout int `json:"etcdDialTimeout"`
}

var (
	G_config *Config
)

func InitConfig(filename string)(err error)  {
	var (
		content []byte
		config Config
	)
	if content,err = ioutil.ReadFile(filename);err!=nil{
		return
	}
	if err=json.Unmarshal(content,&config);err!=nil{
		return
	}
	G_config=&config
	fmt.Println(G_config)
	return
}