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

package set

import (
	api "github.com/atomix/atomix-api/go/atomix/primitive/set"
	"github.com/atomix/atomix-go-client/pkg/atomix/primitive"
)

// Option is a set option
type Option interface {
	primitive.Option
	applyNewSet(options *newSetOptions)
}

// newSetOptions is set options
type newSetOptions struct {
	clientID string
}

// WithClientID sets the client identifier
func WithClientID(id string) Option {
	return &clientIDOption{
		clientID: id,
	}
}

type clientIDOption struct {
	primitive.EmptyOption
	clientID string
}

func (o *clientIDOption) applyNewSet(options *newSetOptions) {
	options.clientID = o.clientID
}

// WatchOption is an option for set Watch calls
type WatchOption interface {
	beforeWatch(request *api.EventsRequest)
	afterWatch(response *api.EventsResponse)
}

// WithReplay returns a Watch option to replay entries
func WithReplay() WatchOption {
	return replayOption{}
}

type replayOption struct{}

func (o replayOption) beforeWatch(request *api.EventsRequest) {
	request.Replay = true
}

func (o replayOption) afterWatch(response *api.EventsResponse) {

}
