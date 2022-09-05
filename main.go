package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

type status struct {
	Name          string      `json:"name"`
	Slug          string      `json:"slug"`
	Tags          string      `json:"tags"`
	Desc          string      `json:"desc"`
	Grace         int         `json:"grace"`
	NPings        int         `json:"n_pings"`
	Status        string      `json:"status"`
	LastPing      time.Time   `json:"last_ping"`
	NextPing      interface{} `json:"next_ping"`
	ManualResume  bool        `json:"manual_resume"`
	Methods       string      `json:"methods"`
	Subject       string      `json:"subject"`
	SubjectFail   string      `json:"subject_fail"`
	SuccessKw     string      `json:"success_kw"`
	FailureKw     string      `json:"failure_kw"`
	FilterSubject bool        `json:"filter_subject"`
	FilterBody    bool        `json:"filter_body"`
	UniqueKey     string      `json:"unique_key"`
	Timeout       int         `json:"timeout"`
}

func usage() {
	log.Fatalln("Usage: API_KEY=read-only-api-key hc-trigger UID cmd [...]")
}

func main() {
	key := os.Getenv("API_KEY")

	if len(os.Args) < 2 || key == "" {
		usage()
	}

	uid := os.Args[1]
	run := os.Args[2:]

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	for i := 0; i < 5; i++ {
		if i > 0 {
			time.Sleep(100 * time.Millisecond)
		}

		req, _ := http.NewRequest("GET", fmt.Sprintf("https://healthchecks.io/api/v1/checks/%s", uid), nil)
		req.Header.Set("X-Api-Key", key)
		resp, err := client.Do(req)

		if err != nil || resp.StatusCode != 200 {
			log.Println(err)
			continue
		}

		body, err := io.ReadAll(resp.Body)

		if err != nil {
			log.Println(err)
			continue
		}

		s := &status{}
		err = json.Unmarshal(body, s)

		if err != nil {
			log.Println(err)
			continue
		}

		if s.Status != "up" {
			log.Println("Trigger is down, nothing to do here")
			return
		}

		break
	}

	fmt.Printf("Triggering command: %v\n", run)

	args := []string{}

	if len(run) > 1 {
		args = run[1:]
	}

	cmd := exec.Command(run[0], args...)

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
