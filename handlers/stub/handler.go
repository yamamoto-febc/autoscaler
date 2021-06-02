// Copyright 2021 The sacloud Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stub

import (
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/version"
)

// Handler 単体テスト用のスタブハンドラ
type Handler struct {
	PreHandleFunc  func(*handler.PreHandleRequest, handlers.ResponseSender) error
	HandleFunc     func(*handler.HandleRequest, handlers.ResponseSender) error
	PostHandleFunc func(*handler.PostHandleRequest, handlers.ResponseSender) error
}

func (h *Handler) Name() string {
	return "stub"
}

func (h *Handler) Version() string {
	return version.FullVersion()
}

func (h *Handler) PreHandle(req *handler.PreHandleRequest, sender handlers.ResponseSender) error {
	if h.PreHandleFunc != nil {
		return h.PreHandleFunc(req, sender)
	}
	return nil
}

func (h *Handler) Handle(req *handler.HandleRequest, sender handlers.ResponseSender) error {
	if h.HandleFunc != nil {
		return h.HandleFunc(req, sender)
	}
	return nil
}

func (h *Handler) PostHandle(req *handler.PostHandleRequest, sender handlers.ResponseSender) error {
	if h.PostHandleFunc != nil {
		return h.PostHandleFunc(req, sender)
	}
	return nil
}
