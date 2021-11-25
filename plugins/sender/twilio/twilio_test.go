package twilio

import (
	"fmt"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestTwilio_Send(test *testing.T) {
	type args struct {
		recipient string
		subject   string
		tmplName  string
		tmplCtx   map[string]interface{}
	}
	tests := []struct {
		name     string
		respCode int
		response interface{}
		conf     *config
		args     args
		wantErr  bool
	}{
		{
			name:     "success send",
			respCode: 200,
			response: "",
			conf: &config{
				AccountSid: "123456",
				AuthToken:  "123456",
				From:       "+380711234567",
				Templates:  map[string]string{"verification": "test/verification.txt"},
			},
			args: args{
				recipient: "+380711234568",
				subject:   "",
				tmplName:  "verification",
				tmplCtx:   map[string]interface{}{"code": "123456"},
			},
			wantErr: false,
		},
		{
			name:     "fail send",
			respCode: 400,
			response: &Exception{
				Status:  400,
				Message: "Bad request",
			},
			conf: &config{
				AccountSid: "123456",
				AuthToken:  "123456",
				From:       "+380711234567",
				Templates:  map[string]string{"verification": "test/verification.txt"},
			},
			args: args{
				recipient: "+380711234568",
				subject:   "",
				tmplName:  "verification",
				tmplCtx:   map[string]interface{}{"code": "123456"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		test.Run(tt.name, func(t1 *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", tt.conf.AccountSid)
			httpmock.RegisterResponder(
				"POST",
				endpoint,
				httpmock.NewJsonResponderOrPanic(tt.respCode, tt.response),
			)

			t := &Twilio{
				conf: tt.conf,
			}
			if err := t.Send(tt.args.recipient, tt.args.subject, tt.args.tmplName, tt.args.tmplCtx); (err != nil) != tt.wantErr {
				t1.Errorf("Send() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTwilio_SendRaw(test *testing.T) {
	type args struct {
		recipient string
		subject   string
		message   string
	}
	tests := []struct {
		name     string
		respCode int
		response interface{}
		conf     *config
		args     args
		wantErr  bool
	}{
		{
			name:     "success raw send",
			respCode: 200,
			response: "",
			conf: &config{
				AccountSid: "123456",
				AuthToken:  "123456",
				From:       "+380711234567",
			},
			args: args{
				recipient: "+380711234568",
				subject:   "",
				message:   "Test message",
			},
			wantErr: false,
		},
		{
			name:     "fail raw send",
			respCode: 400,
			response: &Exception{
				Status:  400,
				Message: "Bad request",
			},
			conf: &config{
				AccountSid: "123456",
				AuthToken:  "123456",
				From:       "+380711234567",
			},
			args: args{
				recipient: "+380711234568",
				subject:   "",
				message:   "Test message",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		test.Run(tt.name, func(t1 *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", tt.conf.AccountSid)
			httpmock.RegisterResponder(
				"POST",
				endpoint,
				httpmock.NewJsonResponderOrPanic(tt.respCode, tt.response),
			)

			t := &Twilio{
				conf: tt.conf,
			}
			if err := t.SendRaw(tt.args.recipient, tt.args.subject, tt.args.message); (err != nil) != tt.wantErr {
				t1.Errorf("SendRaw() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
