package httpjsonrpc

import (
	"DNA/common/log"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	tar "github.com/whyrusleeping/tar-utils"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

type ipfsError struct {
	Message string
}

type ipfsResult struct {
	Name string
	Hash string
}
type httpResponseReader struct {
	resp *http.Response
}

//cluster host
var (
	defaultClusterHost = fmt.Sprintf("127.0.0.1:%d", 9094)
	defaultIPFSHost    = fmt.Sprintf("127.0.0.1:%d", 5001)
	defaultHost        = "127.0.0.1"
	defaultTimeout     = 60
	defaultProtocol    = "http"
)

func (r *httpResponseReader) Read(b []byte) (int, error) {
	n, err := r.resp.Body.Read(b)

	// reading on a closed response body is as good as an io.EOF here
	if err != nil && strings.Contains(err.Error(), "read on closed response body") {
		err = io.EOF
	}
	if err == io.EOF {
		_ = r.resp.Body.Close()

		trailerErr := r.checkError()
		if trailerErr != nil {
			return n, trailerErr
		}
	}
	return n, err
}

func (r *httpResponseReader) checkError() error {
	if e := r.resp.Trailer.Get("X-Stream-Error"); e != "" {
		return errors.New(e)
	}
	return nil
}

func AddFileIPFS(filepath string, useCluster bool) (string, error) {

	resp, err := requestIPFS("POST", "add", filepath)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}
	refpath, err := formatIPFSResponse(resp)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}
	if useCluster {
		resp, err = requestCluster("POST", "/pins/"+refpath, nil)
		if err != nil {
			log.Error(err.Error())
			return "", err
		}
		err = formatClusterResponse(resp)
		if err != nil {
			log.Error(err.Error())
			return "", err
		}
	}
	return refpath, err
}

func GetFileIPFS(ref, outPath string) error {
	putpatharg := fmt.Sprintf("output=%s", outPath)
	resp, err := requestIPFS("GET", "get", ref, putpatharg)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	rr := interface{}(&httpResponseReader{resp})
	outReader := rr.(io.Reader)
	extractor := &tar.Extractor{outPath}
	return extractor.Extract(outReader)
}

func formatIPFSResponse(r *http.Response) (string, error) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {

		return "", err
	}

	var ipfsErr ipfsError
	var ipfsRet ipfsResult
	decodeErr := json.Unmarshal(body, &ipfsErr)

	if r.StatusCode != http.StatusOK {
		var msg string
		if decodeErr == nil {
			msg = fmt.Sprintf("IPFS unsuccessful: %d: %s",
				r.StatusCode, ipfsErr.Message)
		} else {
			msg = fmt.Sprintf("IPFS-get unsuccessful: %d: %s",
				r.StatusCode, body)
		}

		return "", errors.New(msg)
	}

	json.Unmarshal(body, &ipfsRet)
	return ipfsRet.Hash, nil

}
func requestIPFS(method, cmd, path string, args ...string) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(defaultTimeout)*time.Second)
	defer cancel()
	url := apiURL(cmd, path, args)

	body := &bytes.Buffer{}
	var writer *multipart.Writer
	if cmd == "add" {
		file, err := os.Open(path)
		if err != nil {

			return nil, err
		}
		defer file.Close()
		writer = multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", path)
		if err != nil {

			return nil, err
		}
		_, err = io.Copy(part, file)

		writer.Close()
	}

	r, err := http.NewRequest(method, url, body)

	if err != nil {
		log.Error(err)
		return nil, err
	}
	if cmd == "add" {
		r.Header.Set("Content-Type", writer.FormDataContentType())
		r.Header.Set("Content-Disposition", "form-data: name=\"files\"")
	}
	r.WithContext(ctx)

	client := &http.Client{}
	resp, err := client.Do(r)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// apiURL is a short-hand for building the url of the IPFS
// daemon API.
func apiURL(cmd, path string, args []string) string {
	if len(args) > 0 {
		var arglist string
		for i := 0; i < len(args); i++ {
			arglist += fmt.Sprintf("&%s", args[i])
		}
		return fmt.Sprintf("http://%s/api/v0/%s?arg=%s&%s", defaultIPFSHost, cmd, path, arglist)
	}
	return fmt.Sprintf("http://%s/api/v0/%s?arg=%s", defaultIPFSHost, cmd, path)
}

func requestCluster(method, path string, body io.Reader, args ...string) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(defaultTimeout)*time.Second)
	defer cancel()

	u := defaultProtocol + "://" + defaultClusterHost + path
	// turn /a/{param0}/{param1} into /a/this/that
	for i, a := range args {
		p := fmt.Sprintf("{param%d}", i)
		u = strings.Replace(u, p, a, 1)
	}
	u = strings.TrimSuffix(u, "/")

	r, err := http.NewRequest(method, u, body)
	if err != nil {

		return nil, err
	}
	r.WithContext(ctx)

	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {

		return nil, err
	}
	return resp, nil
}

func formatClusterResponse(r *http.Response) error {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {

		return err
	}

	var ipfsErr ipfsError

	decodeErr := json.Unmarshal(body, &ipfsErr)

	if r.StatusCode != http.StatusOK && r.StatusCode != http.StatusAccepted {
		var msg string
		if decodeErr == nil {
			msg = fmt.Sprintf("IPFS cluster unsuccessful: %d: %s",
				r.StatusCode, ipfsErr.Message)
		} else {
			msg = fmt.Sprintf("IPFS cluster unsuccessful: %d: %s",
				r.StatusCode, body)
		}

		return errors.New(msg)
	}

	return nil
}
