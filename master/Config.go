package master

import (
	"io/ioutil"
	"encoding/json"
	"fmt"
)

type Config struct {
	ApiPort int `json:"apiPort"`
	ApiReadTimeOut int `json:"apiReadTimeOut"`
	ApiWriteTimeOut int `json:"apiWriteTimeOut"`
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