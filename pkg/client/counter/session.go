// Copyright 2019-present Open Networking Foundation.
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

package counter

import (
	"context"
	api "github.com/atomix/atomix-api/proto/atomix/counter"
	"github.com/atomix/atomix-go-client/pkg/client/session"
)

type sessionHandler struct {
	client api.CounterServiceClient
}

func (m *sessionHandler) Create(ctx context.Context, s *session.Session) error {
	request := &api.CreateRequest{
		Header: s.GetState(),
	}
	_, err := m.client.Create(ctx, request)
	if err != nil {
		return err
	}
	return nil
}

func (m *sessionHandler) KeepAlive(ctx context.Context, s *session.Session) error {
	return nil
}

func (m *sessionHandler) Close(ctx context.Context, s *session.Session) error {
	request := &api.CloseRequest{
		Header: s.GetState(),
	}
	_, err := m.client.Close(ctx, request)
	return err
}

func (m *sessionHandler) Delete(ctx context.Context, s *session.Session) error {
	request := &api.CloseRequest{
		Header: s.GetState(),
		Delete: true,
	}
	_, err := m.client.Close(ctx, request)
	return err
}
