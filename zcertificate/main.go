/*
 * ZGrab Copyright 2015 Regents of the University of Michigan
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not
 * use this file except in compliance with the License. You may obtain a copy
 * of the License at http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
 * implied. See the License for the specific language governing
 * permissions and limitations under the License.
 */

package main

import (
	"bytes"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"os"

	"github.com/zmap/zgrab/ztools/x509"
	"github.com/zmap/zlint/lints"
	"github.com/zmap/zlint/zlint"
)

func exitErr(a ...interface{}) {
	fmt.Fprint(os.Stderr, "FATAL: ")
	fmt.Fprintln(os.Stderr, a...)
	os.Exit(1)
}

func main() {

	if len(os.Args) != 2 {
		exitErr("No path to certificate provided")
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		exitErr("Could not open specified certificate:", err)
	}
	buf := bytes.NewBuffer(nil)
	io.Copy(buf, f)

	p, _ := pem.Decode(buf.Bytes())
	if p == nil {
		exitErr("Unable to parse PEM file: ", err)
	}
	x509Cert, err := x509.ParseCertificate(p.Bytes)
	if err != nil {
		exitErr("Unable to parse certificate: ", err)
	}

	out, err := json.Marshal(x509Cert)
	if err != nil {
		exitErr("Unable to convert certificate to JSON: ", err)
	}
	fmt.Println(string(out))

	zlintReport, err := runZlint(x509Cert)
	if err != nil {
		exitErr("Unable to run zlint: ", err)
	}

	zlintAppendedToCert, err := appendZlintToCertificate(x509Cert, zlintReport)
	if err != nil {
		exitErr("Unable to Marshal JSON: ", err)
	}

	fmt.Println(string(zlintAppendedToCert))

}

func appendZlintToCertificate(x509Cert *x509.Certificate, lintResult map[string]lints.ResultStruct) ([]byte, error) {
	type intermediateCert x509.Certificate

	return json.Marshal(struct {
		intermediateCert
		Zlint map[string]lints.ResultStruct
	}{
		intermediateCert: intermediateCert(*x509Cert),
		Zlint:            lintResult,
	})
}

func runZlint(x509Cert *x509.Certificate) (map[string]lints.ResultStruct, error) {
	m := make(map[string]int)
	zlintReport, err := zlint.ParsedTestHandler(x509Cert, m)
	if err != nil {
		return nil, err
	}
	return zlintReport, nil
}
