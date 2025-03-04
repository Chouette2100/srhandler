/*!
Copyright © 2022 chouette.21.00@gmail.com
Released under the MIT license
https://opensource.org/licenses/mit-license.php

*/

package srhandler

import (
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dustin/go-humanize"

	yaml "gopkg.in/yaml.v2"

	"github.com/Chouette2100/exsrapi/v2"
	"github.com/Chouette2100/srapi/v2"
)

/*

	ApiRoomStatus() の戻り値を表示する。

	Ver. 0.1.0

*/



//	"/ApiRoomStatus"に対するハンドラー
//	http://localhost:8080/t009top で呼び出される
func HandlerApiRoomStatus(
	w http.ResponseWriter,
	r *http.Request,
) {


	//	cookiejarがセットされたHTTPクライアントを作る
	client, jar, err := exsrapi.CreateNewClient("XXXXXX")
	if err != nil {
		log.Printf("CreateNewClient: %s\n", err.Error())
		return
	}
	//	すべての処理が終了したらcookiejarを保存する。
	defer jar.Save()

	//	テンプレートで使用する関数を定義する
	funcMap := template.FuncMap{
		"Comma":         func(i int) string { return humanize.Comma(int64(i)) },                 //	3桁ごとに","を入れる関数。
		"UnixTimeToYYYYMMDDHHMM": func(i int64) string { return time.Unix(int64(i), 0).Format("2006-01-02 15:04") }, //	UnixTimeを時分に変換する関数。
		"UnixTimeToHHMM": func(i int64) string { return time.Unix(int64(i), 0).Format("15:04") }, //	UnixTimeを時分に変換する関数。
	}

	// テンプレートをパースする
	tpl := template.Must(template.New("").Funcs(funcMap).ParseFiles("templates/apiroomstatus.gtpl", "templates/footer.gtpl"))

	url := r.FormValue("room_url_key")
	if url == "" {
		log.Printf("room_url_key が設定されていません。URLのあとに\"?room_url_key=.....\"を追加してください。\n")
		return
	}

	var roomstatus *srapi.RoomStatus
	roomstatus, err = srapi.ApiRoomStatus(client, url)
	if err != nil {
		log.Printf("ApiRoomStatus(): %s\n", err.Error())
		return
	}

	var top struct {
		Roomstatus []string
	}
	//	top.Roomstatus = fmt.Sprintf("%v", roomstatus)
	data, err := yaml.Marshal(roomstatus)
    if err != nil {
        log.Printf(" yaml.Marshal() returned %v\n",err)
		return
    }


	top.Roomstatus = strings.Split(string(data), "\n")
	

	// テンプレートへのデータの埋め込みを行う
	if err = tpl.ExecuteTemplate(w, "apiroomstatus.gtpl", top); err != nil {
		log.Printf("tpl.ExecuteTemplate() returned error: %s\n", err.Error())
	}

}
