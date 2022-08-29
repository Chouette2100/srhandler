/*!
Copyright © 2022 chouette.21.00@gmail.com
Released under the MIT license
https://opensource.org/licenses/mit-license.php

*/

package srhandler

import (
	"html/template"
	//	"io" //　ログ出力設定用。必要に応じて。
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/Chouette2100/exsrapi"
	"github.com/Chouette2100/srapi"
)

/*

	配信中のルームを開始時刻の降順にソートして表示するためのハンドラー

	Ver. 0.1.0
	Ver. 0.4.0 デフォルトの表示カテゴリーをFreeからOfficialに変更する。

*/


type T009Config struct {
	SR_acct      string //	SHOWROOMのアカウント名
	SR_pswd      string //	SHOWROOMのパスワード
	Category     string //	カテゴリー名
	Aplmin       int    //	訪問ルームリストの有効時間(分)
	Maxnoroom    int    //	訪問候補ルームリストの最大長
	Rvlfilename  string //	訪問済みルームリストファイル名
	Exclfilename string //	除外ルームリストファイル名
}

type T009top struct {
	TimeNow      int64
	SR_acct      string //	SHOWROOMのアカウント名（必須ではない）
	Category     string //	カテゴリー名
	Aplmin       int    //	訪問ルームリストの有効時間(分)
	Maxnoroom    int    //	訪問候補ルームリストの最大長
	Rvlfilename  string //	訪問済みルームリストファイル名
	Exclfilename string //	除外ルームリストファイル名
	ErrMsg       string
	Lives        []srapi.Live //	配信中ルーム情報	（V2ではポインターとはしない）
}

//	"/t009top"に対するハンドラー
//	http://localhost:8080/t009top で呼び出される
func HandlerT009topForm(
	w http.ResponseWriter,
	r *http.Request,
) {

	top := T009top{
		SR_acct:      "999999",
		Category:     "Official",
		Aplmin:       240,
		Maxnoroom:    20,
		Rvlfilename:  "rvl.txt",
		Exclfilename: "excl.txt",
	}

	roomlives := new(srapi.RoomOnlives)

	//	cookiejarがセットされたHTTPクライアントを作る
	client, jar, err := exsrapi.CreateNewClient(top.SR_acct)
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
		"GidToName": func(gid int) string {
			for j := 0; j < len(roomlives.Onlives); j++ {
				if roomlives.Onlives[j].Genre_id == gid {
					return roomlives.Onlives[j].Genre_name
				}
			}
			return "n/a"
		}, //	ジャンルIDをジャンル名に変換する関数
	}

	// テンプレートをパースする
	tpl := template.Must(template.New("").Funcs(funcMap).ParseFiles("templates/t009top.gtpl"))

	category := r.FormValue("category")
	if category != "" {
		top.Category = category
	}

	sr_acct := r.FormValue("sr_acct")
	if sr_acct != "" {
		top.SR_acct = sr_acct
	}

	maxnoroom := r.FormValue("maxnoroom")
	if maxnoroom != "" {
		mno, err := strconv.Atoi(maxnoroom)
		if err != nil {
			log.Printf("strconv.Sto(): %s\n", err.Error())
		} else {
			top.Maxnoroom = mno
		}
	}

	top.TimeNow = time.Now().Unix()

	//	配信しているルームの一覧を取得する
	roomlives, err = srapi.ApiLiveOnlives(client)
	if err != nil {
		log.Printf("ApiLiveOnlives(): %s\n", err.Error())
		return
	}
	log.Printf("*****************************************************************\n")
	log.Printf("配信中ルーム数\n")
	log.Printf("\n")
	log.Printf("　ジャンル数= %d\n", len(roomlives.Onlives))
	log.Printf("\n")
	log.Printf("　ルーム数　ジャンル　ジャンル名\n")
	for _, roomlive := range roomlives.Onlives {
		log.Printf("%10d%10d  %s\n", len(roomlive.Lives), roomlive.Genre_id, roomlive.Genre_name)
	}
	log.Printf("\n")

	roomvisit := new(exsrapi.RoomVisit) //	訪問ルームリスト
	roomvisit.Roomvisit = make(map[int]time.Time)

	excllist := exsrapi.ExclList{}                      //	除外ルームリスト
	err = excllist.Read(top.Category, top.Exclfilename) //	除外ルームリストを読み込む
	if err != nil {
		log.Printf("excllist.Read(): %s\n", err.Error())
		return
	}

	//	訪問ルームリストファイルからすでに星集め、種集めのために訪問したリストを読み込む
	err = roomvisit.Restore(top.Category, top.Rvlfilename, top.Aplmin)
	if err != nil {
		log.Printf("RestoreRVL(): %s\n", err.Error())
		return
	}
	defer roomvisit.Save() //	訪問したリストを保存する。本来星集め、種集めが終わったあと行う処理。

	//	星集め/種集めの対象とするルームのリストを作る
	lives, err := exsrapi.MkRoomsForStarCollec(client, top.Category, top.Aplmin, top.Maxnoroom, &excllist, &roomvisit.Roomvisit)
	if err != nil {
		log.Printf("MkRoomsForStarCollec(): %s\n", err.Error())
		return
	}
	log.Printf(" lenght of lives = %d\n", len(*lives))
	top.Lives = *lives

	// テンプレートへのデータの埋め込みを行う
	if err = tpl.ExecuteTemplate(w, "t009top.gtpl", top); err != nil {
		log.Printf("tpl.ExecuteTemplate() returned error: %s\n", err.Error())
	}

}

