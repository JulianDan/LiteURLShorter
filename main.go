package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var ErrLinkNotFound = errors.New("can't find target short link")
var ErrLinkExists = errors.New("this short link already exists")

func JsonToMap(jsonStr string) (map[string]string, error) {
	m := make(map[string]string)
	err := json.Unmarshal([]byte(jsonStr), &m)
	if err != nil {
		fmt.Printf("Unmarshal with error: %+v\n", err)
		return nil, err
	}
	return m, nil
}

func MapToJson(m map[string]string) (string, error) {
	jsonByte, err := json.Marshal(m)
	if err != nil {
		fmt.Printf("Marshal with error: %+v\n", err)
		return "", nil
	}
	return string(jsonByte), nil
}

func get_long_url(short_url string) (string, error) {
	var long_url string
	//查询数据库是否存在
	_, err := os.Stat("data.json")
	if err == nil {
		//存在则读取文件
		file_data, err := os.ReadFile("data.json")
		if err != nil {
			return "", err
		}
		//解析
		data_map, err := JsonToMap(string(file_data))
		if err != nil {
			return "", err
		}
		long_url = data_map[short_url]
		//如果长度为0说明没有找到对应的值，直接返回not found
		if len(long_url) == 0 {
			return "", ErrLinkNotFound
		}
	} else if os.IsNotExist(err) {
		return "", err
	} else {
		return "", err
	}
	return long_url, nil
}

func write_data(short_url, long_url string) error {
	//查询数据库是否存在
	_, err := os.Stat("data.json")
	if err == nil {
		//存在则读取文件
		file_data, err := os.ReadFile("data.json")
		if err != nil {
			return err
		}
		//解析
		data_map, err := JsonToMap(string(file_data))
		if err != nil {
			return err
		}
		//短链接是否已经存在
		if len(data_map[short_url]) != 0 {
			return ErrLinkExists
		}
		data_map[short_url] = long_url
		//尝试将最终的map转为json字符串
		json_string, err := MapToJson(data_map)
		if err != nil {
			return err
		}
		//尝试写入文件
		err = os.WriteFile("data.json", []byte(json_string), 0755)
		if err != nil {
			return err
		}
	} else if os.IsNotExist(err) {
		return err
	} else {
		return err
	}
	return nil
}

func patch_data(short_url, long_url string) error {
	//查询数据库是否存在
	_, err := os.Stat("data.json")
	if err == nil {
		//存在则读取文件
		file_data, err := os.ReadFile("data.json")
		if err != nil {
			return err
		}
		//解析
		data_map, err := JsonToMap(string(file_data))
		if err != nil {
			return err
		}
		//短链接是否存在
		if len(data_map[short_url]) == 0 {
			return ErrLinkNotFound
		}
		data_map[short_url] = long_url
		//尝试将最终的map转为json字符串
		json_string, err := MapToJson(data_map)
		if err != nil {
			return err
		}
		//尝试写入文件
		err = os.WriteFile("data.json", []byte(json_string), 0755)
		if err != nil {
			return err
		}
	} else if os.IsNotExist(err) {
		return err
	} else {
		return err
	}
	return nil
}

func del_data(short_url string) error {
	//查询数据库是否存在
	_, err := os.Stat("data.json")
	if err == nil {
		//存在则读取文件
		file_data, err := os.ReadFile("data.json")
		if err != nil {
			return err
		}
		//解析
		data_map, err := JsonToMap(string(file_data))
		if err != nil {
			return err
		}
		//短链接是否存在
		if len(data_map[short_url]) == 0 {
			return ErrLinkNotFound
		}
		delete(data_map, short_url)

		//尝试将最终的map转为json字符串
		json_string, err := MapToJson(data_map)
		if err != nil {
			return err
		}
		//尝试写入文件
		err = os.WriteFile("data.json", []byte(json_string), 0755)
		if err != nil {
			return err
		}
	} else if os.IsNotExist(err) {
		return err
	} else {
		return err
	}
	return nil
}

func main() {
	user_info := make(map[string]string)
	_, err := os.Stat("user.json")
	if err == nil {
		//存在则读取文件
		file_data, err := os.ReadFile("user.json")
		if err != nil {
			panic("Can't load user data.")
		}
		err = json.Unmarshal(file_data, &user_info)
		if err != nil {
			panic("Can't parse user data.")
		}
	}

	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f)

	app := gin.Default()
	app.SetTrustedProxies(nil)

	app.GET("/:id", func(context *gin.Context) {
		//获取短链接对应的长连接
		short_url := context.Param("id")
		long_url, err := get_long_url(short_url)
		//错误处理，分为not found和其他类型
		if err != nil {
			if err == ErrLinkNotFound {
				context.JSON(http.StatusNotFound, gin.H{"info": err.Error()})
				return
			} else {
				context.JSON(http.StatusInternalServerError, gin.H{"info": err.Error()})
				return
			}
		}
		//若没有错误就在这里重定向了, 不要用301永久重定向, 以免浏览器缓存导致无法获取到最新数据
		context.Redirect(http.StatusTemporaryRedirect, long_url)
	})
	app.POST("/:id", func(context *gin.Context) {
		short_url := context.Param("id")
		long_url := context.PostForm("long_url")
		user := context.PostForm("user")
		pwd := context.PostForm("pwd")

		bytes := (sha256.Sum256([]byte(pwd)))
		hashcode := hex.EncodeToString(bytes[:])
		if user_info[user] != hashcode {
			context.JSON(http.StatusForbidden, gin.H{"info": "Password incorrect."})
			return
		}
		err := write_data(short_url, long_url)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"info": err.Error()})
			return
		}
		context.JSON(http.StatusCreated, gin.H{"info": "OK"})
	})

	app.PATCH("/:id", func(context *gin.Context) {
		short_url := context.Param("id")
		long_url := context.PostForm("long_url")
		user := context.PostForm("user")
		pwd := context.PostForm("pwd")

		bytes := (sha256.Sum256([]byte(pwd)))
		hashcode := hex.EncodeToString(bytes[:])
		if user_info[user] != hashcode {
			context.JSON(http.StatusForbidden, gin.H{"info": "Password incorrect."})
			return
		}
		err := patch_data(short_url, long_url)
		if err != nil {
			if err == ErrLinkNotFound {
				context.JSON(http.StatusNotFound, gin.H{"info": err.Error()})
				return
			} else {
				context.JSON(http.StatusInternalServerError, gin.H{"info": err.Error()})
				return
			}
		}
		context.JSON(http.StatusOK, gin.H{"info": "OK"})
	})

	app.DELETE("/:id", func(context *gin.Context) {
		short_url := context.Param("id")
		user := context.Query("user")
		pwd := context.Query("pwd")

		bytes := (sha256.Sum256([]byte(pwd)))
		hashcode := hex.EncodeToString(bytes[:])
		if user_info[user] != hashcode {
			context.JSON(http.StatusForbidden, gin.H{"info": "Password incorrect."})
			return
		}
		err := del_data(short_url)
		if err != nil {
			if err == ErrLinkNotFound {
				context.JSON(http.StatusNotFound, gin.H{"info": err.Error()})
				return
			} else {
				context.JSON(http.StatusInternalServerError, gin.H{"info": err.Error()})
				return
			}
		}
		context.JSON(http.StatusNoContent, gin.H{"info": "OK"})
	})

	app.Run(":80")
}
