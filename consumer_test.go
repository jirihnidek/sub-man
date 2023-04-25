package main

import (
	"os"
	"reflect"
	"syscall"
	"testing"
)

// Content of cert file without any block
var wrongConsumerCertFileContent = `Foo`

// Wrong consumer cert file content with wrong block
// with base encoded "Foo" string
var wrongBlockConsumerCertFileContent = `-----BEGIN FOO-----
Rm9vCg==
-----END FOO-----
`

// Wrong consumer cert file content with wrong block
// with base encoded "Foo" string
var corruptedConsumerCertFileContent = `-----BEGIN CERTIFICATE-----
Rm9vCg==
-----END CERTIFICATE-----
`

// Correct consumer cert file content
var consumerCertFileContent = `-----BEGIN CERTIFICATE-----
MIIF9TCCA92gAwIBAgIIMlHEzBduyXgwDQYJKoZIhvcNAQELBQAwOzEaMBgGA1UE
AwwRY2VudG9zOC1jYW5kbGVwaW4xCzAJBgNVBAYTAlVTMRAwDgYDVQQHDAdSYWxl
aWdoMB4XDTIzMDQyMDA3MzUzMVoXDTI4MDQyMDA4MzUzMVowRDETMBEGA1UECgwK
ZG9uYWxkZHVjazEtMCsGA1UEAwwkNTA5ZDU2NzItZGRhMi00ZDQ0LThlN2QtNjJk
ZmJmN2FiOGI4MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA1cv1rofj
6PXvGU3iqpLRBWLrq6/iGfnMLH+S8DX1x1osu8FF2IyBGspxTs6V4UbZFNwhzegq
j3g0y3vUugoTCvm4YS0OiOPQuy01F10bFbngouN8WCbGtyS1BMHKT0GxDz4QQgQr
Fyo+E5VpEYlMZH1TYSkGEMqxN8IJhFLQrc6sRINnFgBkv8jTL5yNMC7l9pY8qvaE
1hGbhpQqol4tv/hu4sqBXudIEE66MxjsNLLL2rSqKX80uQ3hzVFtHIPdXweXZBfu
bFDSPRb1LdOCOwFxY4+bHaqGioW2hNZNEPJnVTmrsNd5dg0XbX4cWyq4MKKk4isA
YfcB7fm2s6hxxN6jEm+l0DU9sSBpATAJxjKuNau9v3ckGCspUWLOV3ldG4L+jiLk
Q6Iuz8MxTIxkj2sKlkgYCwzK6A8DGWyOMJt3IdDJB2UeVsi3k8pff11MKkSbVTxE
1Lqfzu6rGWcwSaanoIgk2flOowYw7JlRe6QpuWtYwRuNuZgonM53Kbj54Tqo77sI
R/Z9Ky6cw8BByAVDsBwTlFMS4G1YebHjUd9tnnz25sfLQbVuKDtdmNllzejeNe2G
cZLoXH/7EKXUy7FeWVOU8R1BTk8CObWDTewZLPJYbUkPS5zndzJs/DbFOr/3sHDh
ZhfcpLfLYdrOU1QVc7ZKEtztcbgJLMYBH40CAwEAAaOB8zCB8DAOBgNVHQ8BAf8E
BAMCBLAwEwYDVR0lBAwwCgYIKwYBBQUHAwIwCQYDVR0TBAIwADARBglghkgBhvhC
AQEEBAMCBaAwHQYDVR0OBBYEFJa8Z6NLWUnBHWgSiHNhYmRfFU1RMB8GA1UdIwQY
MBaAFJE2hokZj5VQLw9nF1KgylvEX5akMGsGA1UdEQRkMGKkRjBEMRMwEQYDVQQK
DApkb25hbGRkdWNrMS0wKwYDVQQDDCQ1MDlkNTY3Mi1kZGEyLTRkNDQtOGU3ZC02
MmRmYmY3YWI4YjikGDAWMRQwEgYDVQQDDAt0aGlua3BhZC1wMTANBgkqhkiG9w0B
AQsFAAOCAgEAU6kUP/MyORoGuzz9xb6YT7aSCq0csnIl8QGlKONUN1W9B8jJK59+
1C7Gba0pO28KB5FUT4f9E3kDSqJkKdVM+fzmHtesX6nl1SMcB0WEEQNkrQBfK+vn
pZgok8GuCibspdg4O9WgKmH1ab/0JHhzCT/Qg4vkrnXSBSL1rNauxBxpfOy4gYhi
SoXVYVfBo5DAppuvcmojWAIMTucOR0rpWTacRY4mNj3ScAmGzk/pbYDBny5NAjV4
qUuR13IQyv6tfDDAuGpF7ZQZkTS8BJ7CTPTTJMb1Dl+448kXF2HZAupIqeibGfSI
EvZNweV3b4KWf1XAvdGnCXnuL0tRab+PTzT9MMSiWxqPDaTHRIfbuaL0yh6CrOak
ei/m8eJWO2P0GtY12rgbNP1hItVSVeNCZXEHcJf3GzHl54PSulxd08Lb0ZqnyrCN
Txj0yBFRxnsHjmMcvmZJc2k+ciIZt8tyVzlIDsB8qi5v0eePlJPbMNFQrpFLCoLY
N/yzFdt4VNbgJ0h0nPwnhm9U+lSp5/2CHlOPAQIRWR7uxppBBrIKKNsVemIMGrx/
1RgW+gcnmcwCkDSZTiy6owoACbGD60kbq24TfybRhXMnMwJZUQYirwgJJ/giCLMU
xTjCbBbRMQaHhiSHg49/VH7yGuzD3l+mgrjBhjo8AemJzWGYvNgazpQ=
-----END CERTIFICATE-----`

