// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"github.com/gin-gonic/gin"
	"github.com/vincent-petithory/dataurl"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

func apiData(data any) any {
	return gin.H{"data": data}
}

func apiError[T any](code, msg string, payload map[string]T) any {
	ret := gin.H{"code": code, "msg": msg}
	for k, v := range payload {
		ret[k] = v
	}
	return ret
}

func createQR(uri string) (ret string, err error) {
	img, err := qrcode.NewWith(
		uri,
		qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionMedium),
	)
	if err != nil {
		return
	}

	buf := new(buffer)
	w := standard.NewWithWriter(
		buf,
		standard.WithQRWidth(4),
		standard.WithBuiltinImageEncoder(standard.PNG_FORMAT),
	)
	if err = img.Save(w); err != nil {
		return
	}

	ret = dataurl.New(buf.Bytes(), "image/png").String()
	return
}
