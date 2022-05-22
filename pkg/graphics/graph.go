package graphics

import (
	"fmt"
	"golang.org/x/xerrors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"tinkoff-invest-bot/internal/config"

	"go.uber.org/zap"
)

const (
	graphsPath string = "./graphs/"
	detailPath string = "/detail/"
)

type graphHandler struct {
	logger *zap.Logger
}

func NewGraphHandler(logger *zap.Logger) *graphHandler {
	return &graphHandler{
		logger: logger,
	}
}

func (h *graphHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var err error

	if len(r.URL.Path) >= len(detailPath) && r.URL.Path[:len(detailPath)] == detailPath {
		err = h.handleDetail(w, r)
		if err != nil {

		}
	} else {
		err = h.handleMain(w)

	}

	if err != nil {
		h.logger.Error("Error in handling http request", zap.Error(err))
	}
}

func (h *graphHandler) handleDetail(w http.ResponseWriter, r *http.Request) error {
	err := config.CreateDirIfNotExist(graphsPath)
	if err != nil {
		return xerrors.Errorf("Can't create dir for graphics: %w", err)
	}

	path := r.URL.Path[len(detailPath):]
	path = path[:len(path)-len(".html")]
	path = strings.ReplaceAll(path, ".", "_")
	path = strings.ReplaceAll(path, "*", "_")
	path = strings.ReplaceAll(path, "/", "_")

	file, err := os.Open(graphsPath + path + ".html")
	if err != nil {
		_, err = fmt.Fprintln(w, "<body>")
		_, err = fmt.Fprintln(w, "No graph for this request")
		_, err = fmt.Fprintln(w, "</body>")

		return xerrors.Errorf("cant find graphic with path: %w", err)
	}

	b, err := ioutil.ReadAll(file)
	_, err = fmt.Fprintln(w, string(b))

	if err = file.Close(); err != nil {
		return err
	}
	return nil
}

func (h *graphHandler) handleMain(w http.ResponseWriter) error {
	err := config.CreateDirIfNotExist(graphsPath)
	if err != nil {
		return xerrors.Errorf("Can't create dir for graphics: %w", err)
	}

	_, err = fmt.Fprintln(w, "<body>")
	_, err = fmt.Fprintln(w, "Welcome to main page<br/><br/>Please select trading bot config name for monitoring:<br/><br/>")

	files, err := ioutil.ReadDir(graphsPath)
	if err != nil {
		return xerrors.Errorf("Can't read directory with graphics: %w", err)
	}

	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".html") {
			s := fmt.Sprintf("<a href='http://localhost:8080/detail/%s'>%s</a><br/><br/>", f.Name(), f.Name())
			_, err = fmt.Fprintln(w, s)
		}
	}
	_, err = fmt.Fprintln(w, "</body>")
	return err
}
