// Copyright 2021 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package main

import (
	"context"
	"github.com/cloudwego/kitex/client"
	"github.com/hzhq1255/daily-learning/kitex-demo/hello/kitex_gen/api"
	"github.com/hzhq1255/daily-learning/kitex-demo/hello/kitex_gen/api/hello"
	"log"
	"time"
)

func main() {
	client, err := hello.NewClient("hello", client.WithHostPorts("0.0.0.0:8888"))
	if err != nil {
		log.Fatal(err)
	}
	for {
		req := &api.Request{Message: "my request"}
		resp, err := client.Echo(context.Background(), req)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(resp)
		time.Sleep(time.Second)

		addReq := &api.AddRequest{First: 5, Second: 10}
		addResp, err := client.Add(context.Background(), addReq)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(addResp)
		time.Sleep(time.Second)
	}
}
