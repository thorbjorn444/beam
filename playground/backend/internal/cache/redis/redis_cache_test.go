// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package redis

import (
	pb "beam.apache.org/playground/backend/internal/api/v1"
	"beam.apache.org/playground/backend/internal/cache"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/google/uuid"
	"reflect"
	"testing"
	"time"
)

func TestRedisCache_GetValue(t *testing.T) {
	pipelineId := uuid.New()
	subKey := cache.RunOutput
	value := "MOCK_OUTPUT"
	client, mock := redismock.NewClientMock()
	marshSubKey, _ := json.Marshal(subKey)
	marshValue, _ := json.Marshal(value)

	type fields struct {
		redisClient *redis.Client
	}
	type args struct {
		ctx        context.Context
		pipelineId uuid.UUID
		subKey     cache.SubKey
	}
	tests := []struct {
		name    string
		mocks   func()
		fields  fields
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "error during HGet operation",
			mocks: func() {
				mock.ExpectHGet(pipelineId.String(), string(marshSubKey)).SetErr(fmt.Errorf("MOCK_ERROR"))
			},
			fields: fields{client},
			args: args{
				ctx:        context.TODO(),
				pipelineId: pipelineId,
				subKey:     subKey,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "all success",
			mocks: func() {
				mock.ExpectHGet(pipelineId.String(), string(marshSubKey)).SetVal(string(marshValue))
			},
			fields: fields{client},
			args: args{
				ctx:        context.TODO(),
				pipelineId: pipelineId,
				subKey:     subKey,
			},
			want:    value,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mocks()
			rc := &Cache{
				tt.fields.redisClient,
			}
			got, err := rc.GetValue(tt.args.ctx, tt.args.pipelineId, tt.args.subKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetValue() got = %v, want %v", got, tt.want)
			}
			mock.ClearExpect()
		})
	}
}

func TestRedisCache_SetExpTime(t *testing.T) {
	pipelineId := uuid.New()
	expTime := time.Second
	client, mock := redismock.NewClientMock()

	type fields struct {
		redisClient *redis.Client
	}
	type args struct {
		ctx        context.Context
		pipelineId uuid.UUID
		expTime    time.Duration
	}
	tests := []struct {
		name    string
		mocks   func()
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "error during Exists operation",
			mocks: func() {
				mock.ExpectExists(pipelineId.String()).SetErr(fmt.Errorf("MOCK_ERROR"))
			},
			fields: fields{client},
			args: args{
				ctx:        context.Background(),
				pipelineId: pipelineId,
				expTime:    expTime,
			},
			wantErr: true,
		},
		{
			name: "Exists operation returns 0",
			mocks: func() {
				mock.ExpectExists(pipelineId.String()).SetVal(0)
			},
			fields: fields{client},
			args: args{
				ctx:        context.Background(),
				pipelineId: pipelineId,
				expTime:    expTime,
			},
			wantErr: true,
		},
		{
			name: "error during Expire operation",
			mocks: func() {
				mock.ExpectExists(pipelineId.String()).SetVal(1)
				mock.ExpectExpire(pipelineId.String(), expTime).SetErr(fmt.Errorf("MOCK_ERROR"))
			},
			fields: fields{client},
			args: args{
				ctx:        context.Background(),
				pipelineId: pipelineId,
				expTime:    expTime,
			},
			wantErr: true,
		},
		{
			name: "all success",
			mocks: func() {
				mock.ExpectExists(pipelineId.String()).SetVal(1)
				mock.ExpectExpire(pipelineId.String(), expTime).SetVal(true)
			},
			fields: fields{client},
			args: args{
				ctx:        context.Background(),
				pipelineId: pipelineId,
				expTime:    expTime,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mocks()
			rc := &Cache{
				tt.fields.redisClient,
			}
			if err := rc.SetExpTime(tt.args.ctx, tt.args.pipelineId, tt.args.expTime); (err != nil) != tt.wantErr {
				t.Errorf("SetExpTime() error = %v, wantErr %v", err, tt.wantErr)
			}
			mock.ClearExpect()
		})
	}
}

func TestRedisCache_SetValue(t *testing.T) {
	pipelineId := uuid.New()
	subKey := cache.Status
	value := pb.Status_STATUS_FINISHED
	client, mock := redismock.NewClientMock()
	marshSubKey, _ := json.Marshal(subKey)
	marshValue, _ := json.Marshal(value)

	type fields struct {
		redisClient *redis.Client
	}
	type args struct {
		ctx        context.Context
		pipelineId uuid.UUID
		subKey     cache.SubKey
		value      interface{}
	}
	tests := []struct {
		name    string
		mocks   func()
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "error during HSet operation",
			mocks: func() {
				mock.ExpectHSet(pipelineId.String(), marshSubKey, marshValue).SetErr(fmt.Errorf("MOCK_ERROR"))
			},
			fields: fields{client},
			args: args{
				ctx:        context.Background(),
				pipelineId: pipelineId,
				subKey:     subKey,
				value:      value,
			},
			wantErr: true,
		},
		{
			name: "all success",
			mocks: func() {
				mock.ExpectHSet(pipelineId.String(), marshSubKey, marshValue).SetVal(1)
				mock.ExpectExpire(pipelineId.String(), time.Minute*15).SetVal(true)
			},
			fields: fields{client},
			args: args{
				ctx:        context.Background(),
				pipelineId: pipelineId,
				subKey:     subKey,
				value:      value,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mocks()
			rc := &Cache{
				tt.fields.redisClient,
			}
			if err := rc.SetValue(tt.args.ctx, tt.args.pipelineId, tt.args.subKey, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("SetValue() error = %v, wantErr %v", err, tt.wantErr)
			}
			mock.ClearExpect()
		})
	}
}

func Test_newRedisCache(t *testing.T) {
	address := "host:port"
	type args struct {
		ctx  context.Context
		addr string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "error during Ping operation",
			args: args{
				ctx:  context.Background(),
				addr: address,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := New(tt.args.ctx, tt.args.addr); (err != nil) != tt.wantErr {
				t.Errorf("newRedisCache() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_unmarshalBySubKey(t *testing.T) {
	status := pb.Status_STATUS_FINISHED
	statusValue, _ := json.Marshal(status)
	output := "MOCK_OUTPUT"
	outputValue, _ := json.Marshal(output)
	type args struct {
		ctx    context.Context
		subKey cache.SubKey
		value  string
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "status subKey",
			args: args{
				ctx:    context.Background(),
				subKey: cache.Status,
				value:  string(statusValue),
			},
			want:    status,
			wantErr: false,
		},
		{
			name: "runOutput subKey",
			args: args{
				subKey: cache.RunOutput,
				value:  string(outputValue),
			},
			want:    output,
			wantErr: false,
		},
		{
			name: "compileOutput subKey",
			args: args{
				subKey: cache.CompileOutput,
				value:  string(outputValue),
			},
			want:    output,
			wantErr: false,
		},
		{
			name: "graph subKey",
			args: args{
				subKey: cache.Graph,
				value:  string(outputValue),
			},
			want:    output,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unmarshalBySubKey(tt.args.subKey, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("unmarshalBySubKey() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("unmarshalBySubKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}
