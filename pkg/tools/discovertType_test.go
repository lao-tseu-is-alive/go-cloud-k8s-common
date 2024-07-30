package tools

import (
	"testing"
	"time"
)

func TestGetOpenApiType(t *testing.T) {
	type args struct {
		t any
	}
	var emptyStringPointer *string
	var stringValue = "toto is cool"
	stringPointer := &stringValue
	var integerValue = 678
	intPointer := &integerValue
	var floatValue = 3.1415
	floatPointer := &floatValue
	type Person struct {
		FirstName string    `json:"firstname"`
		LastName  string    `json:"lastname"`
		BirthDate time.Time `json:"birthdate"`
	}
	bob := Person{
		FirstName: "Bob",
		LastName:  "BLAIR",
		BirthDate: time.Now(),
	}
	personPointer := &bob
	tests := []struct {
		name    string
		args    args
		wantRes string
		wantErr error
	}{
		{name: "it should return string for a empty string",
			args:    args{t: ""},
			wantRes: "string",
			wantErr: nil,
		},
		{name: "it should return string nullable for a uninitialized pointer to string",
			args:    args{t: emptyStringPointer},
			wantRes: "string, nullable",
			wantErr: nil,
		},
		{name: "it should return string nullable for a pointer to a string value",
			args:    args{t: stringPointer},
			wantRes: "string, nullable",
			wantErr: nil,
		},
		{name: "it should return integer nullable for a pointer to a integer value",
			args:    args{t: intPointer},
			wantRes: "integer, nullable",
			wantErr: nil,
		},
		{name: "it should return number nullable for a pointer to a float value",
			args:    args{t: floatPointer},
			wantRes: "number, nullable",
			wantErr: nil,
		},
		{name: "it should return object nullable for a pointer to a struct value",
			args:    args{t: personPointer},
			wantRes: "object, nullable",
			wantErr: nil,
		},
		{name: "it should return string for a string with only space",
			args:    args{t: "  "},
			wantRes: "string",
			wantErr: nil,
		},
		{name: "it should return integer for a zero",
			args:    args{t: 0},
			wantRes: "integer",
			wantErr: nil,
		},
		{name: "it should return int for a negative number",
			args:    args{t: -10},
			wantRes: "integer",
			wantErr: nil,
		},
		{name: "it should return number for a float",
			args:    args{t: floatValue},
			wantRes: "number",
			wantErr: nil,
		},
		{name: "it should return object for a struct",
			args:    args{t: bob},
			wantRes: "object",
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, gotErr := GetOpenApiType(tt.args.t)
			// fmt.Printf("##GetType(%v)[%T] returns: (%v, error: %v), wants: (%v, error:%v)", tt.args.t, tt.args.t, gotRes, gotErr, tt.wantRes, tt.wantErr)
			if gotRes != tt.wantRes || gotErr != tt.wantErr {
				t.Errorf("GetType(%v)[%T] got: (%v, error: %v), wants: (%v, error:%v)",
					tt.args.t, tt.args.t, gotRes, gotErr, tt.wantRes, tt.wantErr)
			}
		})
	}
}
