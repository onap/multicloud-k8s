package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func QueryStatusAPI() string {
	//TODO
	return ""
}

func PushPodInfoToAAI(payload string) error {

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	url := "https://<IP>:30233/aai/v14/cloud-infrastructure/cloud-regions/cloud-region/CloudOwner/RegionOne/tenants/tenant/c673af272d074170881559797f46b89d/vservers/vserver/k8spod"

	fmt.Println("URL:>", url)

	var jsonStr = []byte(`{"vserver-id": "k8spod","vserver-name": "k8spod","vserver-selflink": "example-vserver-selflink-val-57201"}`)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonStr))

	if err != nil {
		return err
	}
	req.Header.Set("X-FromAppId", "SO")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-TransactionId", "get_aai_subscr")
	req.Header.Set("Authorization", "Basic QUFJOkFBSQ==")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println("response Body:", string(body))

	return nil
}

func QueryAAI() {

	for {
		requestPayload := QueryStatusAPI()
		PushPodInfoToAAI(requestPayload)

		time.Sleep(360000 * time.Second)
	}
}

func main() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go QueryAAI()

	fmt.Println("awaiting signal")
	sig := <-c
	fmt.Println(sig)
	fmt.Println("exiting")

}
