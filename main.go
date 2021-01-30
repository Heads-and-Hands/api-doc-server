package main

import (
	"fmt"
	"log"
	"net/http"
	"io/ioutil"
	"os"
	"regexp"
	"encoding/base64"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"errors"
	"crypto/sha1"
	"crypto/subtle"
	"bytes"
	"api-doc-server/mymd5"
)

type compareFunc func(hashedPassword, password []byte) error

var (
	errMismatchedHashAndPassword = errors.New("mismatched hash and password")

	compareFuncs = []struct {
		prefix  string
		compare compareFunc
	}{
		{"", compareMD5HashAndPassword}, // default compareFunc
		{"{SHA}", compareShaHashAndPassword},
		// Bcrypt is complicated. According to crypt(3) from
		// crypt_blowfish version 1.3 (fetched from
		// http://www.openwall.com/crypt/crypt_blowfish-1.3.tar.gz), there
		// are three different has prefixes: "$2a$", used by versions up
		// to 1.0.4, and "$2x$" and "$2y$", used in all later
		// versions. "$2a$" has a known bug, "$2x$" was added as a
		// migration path for systems with "$2a$" prefix and still has a
		// bug, and only "$2y$" should be used by modern systems. The bug
		// has something to do with handling of 8-bit characters. Since
		// both "$2a$" and "$2x$" are deprecated, we are handling them the
		// same way as "$2y$", which will yield correct results for 7-bit
		// character passwords, but is wrong for 8-bit character
		// passwords. You have to upgrade to "$2y$" if you want sant 8-bit
		// character password support with bcrypt. To add to the mess,
		// OpenBSD 5.5. introduced "$2b$" prefix, which behaves exactly
		// like "$2y$" according to the same source.
		{"$2a$", bcrypt.CompareHashAndPassword},
		{"$2b$", bcrypt.CompareHashAndPassword},
		{"$2x$", bcrypt.CompareHashAndPassword},
		{"$2y$", bcrypt.CompareHashAndPassword},
	}
)

func CheckSecret(password, secret string) bool {
	compare := compareFuncs[0].compare
	for _, cmp := range compareFuncs[1:] {
		if strings.HasPrefix(secret, cmp.prefix) {
			compare = cmp.compare
			break
		}
	}
	return compare([]byte(secret), []byte(password)) == nil
}

func compareShaHashAndPassword(hashedPassword, password []byte) error {
	d := sha1.New()
	d.Write(password)
	if subtle.ConstantTimeCompare(hashedPassword[5:], []byte(base64.StdEncoding.EncodeToString(d.Sum(nil)))) != 1 {
		return errMismatchedHashAndPassword
	}
	return nil
}

func compareMD5HashAndPassword(hashedPassword, password []byte) error {
	parts := bytes.SplitN(hashedPassword, []byte("$"), 4)
	if len(parts) != 4 {
		return errMismatchedHashAndPassword
	}
	magic := []byte("$" + string(parts[1]) + "$")
	salt := parts[2]

	if subtle.ConstantTimeCompare(hashedPassword, mymd5.MD5Crypt(password, salt, magic)) != 1 {
		return errMismatchedHashAndPassword
	}
	return nil
}

func getRoot() string {
	path := os.Getenv("ROOT_PATH")
	if path == "" {
		path = "/var/www/"
	}
	return path
}

func main() {
	http.HandleFunc("/", handler)
	log.Println("Run and listen 8383")
	http.ListenAndServe(":8383", nil)
}

func handler(iWrt http.ResponseWriter, iReq *http.Request) {
	var lGet = iReq.URL.Path[1:]
	if lGet == "" || lGet == "/" {
		lGet = "index.html"
	}

	exp := "(.*):[:digit:]*"
	r := regexp.MustCompile(exp)
	host := r.FindString(iReq.Host)
	if host != ""{
		host = host[:len(host) - 1]
	} else {
		host = iReq.Host
	}
	log.Println("Host: " + host)

	path := ""
	if host == ""  || host == "localhost" {
		path = getRoot()
	} else {
		path = getRoot() + host + "/"
	}

	if !basicAuth(iWrt, iReq, path) {
		iWrt.Header().Set("WWW-Authenticate", `Basic realm="Beware! Protected REALM! "`)
		iWrt.WriteHeader(401)
		iWrt.Write([]byte("401 Unauthorized\n"))
		return
	}

	lData := readFile(path + lGet)
	fmt.Fprintln(iWrt, lData)
}

func basicAuth(w http.ResponseWriter, r *http.Request, path string) bool {
	passFile := readFile(path + "htpasswd")
	if passFile == "404" { return true }
	realPair := strings.SplitN(string(passFile), ":", 2)
	if len(realPair) != 2 {	return true }

	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 { return false }
	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {	return false }
	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {	return false }

	return pair[0] == realPair[0] && CheckSecret(pair[1], realPair[1])
}

func readFile(iFileName string) string {
	log.Println("readFile: " + iFileName)
	lData, err := ioutil.ReadFile(iFileName)
	var lOut string
	if !os.IsNotExist(err) {
		lOut = string(lData)
	} else {
		lOut = "404"
	}
	return lOut
}