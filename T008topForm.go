/*!
Copyright © 2022 chouette.21.00@gmail.com
Released under the MIT license
https://opensource.org/licenses/mit-license.php

*/

package srhandler

import (
	"fmt"
	"html/template"
	//	"io" //　ログ出力設定用。必要に応じて。
	"log"
	"net/http"
	"sort" //	ソート用。必要に応じて。
	"time"

	"github.com/dustin/go-humanize"

	"github.com/Chouette2100/exsrapi"
	"github.com/Chouette2100/srapi"
)

/*
$ tar xvf v0.1.0.tar.gz
	$ mv t009srapi-0.1.0 t009srapi
	$ cd t009srapi
	$ go mod init
	$ go mod tidy
	$ go build t009srapi.go
	$ ./t009srapi

	ここでブラウザを起動し
	　　		http://localhost:8080/t009top
	で、実行時点でのイベントの一覧が表示されます。

	【Windows】

	Microsoft Windows [Version 10.0.22000.856]
	(c) Microsoft Corporation. All rights reserved.

	C:\Users\chouette>cd go

	C:\Users\chouette\go>cd src

	作業はかならず %HOMEPATH%\go\src の下で行います。

	以下、要するに https://github.com/Chouette2100/t009srapi/releases にあるv0.1.0のZIPファイルSource code (zip) からソースをとりだしてくださいということなので、ブラウザでダウンロードしてエクスプローラで解凍というこでもけっこうです。なんならこの記事の最後にあるgithubのソースをエディターにコピペで作るということでもかまいません（この場合文字コードはかならずUTF-8にしてください 改行はLFになっています。というようなことを考えるとやっぱりダウンロードして解凍が安全かも）

	C:\Users\chouette\go\src>mkdir t009srapi

	C:\Users\chouette\go\src>cd t009srapi

	C:\Users\chouette\go\src\t009srapi>curl -OL https://github.com/Chouette2100/t009srapi/archive/refs/tags/v0.1.0.zip
	  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
	                                 Dload  Upload   Total   Spent    Left  Speed
	  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0
	100  6265    0  6265    0     0   6777      0 --:--:-- --:--:-- --:--:-- 16400

	C:\Users\chouette\go\src\t009srapi>call powershell -command "Expand-Archive v0.1.0.zip"

	C:\Users\chouette\go\src\t009srapi>tree
	フォルダー パスの一覧
	ボリューム シリアル番号は E2CD-BDF1 です
	C:.
	└─v0.1.0
	    └─t009srapi-0.1.0
	        ├─public
	        └─templates

	C:\Users\chouette\go\src\t009srapi>xcopy /e v0.1.0\t009srapi-0.1.0\*.*
	v0.1.0\t007srapi-0.1.0\freebsd.bat
	v0.1.0\t007srapi-0.1.0\freebsd.sh
	v0.1.0\t007srapi-0.1.0\LICENSE
	v0.1.0\t007srapi-0.1.0\README.md
	v0.1.0\t007srapi-0.1.0\t007srapi.go
	v0.1.0\t007srapi-0.1.0\public\index.html
	v0.1.0\t007srapi-0.1.0\templates\top.gtpl
	7 File(s) copied

	C:\Users\chouette\go\src\t009srapi>rmdir /s /q v0.1.0

	C:\Users\chouette\go\src\t009srapi>del v0.1.0.zip

	ここで次のような構成になっていればOKです。top.gtpl と index.html が所定の場所にあることをかならず確かめてください。

	C:%HOMEPATH%\go\src\t009srapi --+-- t009srapi.go
	                                |
	                                +-- \templates --- t009top.gtpl
	                                |
	                                +-- \public    --- index.html

	ここからはコマンド三つでビルドが完了します。

	C:\Users\chouette\go\src\t009srapi>go mod init
	go: creating new go.mod: module t009srapi
	go: to add module requirements and sums:
	        go mod tidy

	C:\Users\chouette\go\src\t009srapi>go mod tidy
	go: finding module for package github.com/dustin/go-humanize
	go: downloading github.com/dustin/go-humanize v1.0.0
	go: found github.com/dustin/go-humanize in github.com/dustin/go-humanize v1.0.0

	C:\Users\chouette\go\src\t009srapi>go build t009srapi.go

	あとは

	C:\Users\chouette\go\src\t009srapi>t009srapi

	でWebサーバが起動します。ここでセキュリティー上の警告が出ると思いますが、説明をよく読んで問題ないと思ったらアクセスを許可してください（もちろん許可しなければWebサーバは使えなくなります）

	Webサーバを起動したままにしておいてブラウザを開き

	http://localhost:8080/t009top

	で、実行時点でのイベントの一覧が表示されます。

	Ver. 0.1.0

*/

type T008top struct {
	TimeNow    int64
	Totalcount int
	ErrMsg     string
	Eventlist  []srapi.Event
}

//	"/t008top"に対するハンドラー
//	http://localhost:8080/t008top で呼び出される
func HandlerT008topForm(
	w http.ResponseWriter,
	r *http.Request,
) {

	client, cookiejar, err := exsrapi.CreateNewClient("")
	if err != nil {
		log.Printf("exsrapi.CeateNewClient(): %s", err.Error())
		return //	エラーがあれば、ここで終了
	}
	defer cookiejar.Save()

	//	テンプレートで使用する関数を定義する
	funcMap := template.FuncMap{
		"Comma":         func(i int) string { return humanize.Comma(int64(i)) },                       //	3桁ごとに","を入れる関数。
		"UnixTimeToStr": func(i int64) string { return time.Unix(int64(i), 0).Format("01-02 15:04") }, //	UnixTimeを年月日時分に変換する関数。
	}

	// テンプレートをパースする
	tpl := template.Must(template.New("").Funcs(funcMap).ParseFiles("templates/t008top.gtpl"))

	// テンプレートに埋め込むデータ（ポイントやランク）を作成する
	top := new(T008top)
	top.TimeNow = time.Now().Unix()

	top.Eventlist, err = srapi.MakeEventListByApi(client)
	if err != nil {
		err = fmt.Errorf("MakeListOfPoints(): %w", err)
		log.Printf("MakeListOfPoints() returned error %s\n", err.Error())
		top.ErrMsg = err.Error()
	}
	top.Totalcount = len(top.Eventlist)

	//	ソートが必要ないときは次の行とimportの"sort"をコメントアウトする。
	//	無名関数のリターン値でソート条件を変更できます。
	sort.Slice(top.Eventlist, func(i, j int) bool { return top.Eventlist[i].Ended_at > top.Eventlist[j].Ended_at })

	// テンプレートへのデータの埋め込みを行う
	if err = tpl.ExecuteTemplate(w, "t008top.gtpl", top); err != nil {
		log.Printf("tpl.ExecuteTemplate() returned error: %s\n", err.Error())
	}

}