var expectedConsumerUUID = "509d5672-dda2-4d44-8e7d-62dfbf7ab8b8"

// createTemporaryConsumerCert tries to create a temporary file with the content of
// the consumer cert
func createTemporaryConsumerCert(t *testing.T, consumerCertFileContent *string) (*os.File, error) {
	tempDir := t.TempDir()
	tempFile, err := os.CreateTemp(tempDir, "cert.pem")
	if err != nil {
		return nil, err
	}
	defer func(tempFile *os.File) {
		err := tempFile.Close()
		if err != nil {
			panic(err)
		}
	}(tempFile)

	if _, err := tempFile.Write([]byte(*consumerCertFileContent)); err != nil {
		return nil, err
	}

	return tempFile, nil
}

// Test_getConsumerUUID tests the function getConsumerUUID
func Test_getConsumerUUID(t *testing.T) {
	var pathToNonExistingFile = "/path/to/non/existing/file"
	// Create table with test cases
	type args struct {
		consumerCertFileName *string
		consumerCertContent  *string
	}
	tests := []struct {
		name    string
		args    args
		want    *string
		wantErr bool
	}{
		{
			name: "successful reading of consumer uuid",
			args: args{
				consumerCertFileName: nil,
				consumerCertContent:  &consumerCertFileContent,
			},
			want:    &expectedConsumerUUID,
			wantErr: false,
		},
		{
			name: "error reading wrong consumer certificate",
			args: args{
				consumerCertFileName: nil,
				consumerCertContent:  &wrongConsumerCertFileContent,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error reading consumer certificate with wrong block",
			args: args{
				consumerCertFileName: nil,
				consumerCertContent:  &wrongBlockConsumerCertFileContent,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error reading corrupted consumer certificate",
			args: args{
				consumerCertFileName: nil,
				consumerCertContent:  &corruptedConsumerCertFileContent,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error reading non-existing consumer cert",
			args: args{
				consumerCertFileName: &pathToNonExistingFile,
				consumerCertContent:  nil,
			},
			want:    nil,
			wantErr: true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var consumerCertFileName string
			// Create temporary consumer cert file
			if tt.args.consumerCertContent != nil {
				tempConsumerCertFile, err := createTemporaryConsumerCert(t, tt.args.consumerCertContent)
				if err != nil {
					t.Errorf("could not create temporary consumer cert file: %v", err)
				}
				consumerCertFileName = tempConsumerCertFile.Name()
				// Delete temporary consumer cert file
				defer func(path string) {
					err := syscall.Unlink(path)
					if err != nil {
						panic(err)
					}
				}(consumerCertFileName)
			} else {
				// Try to use non-existing file
				consumerCertFileName = *tt.args.consumerCertFileName
			}

			// Run test
			got, err := getConsumerUUID(&consumerCertFileName)

			// Check result
			if (err != nil) != tt.wantErr {
				t.Errorf("getConsumerUUID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getConsumerUUID() got = %v, want %v", got, tt.want)
			}
		})
	}
}
