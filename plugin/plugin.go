// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// Args provides plugin execution arguments.
type Args struct {
	Pipeline

	// Level defines the plugin log level.
	Level string `envconfig:"PLUGIN_LOG_LEVEL"`

	// Settings
	Command   string `envconfig:"PLUGIN_COMMAND"`
	Bucket    string `envconfig:"PLUGIN_BUCKET"`
	Region    string `envconfig:"PLUGIN_REGION"`
	Key       string `envconfig:"PLUGIN_KEY"`
	LocalPath string `envconfig:"PLUGIN_LOCAL_PATH"`

	// Secrets
	SecretID  string `envconfig:"COS_SECRETID"`
	SecretKey string `envconfig:"COS_SECRETKEY"`
}

// Function to ensure variables are set to non-empty values
func (a *Args) Validate() error {
	// set to "upload" if the command is empty
	if a.Command == "" {
		a.Command = "upload"
	}

	if a.Bucket == "" {
		return fmt.Errorf("empty bucket\n")
	}
	if a.Region == "" {
		return fmt.Errorf("empty region\n")
	}
	if a.Key == "" {
		return fmt.Errorf("empty key\n")
	}

	if a.Command != "delete" && a.LocalPath == "" {
		return fmt.Errorf("empty localPath\n")
	}

	if a.SecretID == "" {
		return fmt.Errorf("empty COS_SECRETID\n")
	}
	if a.SecretKey == "" {
		return fmt.Errorf("empty COS_SECRETKEY\n")
	}
	return nil
}

// Uploads all files in a directory to COS
func uploadDirToCOS(ctx context.Context, c *cos.Client, localPath string, baseKey string) {
	keysCh := make(chan string, 100)
	// Use a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Function to upload file to COS
	uploadFn := func() {
		defer wg.Done()

		for filePath := range keysCh {
			// Get the relative file path
			relativePath, err := filepath.Rel(localPath, filePath)
			if err != nil {
				fmt.Printf("Failed to get relative path for file %s: %v\n", filePath, err)
				return
			}

			// Upload the file to COS
			objectKey := filepath.Join(baseKey, relativePath)
			_, err = c.Object.PutFromFile(ctx, objectKey, filePath, nil)
			if err != nil {
				fmt.Printf("Failed to upload file %s to COS %s: %v\n", filePath, objectKey, err)
				return
			}
			fmt.Printf("Successfully uploaded %s to %s\n", filePath, objectKey)
		}
	}

	threadpool := 3
	for i := 0; i < threadpool; i++ {
		wg.Add(1)
		go uploadFn()
	}

	// Walk through the directory
	filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Failed to access a path %q: %v\n", path, err)
			return err
		}

		if !info.IsDir() {
			keysCh <- path
		}

		return nil
	})

	close(keysCh)
	// Wait for all uploads to finish
	wg.Wait()
}

// Download all files in a directory from COS
func downloadFromCos(ctx context.Context, c *cos.Client, localPath string, baseKey string) {
	keysCh := make(chan []string, 100)
	var wg sync.WaitGroup

	// Download files in queue from COS
	downloadFn := func() {
		defer wg.Done()
		for keys := range keysCh {
			key := keys[0]
			filename := keys[1]
			_, err := c.Object.GetToFile(ctx, key, filename, nil)
			if err != nil {
				fmt.Println(err)
			}

			fmt.Printf("Successfully downloaded %s from %s\n", filename, key)
		}
	}

	threadpool := 3
	for i := 0; i < threadpool; i++ {
		wg.Add(1)
		go downloadFn()
	}

	isTruncated := true
	marker := ""

	for isTruncated {
		opt := &cos.BucketGetOptions{
			Prefix:       baseKey,
			Marker:       marker,
			EncodingType: "url", // url 编码
		}
		// 列出目录
		v, _, err := c.Bucket.Get(ctx, opt)
		if err != nil {
			fmt.Println(err)
			break
		}
		for _, c := range v.Contents {
			key, _ := cos.DecodeURIComponent(c.Key) //EncodingType: "url", 先对 key 进行 url decode
			objectKey, err := filepath.Rel(baseKey, key)
			if err != nil {
				fmt.Printf("Failed to get relative path for file %s: %v\n", key, err)
				continue
			}

			localfile := filepath.Join(localPath, objectKey)
			if _, err := os.Stat(path.Dir(localfile)); err != nil && os.IsNotExist(err) {
				os.MkdirAll(path.Dir(localfile), os.ModePerm)
			}
			// 以/结尾的key（目录文件）不需要下载
			if strings.HasSuffix(localfile, "/") {
				continue
			}
			keysCh <- []string{key, localfile}
		}
		marker, _ = cos.DecodeURIComponent(v.NextMarker) // EncodingType: "url"，先对 NextMarker 进行 url decode
		isTruncated = v.IsTruncated
	}

	close(keysCh)
	wg.Wait()
}

// Delete all files in a directory from COS
func deleteFromCos(ctx context.Context, client *cos.Client, baseKey string) {
	isTruncated := true
	marker := ""

	for isTruncated {
		opt := &cos.BucketGetOptions{
			Prefix:       baseKey,
			Marker:       marker,
			EncodingType: "url", // url 编码
		}
		// 列出目录
		v, _, err := client.Bucket.Get(ctx, opt)
		if err != nil {
			fmt.Println(err)
			break
		}
		for _, c := range v.Contents {
			key, _ := cos.DecodeURIComponent(c.Key) //EncodingType: "url", 先对 key 进行 url decode
			_, err := client.Object.Delete(ctx, key)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Printf("Successfully deleted %s\n", key)
		}
		marker, _ = cos.DecodeURIComponent(v.NextMarker) // EncodingType: "url"，先对 NextMarker 进行 url decode
		isTruncated = v.IsTruncated
	}
}

// Exec executes the plugin.
func Exec(ctx context.Context, args Args) error {
	// validate the arguments
	if err := args.Validate(); err != nil {
		return err
	}

	//将<bucket>和<region>修改为真实的信息
	//bucket的命名规则为{name}-{appid} ，此处填写的存储桶名称必须为此格式
	u, _ := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", args.Bucket, args.Region))
	b := &cos.BaseURL{BucketURL: u}
	c := cos.NewClient(b, &http.Client{
		//设置超时时间
		Timeout: 100 * time.Second,
		Transport: &cos.AuthorizationTransport{
			//如实填写账号和密钥，也可以设置为环境变量
			SecretID:  args.SecretID,
			SecretKey: args.SecretKey,
		},
	})

	//上传文件
	if args.Command == "upload" {
		localPath, err := filepath.Rel(".", args.LocalPath)
		if err != nil {
			return err
		}

		// Check if source file is a directory
		fi, err := os.Stat(localPath)
		if err != nil {
			return err
		}

		if fi.IsDir() {
			uploadDirToCOS(ctx, c, localPath, args.Key)

			return nil
		}

		_, err = c.Object.PutFromFile(ctx, args.Key, localPath, nil)
		if err != nil {
			return err
		}

		fmt.Printf("Successfully uploaded %s to %s\n", localPath, args.Key)
		return nil
	}

	//下载文件
	if args.Command == "download" {
		downloadFromCos(ctx, c, args.LocalPath, args.Key)
		return nil
	}

	//删除文件
	if args.Command == "delete" {
		deleteFromCos(ctx, c, args.Key)
		return nil
	}

	fmt.Printf("Unknown command %s\n", args.Command)
	return nil
}
