package crawler_test

import (
	"monitor2/src/crawler"
	"path/filepath"
	"testing"

  "log"
)

func printByteArray(arr [][]byte){
  for i, el := range arr {
    log.Printf("%+v: %+v", i, string(el))
  }
}

func TestHtmlHandler(t *testing.T){
  body := []byte(`<h1>asdasjdh</h1>`)
  path, err := filepath.Abs("./scripts/crawl_html.py")
  if err != nil { t.Fatal(err) }

  res, err := crawler.HtmlHandler(path, body, "h1")
  if err != nil { t.Fatal(err) }
  if len(res) != 1 {
    t.Fatal()
  }
}

func TestJsHandler(t *testing.T){
  body := []byte(`<script src="asdklasjdklas.com"></script>`)
  path, err := filepath.Abs("./scripts/crawl_js.py")
  if err != nil { t.Fatal(err) }

  res, err := crawler.JsHandler(path, body)
  if err != nil { t.Fatal(err) }
  if len(res) != 1 {
    t.Fatal()
  }
}

func TestJsHandlerRemovesDupes(t *testing.T){
  body := []byte(`
    <script src="asdklasjdklas.com"></script>
    <script src="asdklasjdklas.com"></script>
    <script src="rsdklasjdklas.com"></script>
  `)
  path, err := filepath.Abs("./scripts/crawl_js.py")
  if err != nil { t.Fatal(err) }

  res, err := crawler.JsHandler(path, body)
  if err != nil { t.Fatal(err) }
  if len(res) != 2 {
    t.Fatal()
  }
}

func TestFilterMatches(t *testing.T){
	files := [][]byte{
		[]byte("jasdlfkj.png"),
		[]byte("asd.jpg"),
		[]byte("a.sd.js"),
		[]byte("A.sd.js"),
	}

  files = crawler.FilterMatches(files)
  if len(files) != 2 {
    t.Fatal()
  }
}
