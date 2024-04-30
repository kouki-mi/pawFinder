package main

import (
	"reflect"
	"testing"
)

func Test_validateRequest(t *testing.T) {
	type args struct {
		request string
	}
	tests := []struct {
		name    string
		args    args
		want    LineEvents
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "SuccessCase",
			args: args{
				request: `{
					"events": [
						{
							"type": "message",
							"replyToken": "nHuyWiB7yP5Zw",
							"message": {
								"type": "text",
								"id": "325708",
								"text": "Hello, world"
							}
						}
					]
				}`,
			},
			want: LineEvents{
				Events: []struct {
					Type      string `json:"type"`
					ReplyToken string `json:"replyToken"`
					Message   struct {
						Type string `json:"type"`
						Id string 	`json:"id"`
						Text string `json:"text"`
					} `json:"message"`
				}{
					{
						Type: "message",
						ReplyToken: "nHuyWiB7yP5Zw",
						Message: struct {
							Type string `json:"type"`
							Id string 	`json:"id"`
							Text string `json:"text"`
						}{
							Type: "text",
							Id: "325708",
							Text: "Hello, world",
						},
					},
				},
			},
			wantErr: false,
		},{
			name: "FailedCase",
			args: args{
				request: ``,
			},
			want: LineEvents{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unmarshalRequest(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("validateRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
