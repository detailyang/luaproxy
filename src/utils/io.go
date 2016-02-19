/*
* @Author: detailyang
* @Date:   2016-02-11 22:45:38
* @Last Modified by:   detailyang
* @Last Modified time: 2016-02-19 16:59:40
 */

package utils

import (
	"io"
)

type StringReaderCloser struct {
	io.Reader
}

func (StringReaderCloser) Close() error {
	return nil
}
