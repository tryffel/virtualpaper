package models

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProperty_ValidateValue(t *testing.T) {
	type fields struct {
		Id        int
		User      int
		Name      string
		Type      PropertyType
		Global    bool
		Unique    bool
		Exclusive bool
		Counter   int
		Offset    int
		Prefix    string
		Mode      string
		Readonly  bool
		DateFmt   string
		Timestamp Timestamp
	}
	type args struct {
		val string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "bad json",
			fields: fields{
				Id:        0,
				User:      0,
				Name:      "",
				Type:      JsonProperty,
				Global:    false,
				Unique:    false,
				Exclusive: false,
				Counter:   0,
				Offset:    0,
				Prefix:    "",
				Mode:      "",
				Readonly:  false,
				DateFmt:   "",
				Timestamp: Timestamp{},
			},
			args:    args{val: "{a}"},
			wantErr: assert.Error,
		},
		{
			name: "valid json",
			fields: fields{
				Id:        0,
				User:      0,
				Name:      "",
				Type:      JsonProperty,
				Global:    false,
				Unique:    false,
				Exclusive: false,
				Counter:   0,
				Offset:    0,
				Prefix:    "",
				Mode:      "",
				Readonly:  false,
				DateFmt:   "",
				Timestamp: Timestamp{},
			},
			args:    args{val: `{"a": 1}`},
			wantErr: assert.NoError,
		},
		{
			name: "valid json array",
			fields: fields{
				Id:        0,
				User:      0,
				Name:      "",
				Type:      JsonProperty,
				Global:    false,
				Unique:    false,
				Exclusive: false,
				Counter:   0,
				Offset:    0,
				Prefix:    "",
				Mode:      "",
				Readonly:  false,
				DateFmt:   "",
				Timestamp: Timestamp{},
			},
			args:    args{val: `[1,2]`},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Property{
				Id:        tt.fields.Id,
				User:      tt.fields.User,
				Name:      tt.fields.Name,
				Type:      tt.fields.Type,
				Global:    tt.fields.Global,
				Unique:    tt.fields.Unique,
				Exclusive: tt.fields.Exclusive,
				Counter:   tt.fields.Counter,
				Prefix:    tt.fields.Prefix,
				Mode:      tt.fields.Mode,
				Readonly:  tt.fields.Readonly,
				DateFmt:   tt.fields.DateFmt,
				Timestamp: tt.fields.Timestamp,
			}
			tt.wantErr(t, p.ValidateValue(tt.args.val), fmt.Sprintf("ValidateValue(%v)", tt.args.val))
		})
	}
}
