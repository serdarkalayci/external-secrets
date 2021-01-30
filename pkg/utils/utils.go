/*
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

package utils

import (
	"crypto/rand"
	"math/big"

	ctrl "sigs.k8s.io/controller-runtime"
)

// Merge maps.
func Merge(src, dst map[string][]byte) map[string][]byte {
	for k, v := range dst {
		src[k] = v
	}
	return src
}

const validObjChars = "0123456789abcdefghijklmnopqrstuvwxyz"

var (
	log = ctrl.Log.WithName("asm")
)

// RandomBytes generate random bytes
func RandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// RandomInt returns a random int64
func RandomInt() (int64, error) {
	randomInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(validObjChars))))
	if err != nil {
		return 0, err
	}

	return randomInt.Int64(), nil
}

// RandomStringObjectSafe returns a random string that is safe to use as an k8s object Name
//  https://kubernetes.io/docs/concepts/overview/working-with-objects/names/
func RandomStringObjectSafe(n int) (string, error) {
	b, err := RandomBytes(n)
	if err != nil {
		return "", err
	}

	for i := range b {
		randomInt, err := RandomInt()
		if err != nil {
			return "", err
		}
		b[i] = validObjChars[randomInt]
	}
	return string(b), nil

}
