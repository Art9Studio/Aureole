package email

import "testing"

func TestEmail_Send(t *testing.T) {
	type fields struct {
		Conf *config
	}
	type args struct {
		recipient string
		subject   string
		tmplName  string
		tmplCtx   map[string]interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "email plaintext template",
			fields: fields{Conf: &config{
				Host:      "smtp.gmail.com:587",
				Username:  "test.aureole@gmail.com",
				Password:  "aureole_secret",
				From:      "test.aureole@gmail.com",
				Bcc:       nil,
				Cc:        nil,
				Templates: map[string]string{"default_txt": "../../../templates/default.txt"},
			}},
			args: args{
				recipient: "test.aureole@gmail.com",
				subject:   "Send test with plaintext",
				tmplName:  "default_txt",
				tmplCtx:   map[string]interface{}{"name": "Andrew"},
			},
			wantErr: false,
		},
		{
			name: "email html template",
			fields: fields{Conf: &config{
				Host:      "smtp.gmail.com:587",
				Username:  "test.aureole@gmail.com",
				Password:  "aureole_secret",
				From:      "test.aureole@gmail.com",
				Bcc:       nil,
				Cc:        nil,
				Templates: map[string]string{"default_html": "../../../templates/default.html"},
			}},
			args: args{
				recipient: "test.aureole@gmail.com",
				subject:   "Send test with html",
				tmplName:  "default_html",
				tmplCtx:   map[string]interface{}{"name": "Andrew"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &sender{
				conf: tt.fields.Conf,
			}
			if err := e.Send(tt.args.recipient, tt.args.subject, tt.args.tmplName, tt.args.tmplCtx); (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEmail_SendRaw(t *testing.T) {
	type fields struct {
		Conf *config
	}
	type args struct {
		recipient string
		subject   string
		message   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "raw email",
			fields: fields{Conf: &config{
				Host:      "smtp.gmail.com:587",
				Username:  "test.aureole@gmail.com",
				Password:  "aureole_secret",
				From:      "test.aureole@gmail.com",
				Bcc:       nil,
				Cc:        nil,
				Templates: nil,
			}},
			args: args{
				recipient: "test.aureole@gmail.com",
				subject:   "SendRaw test",
				message:   "this is test message",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &sender{
				conf: tt.fields.Conf,
			}
			if err := e.SendRaw(tt.args.recipient, tt.args.subject, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("SendRaw() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
