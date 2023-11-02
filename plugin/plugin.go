// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// Args provides plugin execution arguments.
type Args struct {
	Pipeline

	// Level defines the plugin log level.
	Level string `envconfig:"PLUGIN_LOG_LEVEL"`

	// Settings
	Command string `envconfig:"PLUGIN_COMMAND"`
	Bucket string `envconfig:"PLUGIN_BUCKET"`
	Region string `envconfig:"PLUGIN_REGION"`
	Source string `envconfig:"PLUGIN_SOURCE"`
	Target string `envconfig:"PLUGIN_TARGET"`
}

// Exec executes the plugin.
func Exec(ctx context.Context, args Args) error {
	// write code here

	//将<bucket>和<region>修改为真实的信息
	//bucket的命名规则为{name}-{appid} ，此处填写的存储桶名称必须为此格式
	/*
	u, _ := url.Parse("https://<bucket>.cos.<region>.myqcloud.com")
	b := &cos.BaseURL{BucketURL: u}
	c := cos.NewClient(b, &http.Client{
		//设置超时时间
		Timeout: 100 * time.Second,
		Transport: &cos.AuthorizationTransport{
			//如实填写账号和密钥，也可以设置为环境变量
			SecretID:  os.Getenv("COS_SECRETID"),
			SecretKey: os.Getenv("COS_SECRETKEY"),
		},
	})
	*/

	return nil
}
