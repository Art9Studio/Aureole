package postgresql

import (
	"reflect"
	"testing"
)

func TestString(t *testing.T) {
	type fields struct {
		Adapter  string
		User     string
		Password string
		Host     string
		Port     string
		Database string
		Options  map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "full config to sring",
			fields: fields{
				User:     "admin",
				Password: "admin",
				Host:     "localhost",
				Port:     "5432",
				Database: "gouth",
				Options:  map[string]string{"search_path": "public"},
			},
			want:    "postgresql://admin:admin@localhost:5432/gouth?search_path=public",
			wantErr: false,
		},
		{
			name: "config without opts to sring",
			fields: fields{
				User:     "admin",
				Password: "admin",
				Host:     "localhost",
				Port:     "5432",
				Database: "gouth",
				Options:  nil,
			},
			want:    "postgresql://admin:admin@localhost:5432/gouth",
			wantErr: false,
		},
		{
			name: "config without userinfo to sring",
			fields: fields{
				Adapter:  "postgresql",
				Host:     "localhost",
				Port:     "5432",
				Database: "gouth",
				Options:  nil,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "config without host to sring",
			fields: fields{
				User:     "admin",
				Password: "admin",
				Port:     "5432",
				Database: "gouth",
				Options:  nil,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "config without port to sring",
			fields: fields{
				User:     "admin",
				Password: "admin",
				Host:     "localhost",
				Database: "gouth",
				Options:  nil,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "config without db name to sring",
			fields: fields{
				User:     "admin",
				Password: "admin",
				Host:     "localhost",
				Port:     "5432",
				Options:  nil,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connConf := ConnectionConfig{
				User:     tt.fields.User,
				Password: tt.fields.Password,
				Host:     tt.fields.Host,
				Port:     tt.fields.Port,
				Database: tt.fields.Database,
				Options:  tt.fields.Options,
			}
			got, err := connConf.String()
			if (err != nil) != tt.wantErr {
				t.Errorf("String() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("String() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseUrl(t *testing.T) {
	type args struct {
		connUrl string
	}
	tests := []struct {
		name         string
		args         args
		wantConnConf ConnectionConfig
		wantErr      bool
	}{
		{
			name: "parse full url",
			args: args{
				connUrl: "postgresql://admin:admin@localhost:5432/gouth?search_path=public&sslmode=disable",
			},
			wantConnConf: ConnectionConfig{
				User:     "admin",
				Password: "admin",
				Host:     "localhost",
				Port:     "5432",
				Database: "gouth",
				Options:  map[string]string{"search_path": "public", "sslmode": "disable"},
			},
			wantErr: false,
		},
		{
			name: "parse url without adapter",
			args: args{
				connUrl: "//admin:admin@localhost:5432/gouth?search_path=public",
			},
			wantConnConf: ConnectionConfig{},
			wantErr:      true,
		},
		{
			name: "parse url without userinfo",
			args: args{
				connUrl: "postgresql://localhost:5432/gouth?search_path=public",
			},
			wantConnConf: ConnectionConfig{},
			wantErr:      true,
		},
		{
			name: "parse url without port",
			args: args{
				connUrl: "postgresql://admin:admin@localhost/gouth?search_path=public",
			},
			wantConnConf: ConnectionConfig{},
			wantErr:      true,
		},
		{
			name: "parse url without host",
			args: args{
				connUrl: "postgresql://admin:admin@:5432/gouth?search_path=public",
			},
			wantConnConf: ConnectionConfig{},
			wantErr:      true,
		},
		{
			name: "parse url without db name",
			args: args{
				connUrl: "postgresql://admin:admin@localhost:5432/?search_path=public",
			},
			wantConnConf: ConnectionConfig{},
			wantErr:      true,
		},
		{
			name: "parse url without opts",
			args: args{
				connUrl: "postgresql://admin:admin@localhost:5432/gouth",
			},
			wantConnConf: ConnectionConfig{
				User:     "admin",
				Password: "admin",
				Host:     "localhost",
				Port:     "5432",
				Database: "gouth",
				Options:  make(map[string]string),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotConnConf, err := ParseUrl(tt.args.connUrl)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseUrl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotConnConf, tt.wantConnConf) {
				t.Errorf("ParseUrl() gotConnConf = %v, want %v", gotConnConf, tt.wantConnConf)
			}
		})
	}
}
