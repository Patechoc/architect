package trace

import (
	"bytes"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"
)

func NewTracer(sporingsUrl string, context string) *Tracer {
	return &Tracer{
		url:     sporingsUrl,
		context: context,
		enabled: sporingsUrl != "" && context != "",
	}
}

type Tracer struct {
	url     string
	context string
	enabled bool
}

func (t *Tracer) AddImageMetadata(data interface{}) {
	if t.enabled {
		d, err := json.Marshal(data)
		if err != nil {
			logrus.Warnf("Unable to unmarshal image metadata. Got error %s", err)
			return
		}
		t.send(string(d))
	}
}

func (t *Tracer) send(jsonStr string) {
	uri := t.url + "/api/v1/trace/" + t.context

	if t.enabled {
		req, err := http.NewRequest("POST", uri, bytes.NewBuffer([]byte(jsonStr)))
		if err != nil {
			logrus.Warnf("Unable to create request: %s", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			logrus.Warnf("Request failed: %s", err)
			return
		}
		defer resp.Body.Close()
	}
}
