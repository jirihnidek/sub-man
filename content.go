package main

import (
	"bytes"
	"compress/zlib"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"time"
)

type YumRepo struct {
	Name string
	Id   string
	URL  string
}

type EntitlementContent struct {
	Consumer     string `json:"consumer"`
	Subscription struct {
		Sku  string `json:"sku"`
		Name string `json:"name"`
	} `json:"subscription"`
	Order struct {
		Start time.Time `json:"start"`
		End   time.Time `json:"end"`
	} `json:"order"`
	Products []struct {
		Id            string        `json:"id"`
		Name          string        `json:"name"`
		Version       string        `json:"version"`
		Architectures []interface{} `json:"architectures"`
		Content       []struct {
			Id             string   `json:"id"`
			Type           string   `json:"type"`
			Name           string   `json:"name"`
			Label          string   `json:"label"`
			Vendor         string   `json:"vendor"`
			Path           string   `json:"path"`
			Enabled        bool     `json:"enabled,omitempty"`
			Arches         []string `json:"arches"`
			GpgUrl         string   `json:"gpg_url,omitempty"`
			MetadataExpire int      `json:"metadata_expire,omitempty"`
			RequiredTags   []string `json:"required_tags,omitempty"`
		} `json:"content"`
	} `json:"products"`
	Pool struct {
	} `json:"pool"`
}

func generateContentFromEntCert(entCert *string) (*[]YumRepo, error) {
	data := []byte(*entCert)
	for data != nil {
		block, rest := pem.Decode(data)
		if block == nil {
			break
		} else {
			fmt.Printf("Block type: %s\n", block.Type)
			if block.Type == "CERTIFICATE" {
				certificate, err := x509.ParseCertificate(block.Bytes)
				if err == nil {
					fmt.Printf("\tCertificate subject: %s\n", certificate.Subject)
				}
			}
			if block.Type == "ENTITLEMENT DATA" {
				// The "block.Bytes" is already base64 decoded. We can try to un-compress.
				b := bytes.NewReader(block.Bytes)
				zReader, err := zlib.NewReader(b)
				if err != nil {
					return nil, fmt.Errorf("unable to create new zlib readed for ENTITLEMENT DATA: %s", err)
				}
				p, err := ioutil.ReadAll(zReader)
				if err != nil {
					return nil, fmt.Errorf("unable to uncompress ENTITLEMENT DATA: %s", err)
				}
				fmt.Printf("Uncompressed data: %s\n", string(p))
				_ = zReader.Close()
			}
		}
		data = rest
	}
	return nil, nil
}
