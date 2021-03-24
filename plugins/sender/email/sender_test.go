package email

import "testing"

func TestEmail_Send(t *testing.T) {
	type fields struct {
		Conf *config
	}
	type args struct {
		recipient    string
		subject      string
		tmplFileName string
		tmplCtx      map[string]interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Email{
				Conf: tt.fields.Conf,
			}
			if err := e.Send(tt.args.recipient, tt.args.subject, tt.args.tmplFileName, tt.args.tmplCtx); (err != nil) != tt.wantErr {
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
			name: "email",
			fields: fields{Conf: &config{
				Host:      "",
				Username:  "",
				Password:  "",
				From:      "",
				Bcc:       nil,
				Cc:        nil,
				Templates: nil,
			}},
			args: args{
				recipient: "",
				subject:   "",
				message:   "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Email{
				Conf: tt.fields.Conf,
			}
			if err := e.SendRaw(tt.args.recipient, tt.args.subject, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("SendRaw() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
