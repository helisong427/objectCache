package internal

import (
	"reflect"
	"testing"
)

func TestHashFunc(t *testing.T) {

}



func TestBytes2String(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"1", args{[]byte("aaaaa")}, "aaaaa" },
		{"2", args{[]byte("")}, "" },
		{"3", args{[]byte("b")}, "b" },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Bytes2String(tt.args.b); got != tt.want {
				t.Errorf("Bytes2String() = %v, want %v", got, tt.want)
			}
		})
	}
}


func TestString2Bytes(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{"1", args{"abcdefdg"}, []byte("abcdefdg")},
		{"2", args{""}, []byte("")},
		{"3", args{"a"}, []byte("a")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := String2Bytes(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("String2Bytes() = %v, want %v", got, tt.want)
			}
		})
	}
}