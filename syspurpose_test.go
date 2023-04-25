package main

import (
	"os"
	"reflect"
	"syscall"
	"testing"
)

// Correct syspurpose.json content
var correctSyspurposeJSONContent = `{
	"role": "FooRole",
	"service_level_agreement": "FooSLA",
	"usage": "FooUsage"
}`

// Correct syspurpose.json content with addons
var correctSyspurposeJSONContentWithAddons = `{
	"role": "FooRole",
	"service_level_agreement": "FooSLA",
	"usage": "FooUsage",
	"addons": ["addon1", "addon2"]
}`

// Correct syspurpose.json content with empty content
var emptySyspurposeJSONContent = `{}`

// Incorrect syspurpose.json content
var incorrectSyspurposeJSONContent = `{
	"role": "FooRole,
	"service_level_agreement":,
	usage: "FooUsage"
}`

// createTemporarySystemPurposeFile tries to create temporary syspurpose file
func createTemporarySystemPurposeFile(t *testing.T, syspurposeJsonContent *string) *os.File {
	tempDir := t.TempDir()
	tempFile, err := os.CreateTemp(tempDir, "syspurpose.json")
	if err != nil {
		panic(err)
	}
	_, err = tempFile.Write([]byte(*syspurposeJsonContent))
	if err != nil {
		panic(err)
	}
	return tempFile
}

func Test_getSystemPurpose(t *testing.T) {
	pathToNonExistingFile := "/path/to/non/existing/file"
	// Create table with test cases
	type args struct {
		syspurposeJSONFilePath *string
		syspurposeJSONContent  *string
	}
	tests := []struct {
		name    string
		args    args
		want    *SysPurposeJSON
		wantErr bool
	}{
		{
			name: "successful reading of syspurpose values",
			args: args{
				nil,
				&correctSyspurposeJSONContent,
			},
			want: &SysPurposeJSON{
				Role:                  "FooRole",
				ServiceLevelAgreement: "FooSLA",
				Usage:                 "FooUsage",
			},
			wantErr: false,
		},
		{
			name: "successful reading of syspurpose values with addons",
			args: args{
				nil,
				&correctSyspurposeJSONContentWithAddons,
			},
			want: &SysPurposeJSON{
				Role:                  "FooRole",
				ServiceLevelAgreement: "FooSLA",
				Usage:                 "FooUsage",
			},
			wantErr: false,
		},
		{
			name: "reading empty syspurpose values",
			args: args{
				nil,
				&emptySyspurposeJSONContent,
			},
			want: &SysPurposeJSON{
				Role:                  "",
				ServiceLevelAgreement: "",
				Usage:                 "",
			},
			wantErr: false,
		},
		{
			name: "error reading wrong syspurpose values",
			args: args{
				nil,
				&incorrectSyspurposeJSONContent,
			},
			want: &SysPurposeJSON{
				Role:                  "",
				ServiceLevelAgreement: "",
				Usage:                 "",
			},
			wantErr: true,
		},
		{
			name: "reading of non-existing syspurpose file",
			args: args{
				&pathToNonExistingFile,
				nil,
			},
			want: &SysPurposeJSON{
				Role:                  "",
				ServiceLevelAgreement: "",
				Usage:                 "",
			},
			wantErr: true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Crete temporary syspurpose file
			var syspurposeFileName string
			if tt.args.syspurposeJSONContent != nil {
				tempFile := createTemporarySystemPurposeFile(t, tt.args.syspurposeJSONContent)
				syspurposeFileName = tempFile.Name()
				defer func(path string) {
					err := syscall.Unlink(path)
					if err != nil {
						panic(err)
					}
				}(tempFile.Name())
			} else {
				syspurposeFileName = *tt.args.syspurposeJSONFilePath
			}

			// Run test
			got, err := getSystemPurpose(&syspurposeFileName)

			// Check results
			if (err != nil) != tt.wantErr {
				t.Errorf("getSystemPurpose() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSystemPurpose() got = %v, want %v", got, tt.want)
			}
		})
	}
}
