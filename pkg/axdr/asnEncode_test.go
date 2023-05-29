package axdr

import (
	"reflect"
	"testing"
	"time"
)

func TestAsnEncode(t *testing.T) {
	time1 := time.Date(2016, time.April, 1, 10, 0, 0, 0, time.UTC)
	dateTime1 := time.Date(2006, time.January, 2, 0, 0, 0, 0, time.UTC)
	timeTime := time.Date(0, time.January, 1, 15, 4, 5, 0, time.UTC)
	time2 := time.Date(2016, time.April, 1, 10, 10, 0, 0, time.UTC)

	tests := []struct {
		name    string
		v       string
		want    *DlmsData
		wantErr bool
	}{
		{
			name:    "null data",
			v:       "null_data{}",
			want:    CreateAxdrNull(),
			wantErr: false,
		},
		{
			name:    "null array",
			v:       "array{}",
			want:    CreateAxdrArray([]*DlmsData{}),
			wantErr: false,
		},
		{
			name:    "simple array",
			v:       "array{long_unsigned{2}long_unsigned{4}}",
			want:    CreateAxdrArray([]*DlmsData{CreateAxdrLongUnsigned(2), CreateAxdrLongUnsigned(4)}),
			wantErr: false,
		},
		{
			name:    "complex array",
			v:       "array{long_unsigned{2}long_unsigned{4}structure{long_unsigned{2}long_unsigned{4}}}",
			want:    CreateAxdrArray([]*DlmsData{CreateAxdrLongUnsigned(2), CreateAxdrLongUnsigned(4), CreateAxdrStructure([]*DlmsData{CreateAxdrLongUnsigned(2), CreateAxdrLongUnsigned(4)})}),
			wantErr: false,
		},
		{
			name:    "simple structure",
			v:       "structure{long_unsigned{8}octet_string{00 00 01 00 00 ff}integer{2}long_unsigned{0}}",
			want:    CreateAxdrStructure([]*DlmsData{CreateAxdrLongUnsigned(8), CreateAxdrOctetString("00 00 01 00 00 ff"), CreateAxdrInteger(2), CreateAxdrLongUnsigned(0)}),
			wantErr: false,
		},
		{
			name:    "boolean",
			v:       "boolean{true}",
			want:    CreateAxdrBoolean(true),
			wantErr: false,
		},
		{
			name:    "bit_string",
			v:       "bit_string{1010000010}",
			want:    CreateAxdrBitString("1010000010"),
			wantErr: false,
		},
		{
			name:    "double_long",
			v:       "double_long{-123456789}",
			want:    CreateAxdrDoubleLong(-123456789),
			wantErr: false,
		},
		{
			name:    "double_long_unsigned",
			v:       "double_long_unsigned{123456789}",
			want:    CreateAxdrDoubleLongUnsigned(123456789),
			wantErr: false,
		},
		{
			name:    "floating_point",
			v:       "floating_point{4.59}",
			want:    CreateAxdrFloatingPoint(4.59),
			wantErr: false,
		},
		{
			name:    "octet_string",
			v:       "octet_string{test_string}",
			want:    CreateAxdrOctetString("test_string"),
			wantErr: false,
		},
		{
			name:    "visible_string",
			v:       "visible_string{123}",
			want:    CreateAxdrVisibleString("123"),
			wantErr: false,
		},
		{
			name:    "bcd",
			v:       "bcd{25}",
			want:    CreateAxdrBCD(25),
			wantErr: false,
		},
		{
			name:    "integer",
			v:       "integer{2}",
			want:    CreateAxdrInteger(2),
			wantErr: false,
		},
		{
			name:    "long",
			v:       "long{-34}",
			want:    CreateAxdrLong(-34),
			wantErr: false,
		},
		{
			name:    "unsigned",
			v:       "unsigned{2}",
			want:    CreateAxdrUnsigned(2),
			wantErr: false,
		},
		{
			name:    "long_unsigned",
			v:       "long_unsigned{2}",
			want:    CreateAxdrLongUnsigned(2),
			wantErr: false,
		},
		{
			name:    "long64",
			v:       "long64{-1234567890123456789}",
			want:    CreateAxdrLong64(-1234567890123456789),
			wantErr: false,
		},
		{
			name:    "long64_unsigned",
			v:       "long64_unsigned{1234567890123456789}",
			want:    CreateAxdrLong64Unsigned(1234567890123456789),
			wantErr: false,
		},
		{
			name:    "enum",
			v:       "enum{8}",
			want:    CreateAxdrEnum(8),
			wantErr: false,
		},
		{
			name:    "float_32",
			v:       "float_32{1.25}",
			want:    CreateAxdrFloat32(1.25),
			wantErr: false,
		},
		{
			name:    "float_64",
			v:       "float_64{1.23456789}",
			want:    CreateAxdrFloat64(1.23456789),
			wantErr: false,
		},
		{
			name:    "date_time",
			v:       "date_time{2016/04/01 10:00:00}",
			want:    CreateAxdrDateTime(time1),
			wantErr: false,
		},
		{
			name:    "date",
			v:       "date{2006/01/02}",
			want:    CreateAxdrDate(dateTime1),
			wantErr: false,
		},
		{
			name:    "time",
			v:       "time{15:04:05}",
			want:    CreateAxdrTime(timeTime),
			wantErr: false,
		},
		{
			name:    "complex structure",
			v:       "structure{structure{long_unsigned{8}octet_string{00 00 01 00 00 ff}integer{2}long_unsigned{0}}date_time{2016/04/01 10:00:00}date_time{2016/04/01 10:10:00}array{}}",
			want:    CreateAxdrStructure([]*DlmsData{CreateAxdrStructure([]*DlmsData{CreateAxdrLongUnsigned(8), CreateAxdrOctetString("00 00 01 00 00 ff"), CreateAxdrInteger(2), CreateAxdrLongUnsigned(0)}), CreateAxdrDateTime(time1), CreateAxdrDateTime(time2), CreateAxdrArray([]*DlmsData{})}),
			wantErr: false,
		},
		{
			name:    "wrong array",
			v:       "array{long_unsigned{-4}long_unsigned{4}}",
			wantErr: true,
		},

		{
			name:    "wrong structure",
			v:       "structure{long_unsigned{8}octet_string{00 00 01 00 00 ff}integer{eee}long_unsigned{0}}",
			wantErr: true,
		},
		{
			name:    "wrong boolean",
			v:       "boolean{hello}",
			wantErr: true,
		},
		{
			name:    "double_long",
			v:       "double_long{hello}",
			wantErr: true,
		},
		{
			name:    "double_long_unsigned",
			v:       "double_long_unsigned{-123456789}",
			wantErr: true,
		},
		{
			name:    "floating_point",
			v:       "floating_point{hello}",
			wantErr: true,
		},
		{
			name:    "bcd",
			v:       "bcd{325}",
			wantErr: true,
		},
		{
			name:    "integer",
			v:       "integer{hello}",
			wantErr: true,
		},
		{
			name:    "long",
			v:       "long{hello}",
			wantErr: true,
		},
		{
			name:    "unsigned",
			v:       "unsigned{-2}",
			wantErr: true,
		},
		{
			name:    "long_unsigned",
			v:       "long_unsigned{-2}",
			wantErr: true,
		},
		{
			name:    "long64",
			v:       "long64{hello}",
			wantErr: true,
		},
		{
			name:    "long64_unsigned",
			v:       "long64_unsigned{-1234567890123456789}",
			wantErr: true,
		},
		{
			name:    "enum",
			v:       "enum{a}",
			wantErr: true,
		},
		{
			name:    "float_32",
			v:       "float_32{hello}",
			wantErr: true,
		},
		{
			name:    "float_64",
			v:       "float_64{hello}}",
			wantErr: true,
		},
		{
			name:    "date_time",
			v:       "date_time{2016/13/01 10:00:00}",
			wantErr: true,
		},
		{
			name:    "date",
			v:       "date{2006/01/32}",
			wantErr: true,
		},
		{
			name:    "time",
			v:       "time{15:60:05}",
			wantErr: true,
		},
		{
			name:    "no exist",
			v:       "no_exist{}",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := AsnEncode(tt.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("AsnEncode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AsnEncode() = %v, want %v", got, tt.want)
			}
		})
	}
}
