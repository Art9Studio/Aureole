package postgresql

import (
	"testing"
)

func Test_ConnectionConfig_String(t *testing.T) {
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
				Options:  map[string]string{},
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
				Options:  map[string]string{},
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
				Options:  map[string]string{},
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
				Options:  map[string]string{},
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
				Options:  map[string]string{},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connConf := ConnConfig{
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
