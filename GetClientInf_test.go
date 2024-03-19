/*
!
Copyright © 2022 chouette.21.00@gmail.com
Released under the MIT license
https://opensource.org/licenses/mit-license.php
*/
package srhandler

import (
	"fmt"

	"net/http"
	//	"reflect"
	"testing"
)

func TestGetUserInf(t *testing.T) {
	/*
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name    string
		args    args
		wantCi  *ClientInf
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"test1",
			args{},
			&ClientInf{},
			false,
		},
	}
	*/

	http.HandleFunc("/", Handler)
	// サーバーを起動 --- (*5)
	http.ListenAndServe(":8888", nil)

	/*
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCi, err := GetClientInf(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserInf() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCi, tt.wantCi) {
				t.Errorf("GetUserInf() = %v, want %v", gotCi, tt.wantCi)
			}
		})
	}
	*/
}
func Handler(w http.ResponseWriter, r *http.Request) {

	ci, err := GetClientInf(r)

	fmt.Printf("ci=%v\n", ci)
	fmt.Printf("err=%v\n", err)

	htmlBody := "<html><head></head>" +
		"<body><h1>テスト</h1>" +
		"<p>テスト</p>" +
		"<p>テスト</p>" +
		"</body></html>"
	w.Write([]byte(htmlBody))
}
