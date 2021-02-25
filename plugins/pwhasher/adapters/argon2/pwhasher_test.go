package argon2

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestArgon2_Hash(t *testing.T) {
	type fields struct {
		conf *HashConfig
	}
	type args struct {
		data string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "argon2i default",
			fields:  fields{conf: DefaultConfig},
			args:    args{data: "qwerty"},
			wantErr: false,
		},
		{
			name: "argon2i 32/64",
			fields: fields{conf: &HashConfig{
				Kind:        "argon2i",
				Iterations:  4,
				Parallelism: 4,
				SaltLen:     32,
				KeyLen:      64,
				Memory:      32 * 1024,
			}},
			args:    args{data: "hdg36*/*12bd6"},
			wantErr: false,
		},
		{
			name: "argon2i 64/128",
			fields: fields{conf: &HashConfig{
				Kind:        "argon2i",
				Iterations:  3,
				Parallelism: 2,
				SaltLen:     64,
				KeyLen:      128,
				Memory:      32 * 1024,
			}},
			args:    args{data: "hdg36*/*12bd6"},
			wantErr: false,
		},
		{
			name: "argon2id default",
			fields: fields{conf: &HashConfig{
				Kind:        "argon2id",
				Iterations:  3,
				Parallelism: 2,
				SaltLen:     16,
				KeyLen:      32,
				Memory:      32 * 1024,
			}},
			args:    args{data: "qwerty"},
			wantErr: false,
		},
		{
			name: "argon2id 32/64",
			fields: fields{conf: &HashConfig{
				Kind:        "argon2id",
				Iterations:  1,
				Parallelism: 2,
				SaltLen:     32,
				KeyLen:      64,
				Memory:      32 * 1024,
			}},
			args:    args{data: "hyy167hsjsj-12g"},
			wantErr: false,
		},
		{
			name: "argon2id 64/128",
			fields: fields{conf: &HashConfig{
				Kind:        "argon2id",
				Iterations:  3,
				Parallelism: 2,
				SaltLen:     64,
				KeyLen:      128,
				Memory:      32 * 1024,
			}},
			args:    args{data: "hyy167hsjsj-12g"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Argon2{
				conf: tt.fields.conf,
			}
			got, err := a.HashPw(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPw() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NotEmpty(t, got)
			println(got)
		})
	}
}

func TestArgon2_Compare(t *testing.T) {
	type fields struct {
		conf *HashConfig
	}
	type args struct {
		data string
		hash string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:   "argon2i valid data",
			fields: fields{conf: DefaultConfig},
			args: args{
				data: "qwerty",
				hash: "$argon2i$v=19$m=32768,t=3,p=2$VDkrfTNOys4cBijO2rNTBw$2NP3RaDtHrXrMU+kKlcyTvxjyZOfHYoSAxmUjxS4w1Q",
			},
			want:    true,
			wantErr: false,
		},
		{
			name:   "argon2i invalid data",
			fields: fields{conf: DefaultConfig},
			args: args{
				data: "123456",
				hash: "$argon2i$v=19$m=32768,t=3,p=2$VDkrfTNOys4cBijO2rNTBw$2NP3RaDtHrXrMU+kKlcyTvxjyZOfHYoSAxmUjxS4w1Q",
			},
			want:    false,
			wantErr: false,
		},
		{
			name:   "argon2i invalid pwhasher",
			fields: fields{conf: DefaultConfig},
			args: args{
				data: "qwerty",
				hash: "$v=19$m=32768,t=3,p=2$VDkrfTNOys4cBijO2rNTBw$2NP3RaDtHrXrMU+kKlcyTvxjyZOfHYoSAxmUjxS4w1Q",
			},
			want:    false,
			wantErr: true,
		},
		{
			name:   "argon2i incompatible version",
			fields: fields{conf: DefaultConfig},
			args: args{
				data: "qwerty",
				hash: "$argon2i$v=10$m=32768,t=3,p=2$VDkrfTNOys4cBijO2rNTBw$2NP3RaDtHrXrMU+kKlcyTvxjyZOfHYoSAxmUjxS4w1Q",
			},
			want:    false,
			wantErr: true,
		},
		{
			name:   "argon2id valid data",
			fields: fields{conf: DefaultConfig},
			args: args{
				data: "qwerty",
				hash: "$argon2id$v=19$m=32768,t=3,p=2$7Jr8EtPeJsqJ1RxoxHC4eQ$XfSHQ28xgqc2/2LyE4YEkAI2CIilixOAvjh2Ds2s0+Y",
			},
			want:    true,
			wantErr: false,
		},
		{
			name:   "argon2id invalid data",
			fields: fields{conf: DefaultConfig},
			args: args{
				data: "123456",
				hash: "$argon2id$v=19$m=32768,t=3,p=2$7Jr8EtPeJsqJ1RxoxHC4eQ$XfSHQ28xgqc2/2LyE4YEkAI2CIilixOAvjh2Ds2s0+Y",
			},
			want:    false,
			wantErr: false,
		},
		{
			name:   "argon2id invalid pwhasher",
			fields: fields{conf: DefaultConfig},
			args: args{
				data: "qwerty",
				hash: "$m=32768,t=3,p=2$7Jr8EtPeJsqJ1RxoxHC4eQ$XfSHQ28xgqc2/2LyE4YEkAI2CIilixOAvjh2Ds2s0+Y",
			},
			want:    false,
			wantErr: true,
		},
		{
			name:   "argon2id incompatible version",
			fields: fields{conf: DefaultConfig},
			args: args{
				data: "qwerty",
				hash: "$argon2id$v=17$m=32768,t=3,p=2$7Jr8EtPeJsqJ1RxoxHC4eQ$XfSHQ28xgqc2/2LyE4YEkAI2CIilixOAvjh2Ds2s0+Y",
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Argon2{
				conf: tt.fields.conf,
			}
			got, err := a.ComparePw(tt.args.data, tt.args.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("ComparePw() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ComparePw() got = %v, wantConf %v", got, tt.want)
			}
		})
	}
}
