package pbkdf2

import (
	"crypto/sha256"
	"crypto/sha512"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPbkdf2_HashPw(t *testing.T) {
	type fields struct {
		conf *HashConfig
	}
	type args struct {
		pw string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "pbkdf2 sha1",
			fields:  fields{conf: DefaultConfig},
			args:    args{pw: "qwerty"},
			wantErr: false,
		},
		{
			name: "pbkdf2 sha224",
			fields: fields{conf: &HashConfig{
				Iterations: 4096,
				SaltLen:    16,
				KeyLen:     32,
				FuncName:   "sha224",
				Function:   sha256.New224,
			}},
			args:    args{pw: "qwerty"},
			wantErr: false,
		},
		{
			name: "pbkdf2 sha256",
			fields: fields{conf: &HashConfig{
				Iterations: 4096,
				SaltLen:    16,
				KeyLen:     32,
				FuncName:   "sha256",
				Function:   sha256.New,
			}},
			args:    args{pw: "qwerty"},
			wantErr: false,
		},
		{
			name: "pbkdf2 sha384",
			fields: fields{conf: &HashConfig{
				Iterations: 4096,
				SaltLen:    16,
				KeyLen:     32,
				FuncName:   "sha384",
				Function:   sha512.New384,
			}},
			args:    args{pw: "qwerty"},
			wantErr: false,
		},
		{
			name: "pbkdf2 sha512",
			fields: fields{conf: &HashConfig{
				Iterations: 4096,
				SaltLen:    16,
				KeyLen:     32,
				FuncName:   "sha512",
				Function:   sha512.New,
			}},
			args:    args{pw: "qwerty"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Pbkdf2{
				conf: tt.fields.conf,
			}
			got, err := p.HashPw(tt.args.pw)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPw() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NotEmpty(t, got)
			println(got)
		})
	}
}

func TestPbkdf2_ComparePw(t *testing.T) {
	type fields struct {
		conf *HashConfig
	}
	type args struct {
		pw   string
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
			name:   "pbkdf2 sha1",
			fields: fields{conf: DefaultConfig},
			args: args{
				pw:   "qwerty",
				hash: "pbkdf2_sha1$4096$c9Bp0I0FRcXSBmuOPrcD2w$dTCHD12APSrk1gToimJV5Qiz2jactN6vMgDF64tuALg",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "pbkdf2 sha1",
			fields: fields{conf: &HashConfig{
				Iterations: 4096,
				SaltLen:    16,
				KeyLen:     32,
				FuncName:   "sha224",
				Function:   sha256.New224,
			}},
			args: args{
				pw:   "qwerty",
				hash: "pbkdf2_sha224$4096$stUUuY80UAP849MUctylDw$M+feCQ2ddALY5iDbMUJBEGb4sBJz8UObfPye4RHplgg",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "pbkdf2 sha1",
			fields: fields{conf: &HashConfig{
				Iterations: 4096,
				SaltLen:    16,
				KeyLen:     32,
				FuncName:   "sha256",
				Function:   sha256.New,
			}},
			args: args{
				pw:   "qwerty",
				hash: "pbkdf2_sha256$4096$jy6BcRAh36wA20njEWNw6g$pyGrYuJ+bGP2r8DnXkdFZ8hwBuBfQyKF7/OdAQ/dv1U",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "pbkdf2 sha1",
			fields: fields{conf: &HashConfig{
				Iterations: 4096,
				SaltLen:    16,
				KeyLen:     32,
				FuncName:   "sha384",
				Function:   sha512.New384,
			}},
			args: args{
				pw:   "qwerty",
				hash: "pbkdf2_sha384$4096$5sXTC/xGur/zQSW9ORWZ4A$XZ5E39CTlZvgYhQ2iMcM05EbzAQCochfoMpif6Jv6w0",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "pbkdf2 sha1",
			fields: fields{conf: &HashConfig{
				Iterations: 4096,
				SaltLen:    16,
				KeyLen:     32,
				FuncName:   "sha512",
				Function:   sha512.New,
			}},
			args: args{
				pw:   "qwerty",
				hash: "pbkdf2_sha512$4096$QpwrZuuieUAlWZWCGjX9MQ$bKsTrRVmJ5v/p0dV9Fnt+guPRl501SXptZGJCkUDxgY",
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Pbkdf2{
				conf: tt.fields.conf,
			}
			got, err := p.ComparePw(tt.args.pw, tt.args.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("ComparePw() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ComparePw() got = %v, want %v", got, tt.want)
			}
		})
	}
}
