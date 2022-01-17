/*
Copyright 2022 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"testing"
)

func TestEvent_TypeEquals(t *testing.T) {
	type fields struct {
		Type string
	}
	type args struct {
		eventType EventType
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{{
		name: "Should return false if types are different",
		fields: fields{
			Type: "fake.event.type.1",
		},
		args: args{
			eventType: EventType("fake.event.type.2"),
		},
		want: false,
	}, {
		name: "Should return true if types are equal",
		fields: fields{
			Type: "fake.event.type",
		},
		args: args{
			eventType: EventType("fake.event.type"),
		},
		want: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &Event{
				Type: tt.fields.Type,
			}
			if got := event.TypeEquals(tt.args.eventType); got != tt.want {
				t.Errorf("Event.TypeEquals() = %v, want %v", got, tt.want)
			}
		})
	}
}
