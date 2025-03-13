/*
Copyright 2019-2022 The KubeSphere Authors.

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

package hashutil

import (
	"encoding/hex"
	"io"

	"code.cloudfoundry.org/bytefmt"
	"github.com/kubesphere/ks-devops/pkg/utils/readerutils"
)

// GetMD5 returns the md5 value from a reader
func GetMD5(reader io.ReadCloser) (result string, err error) {
	md5reader := readerutils.NewMD5Reader(reader)
	data := make([]byte, bytefmt.KILOBYTE)
	for {
		_, err = md5reader.Read(data)
		if err != nil {
			if err == io.EOF {
				break
			}
			return
		}
	}
	if err = reader.Close(); err == nil {
		result = hex.EncodeToString(md5reader.MD5())
	}
	return
}
