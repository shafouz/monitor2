package crawler

import (
	"bytes"
	"fmt"
	"io"
	"monitor2/src/alerts"
	database "monitor2/src/db"
	models "monitor2/src/db/models"
	"monitor2/utils"
	diff "monitor2/utils"
	"path/filepath"

	"net/http"

	"github.com/rs/zerolog/log"
)

func RunBySchedule(schedule int, db *database.Database) (int, []error) {
	var errors []error

	endpoints, err := db.GetManyEndpointsBySchedule(schedule)
	if err != nil {
		log.Err(err).Caller().Msg("")
		errors = append(errors, err)
		return 0, errors
	}

	for _, endpoint := range endpoints {
		err := RunSingle(&endpoint)
		if err != nil {
			log.Err(err).Caller().Msg("")
			errors = append(errors, err)
		}

		err = db.UpdateEndpointByUrl(endpoint, true)
		if err != nil {
			log.Err(err).Caller().Msg("")
			errors = append(errors, err)
		}
	}

	return len(endpoints), errors
}

func filter_matches(matches [][]byte) [][]byte {
	ret := [][]byte{}
	blocklist := [][]byte{
		[]byte(".png"),
		[]byte(".jpg"),
		[]byte(".jpeg"),
		[]byte(".css"),
		[]byte(".svg"),
		[]byte(".ico"),
		[]byte(".gif"),
		[]byte(".webp"),
		[]byte(".ttf"),
		[]byte(".otf"),
	}

	for _, match := range matches {
		push := true
		for _, block := range blocklist {
			if bytes.HasSuffix(match, block) {
				push = false
			}
		}

		if push {
			ret = append(ret, match)
		}
	}

	return ret
}

func RunSingle(endpoint *models.Endpoint) error {
	var response_body [][]byte
	var err error

	body, status_code, err := crawl(endpoint)
	if err != nil {
		log.Err(err).Caller().Msg("")
		return err
	}

	switch endpoint.Profile {
	case "html":
		path, err := filepath.Abs("src/crawler/scripts/crawl_html.py")
		if err != nil {
			log.Err(err).Caller().Msg("")
			return err
		}
		response_body, err = html_handler(path, body, endpoint.Selector)
		if err != nil {
			log.Err(err).Caller().Msg("")
			return err
		}
	default:
		path, err := filepath.Abs("src/crawler/scripts/crawl_js.py")
		if err != nil {
			log.Err(err).Caller().Msg("")
			return err
		}
		response_body, err = js_handler(path, body)
		if err != nil {
			log.Err(err).Caller().Msg("")
			return err
		}
	}

	if diff := run_diff(response_body, utils.SplitTerminator(endpoint.ResponseBody, "\n"), endpoint.Url); len(diff) > 0 {
		err = alerts.Alert(endpoint.Url, diff, "diff")
		if err != nil {
			log.Err(err).Caller().Msg("")
			return err
		}
	}

	if endpoint.StatusCode != 0 && endpoint.StatusCode != status_code {
		msg := fmt.Sprintf("endpoint: %s\nstatus code has changed: \nprevious: %+v\nnew: %+v\n", endpoint.Url, endpoint.StatusCode, status_code)
		err = alerts.Alert(msg, "", "basic")
		if err != nil {
			log.Err(err).Caller().Msg("")
			return err
		}
	}

	// update endpoint
	endpoint.PreviousResponseBody = endpoint.ResponseBody
	endpoint.ResponseBody = bytes.Join(response_body, []byte("\n"))
	endpoint.StatusCode = status_code

	return nil
}

func run_diff(response_body [][]byte, previous_response_body [][]byte, endpoint string) string {
	t1 := bytes.Join(response_body, []byte("\n"))
	t2 := bytes.Join(previous_response_body, []byte("\n"))
	return string(diff.Diff(endpoint, []byte(t2), endpoint, []byte(t1)))
}

func crawl(endpoint *models.Endpoint) ([]byte, int, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", endpoint.Url, nil)

	if err != nil {
		log.Err(err).Caller().Msg("")
		return []byte{}, 0, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:121.0) Gecko/20100101 Firefox/121.0")
	response, err := client.Do(req)
	if err != nil {
		log.Err(err).Caller().Msg("")
		return []byte{}, 0, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Err(err).Caller().Msg("")
		return []byte{}, 0, err
	}

	log.Info().
		Caller().
		Str("endpoint", endpoint.Url).
		Int("body_length", len(body)).
		Int("status_code", response.StatusCode).
		Msg("")
	return body, response.StatusCode, nil
}

func html_handler(abs_path string, body []byte, extra_args ...string) ([][]byte, error) {
	out, err := utils.RunPyScript(abs_path, body, extra_args)
	if err != nil {
		log.Err(err).Caller().Msg("")
		return nil, err
	}

	return process_crawler_output(out), nil
}

// one line per record format
func js_handler(abs_path string, body []byte, extra_args ...string) ([][]byte, error) {
	out, err := utils.RunPyScript(abs_path, body, extra_args)
	if err != nil {
		log.Err(err).Caller().Msg("")
		return nil, err
	}

	return process_crawler_output(out), nil
}

func process_crawler_output(out []byte) [][]byte {
	split := utils.SplitTerminator(out, "\n")
	filtered := filter_matches(split)
	utils.SortBytes(filtered)
	return utils.CompactBytes(filtered)
}
