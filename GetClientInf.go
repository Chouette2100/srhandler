/*
!
Copyright © 2022 chouette.21.00@gmail.com
Released under the MIT license
https://opensource.org/licenses/mit-license.php
*/
package srhandler

import (
	"log"
	"runtime"
	"strings"
	"time"

	"net/http"
)

type ClientInf struct {
	Actime time.Time //	実行時の日時
	Fn     string    //	実行中の関数（＝親の関数）の名称
	Ra     string    //	リモートアドレス
	Port   string    //	リモートポート
	Ua     string    //	ユーザーエージェント
}

/*
ファンクション名とリモートアドレス、ユーザーエージェントを表示する。
*/
func GetClientInf(
	r *http.Request,
) (
	ci *ClientInf,
	err error,
) {

	ci = &ClientInf{}

	//	実行時日時を取得する。
	ci.Actime = time.Now()

	//	実行中の関数（＝親の関数）の名称を求める
	pt, _, _, ok := runtime.Caller(1) //	スタックトレースへのポインターを得る。1は一つ上のファンクション。
	fn := ""
	if !ok {
		fn = "unknown"
	} else {
		fn = runtime.FuncForPC(pt).Name()
	}
	fna := strings.Split(fn, ".")
	ci.Fn = fna[len(fna)-1]

	ra := r.RemoteAddr
	raa := strings.Split(ra, ":")
	ci.Ra = raa[0]
	ci.Port = raa[1]

	ci.Ua = r.UserAgent()

	if false {
		//	ログ出力の例
		log.Printf("***** %s() from %s by %s\n", ci.Fn, ci.Ra+":"+ci.Port, ci.Ua)
	}

	return
}
