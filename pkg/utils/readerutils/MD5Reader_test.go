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

package readerutils

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMD5Reader(t *testing.T) {
	buf := bytes.NewBufferString("abc")
	reader := NewMD5Reader(buf)
	assert.NotNil(t, reader)
	assert.Equal(t, "\xd4\x1d\x8cُ\x00\xb2\x04\xe9\x80\t\x98\xec\xf8B~", string(reader.MD5()))
	count, err := reader.Read([]byte("abc"))
	assert.Equal(t, 3, count)
	assert.Nil(t, err)
}
