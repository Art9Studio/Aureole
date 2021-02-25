package postgresql

import (
	"aureole/plugins/storage"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_pgAdapter_OpenConfig(t *testing.T) {
	adapter := pgAdapter{}

	validConnConf := ConnConfig{
		User:     "root",
		Password: "password",
		Host:     "localhost",
		Port:     "5432",
		Database: "test",
		Options:  map[string]string{},
	}
	invalidConnConf := ConnConfig{
		User:     "",
		Password: "password",
		Host:     "localhost",
		Port:     "",
		Database: "test",
		Options:  map[string]string{},
	}

	usersSess, err := adapter.OpenWithConfig(validConnConf)
	assert.NoError(t, err)
	assert.NotNil(t, usersSess)

	usersSess, err = adapter.OpenWithConfig(invalidConnConf)
	assert.Error(t, err)
	assert.Nil(t, usersSess)
}

func Test_pgAdapter_ParseUrl(t *testing.T) {
	type args struct {
		connUrl string
	}
	tests := []struct {
		name    string
		args    args
		want    storage.ConnConfig
		wantErr bool
	}{
		{
			name: "full raw connection url",
			args: args{
				connUrl: "postgresql://root:password@localhost:5432/test?search_path=public&sslmode=disable",
			},
			want: ConnConfig{
				User:     "root",
				Password: "password",
				Host:     "localhost",
				Port:     "5432",
				Database: "test",
				Options:  map[string]string{"search_path": "public", "sslmode": "disable"},
			},
			wantErr: false,
		},
		{
			name: "raw connection url without userinfo",
			args: args{
				connUrl: "postgresql://localhost:5432/test?search_path=public&sslmode=disable",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "raw connection url without password",
			args: args{
				connUrl: "postgresql://root@localhost:5432/test?search_path=public&sslmode=disable",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "raw connection url without port",
			args: args{
				connUrl: "postgresql://root:password@localhost/test?search_path=public&sslmode=disable",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "raw connection url without host",
			args: args{
				connUrl: "postgresql://root:password@:5432/test?search_path=public&sslmode=disable",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "raw connection url without db name",
			args: args{
				connUrl: "postgresql://root:password@localhost:5432/?search_path=public&sslmode=disable",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "raw connection url without options",
			args: args{
				connUrl: "postgresql://root:password@localhost:5432/test",
			},
			want: ConnConfig{
				User:     "root",
				Password: "password",
				Host:     "localhost",
				Port:     "5432",
				Database: "test",
				Options:  make(map[string]string),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pg := pgAdapter{}
			got, err := pg.ParseUrl(tt.args.connUrl)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseUrl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseUrl() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_pgAdapter_NewConfig(t *testing.T) {
	type args struct {
		data map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    storage.ConnConfig
		wantErr bool
	}{
		{
			name: "full raw connection configs",
			args: args{map[string]interface{}{
				"adapter":  "postgresql",
				"username": "root",
				"password": "password",
				"host":     "localhost",
				"port":     "5432",
				"db_name":  "test",
				"options": map[string]interface{}{
					"sslmode":     "disable",
					"search_path": "public",
				},
			},
			},
			want: ConnConfig{
				User:     "root",
				Password: "password",
				Host:     "localhost",
				Port:     "5432",
				Database: "test",
				Options:  map[string]string{"search_path": "public", "sslmode": "disable"},
			},
			wantErr: false,
		},
		{
			name: "connection configs without userinfo",
			args: args{map[string]interface{}{
				"adapter": "postgresql",
				"host":    "localhost",
				"port":    "5432",
				"db_name": "test",
				"options": map[string]interface{}{
					"sslmode":     "disable",
					"search_path": "public",
				},
			},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "raw connection url without password",
			args: args{map[string]interface{}{
				"adapter":  "postgresql",
				"username": "root",
				"host":     "localhost",
				"port":     "5432",
				"db_name":  "test",
				"options": map[string]interface{}{
					"sslmode":     "disable",
					"search_path": "public",
				},
			},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "raw connection configs without port",
			args: args{map[string]interface{}{
				"adapter":  "postgresql",
				"username": "root",
				"password": "password",
				"host":     "localhost",
				"db_name":  "test",
				"options": map[string]interface{}{
					"sslmode":     "disable",
					"search_path": "public",
				},
			},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "raw connection configs without host",
			args: args{map[string]interface{}{
				"adapter":  "postgresql",
				"username": "root",
				"password": "password",
				"port":     "5432",
				"db_name":  "test",
				"options": map[string]interface{}{
					"sslmode":     "disable",
					"search_path": "public",
				},
			},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "raw connection configs without db name",
			args: args{map[string]interface{}{
				"adapter":  "postgresql",
				"username": "root",
				"password": "password",
				"host":     "localhost",
				"port":     "5432",
				"options": map[string]interface{}{
					"sslmode":     "disable",
					"search_path": "public",
				},
			},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "raw connection configs without options",
			args: args{map[string]interface{}{
				"adapter":  "postgresql",
				"username": "root",
				"password": "password",
				"host":     "localhost",
				"port":     "5432",
				"db_name":  "test",
			},
			},
			want: ConnConfig{
				User:     "root",
				Password: "password",
				Host:     "localhost",
				Port:     "5432",
				Database: "test",
				Options:  make(map[string]string),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pg := pgAdapter{}
			got, err := pg.NewConfig(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
