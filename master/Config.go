package master

import (
	"io/ioutil"
	"encoding/json"
)

type Config struct {
	ApiPort int `json:"apiPort"`
	ApiReadTimeout int `json:"apiReadTimeout"`
	ApiWriteTimeout int `json:"apiWriteTimeout"`
	EtcdEndpoints []string `json:"etcdEndpoints"`
	EtcdDialTimeout int `json:"etcdDialTimeout"`
	Webroot string `json:"webroot"`
	MongodbUri string `mongodbUri`
	MongodbConnectTimeout int `json:"mongodbConnectTimeout"`
}

var (
	G_config *Config
)

func InitConfig(filename string) (err error)  {
	//把配置文件读进来
	var (
		content []byte
		conf Config
	)
	if content,err=ioutil.ReadFile(filename);err!=nil{
		return
	}
	//做json反序列化
	if err = json.Unmarshal(content,&conf);err!=nil{
		return
	}
	//单例赋值
	G_config = &conf
	return
}