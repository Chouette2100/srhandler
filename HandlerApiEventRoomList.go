/*!
Copyright © 2022 chouette.21.00@gmail.com
Released under the MIT license
https://opensource.org/licenses/mit-license.php

*/

package srhandler

import (
	//	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"html/template"
	"net/http"

	"github.com/dustin/go-humanize"

	"github.com/Chouette2100/exsrapi"
	"github.com/Chouette2100/srapi"
)

/*

	ApiEventRoomList() の戻り値を表示する。

	Ver. 0.1.0

*/

//	"/ApiEventRoomList()"に対するハンドラー
//	http://localhost:8080/apieventroomlist で呼び出される
func HandlerApiEventRoomList(
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
		"Comma":          func(i int) string { return humanize.Comma(int64(i)) },                           //	3桁ごとに","を入れる関数。
		"UnixtimeToTime": func(i int64, tfmt string) string { return time.Unix(int64(i), 0).Format(tfmt) }, //	UnixTimeを時分に変換する関数。
	}

	// テンプレートをパースする
	tpl := template.Must(template.New("").Funcs(funcMap).ParseFiles("templates/apieventroomlist.gtpl", "templates/footer.gtpl"))

	var erl struct {
		Eventid     int
		Eventname   string
		Eventurl    string
		Ib		int
		Ie		int
		Roomlistinf *srapi.RoomListInf
		Msg         string
		Eventlist 	[]srapi.Event
	}

	seventid := r.FormValue("eventid")
	if seventid == "" {
		/*
		err = errors.New("eventid が設定されていません。URLのあとに\"?eventid=.....\"を追加してください。<br>あるいは「開催中イベント一覧表」から参加者一覧が必要なイベントを指定してください。")
		erl.Msg = err.Error()
		log.Printf("%s\n", erl.Msg)
		*/
		erl.Eventid = 0
		erl.Eventlist, err = srapi.MakeEventListByApi(client)
		if err != nil {
			err = fmt.Errorf("MakeListOfPoints(): %w", err)
			log.Printf("MakeListOfPoints() returned error %s\n", err.Error())
			erl.Msg = err.Error()
		}
		//	erl.Totalcount = len(top.Eventlist)
	
		//	ソートが必要ないときは次の行とimportの"sort"をコメントアウトする。
		//	無名関数のリターン値でソート条件を変更できます。
		//	ここではエベント終了日時が近い順にソートしています。
		sort.Slice(erl.Eventlist, func(i, j int) bool { return erl.Eventlist[i].Ended_at < erl.Eventlist[j].Ended_at })
	

	} else {

		erl.Eventid, err = strconv.Atoi(seventid)
		if err != nil {
			err = fmt.Errorf("HandlerEventRoomList(): %w", err)
			erl.Msg = err.Error()
			log.Printf("%s\n", erl.Msg)
		} else {

		sib := r.FormValue("ib")
		erl.Ib, err = strconv.Atoi(sib)
		if err != nil {
			erl.Ib = 1
		}

		sie := r.FormValue("ie")
		erl.Ie, err = strconv.Atoi(sie)
		if err != nil {
			erl.Ie = 10
		}

		if erl.Ie < erl.Ib {
			erl.Ie = erl.Ib
		}


			//	イベント参加ルーム一覧を取得する。
			erl.Roomlistinf, err = srapi.GetRoominfFromEventByApi(client, erl.Eventid, erl.Ib, erl.Ie)
			if err != nil {
				err = fmt.Errorf("HandlerEventRoomList(): %w", err)
				erl.Msg = err.Error()
				log.Printf("%s\n", erl.Msg)
			}

			//	ルーム一覧にあるそれぞれのルームについて補足的なデータを取得する。
			lrank := -1
			rank := -1
			for i, room := range erl.Roomlistinf.RoomList {
				if i == 0 {
					//	最初のルーム
					//	順位、ポイント、上位との差とイベント名、イベントのURLを取得する。
					//	DBを使っているときはイベント名とイベントのURLはイベントマスターから取得すべき。
					erl.Roomlistinf.RoomList[i].Point, erl.Roomlistinf.RoomList[i].Rank, erl.Roomlistinf.RoomList[i].Gap,
						_, erl.Eventurl, erl.Eventname, _, err = srapi.GetPointByApi(client, room.Room_id)
					//	erl.Roomlistinf.RoomList[i].Gap = -1
				} else {
					//	2番目以降のルーム
					//	順位、ポイント、上位との差を取得する。
					erl.Roomlistinf.RoomList[i].Point, rank, erl.Roomlistinf.RoomList[i].Gap,
						_, _, _, _, err = srapi.GetPointByApi(client, room.Room_id)
					if rank == lrank {
						erl.Roomlistinf.RoomList[i].Rank = -1
					} else {
						erl.Roomlistinf.RoomList[i].Rank = rank
					}
					lrank = rank
				}
				if err != nil {
					err = fmt.Errorf("HandlerEventRoomList(): %w", err)
					erl.Msg = err.Error()
					log.Printf("%s\n", erl.Msg)
					break
				}

				//	ルーム状況（配信中か、配信開始時刻、公式か）を取得する。
				roomstatus, err := srapi.ApiRoomStatus(client, room.Room_url_key)
				if err != nil {
					err = fmt.Errorf("HandlerEventRoomList(): %w", err)
					erl.Msg = err.Error()
					log.Printf("%s\n", erl.Msg)
					break
				}
				erl.Roomlistinf.RoomList[i].Islive = roomstatus.Is_live
				erl.Roomlistinf.RoomList[i].Isofficial = roomstatus.Is_official
				erl.Roomlistinf.RoomList[i].Startedat = roomstatus.Started_at

				//	次枠配信開始時刻を取得する。
				roomnextlive, err := srapi.ApiRoomNextlive(client, room.Room_id)
				if err != nil {
					err = fmt.Errorf("HandlerEventRoomList(): %w", err)
					erl.Msg = err.Error()
					log.Printf("%s\n", erl.Msg)
					break
				}
				erl.Roomlistinf.RoomList[i].Nextlive = roomnextlive.Epoch

			}
		}
	}

	// テンプレートへのデータの埋め込みを行う
	if err = tpl.ExecuteTemplate(w, "apieventroomlist.gtpl", erl); err != nil {
		err = fmt.Errorf("HandlerEventRoomList(): %w", err)
		log.Printf("%s\n", err.Error())
	}

}
