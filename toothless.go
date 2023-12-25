package toothless

import (
	"fmt"
	"github.com/CloudyKit/jet/v6"
	"github.com/carlmjohnson/versioninfo"
	"github.com/getDragon-dev/toothless/render"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Toothless struct {
	AppName  string
	Debug    bool
	Version  string
	ErrorLog *log.Logger
	InfoLog  *log.Logger
	RootPath string
	Routes   *chi.Mux
	Render   *render.Render
	JetViews *jet.Set
	config   config
}

type config struct {
	port     string
	renderer string
}

func (t *Toothless) New(rootPath string) error {
	pathConfig := initPaths{
		rootPath:    rootPath,
		folderNames: []string{"handlers", "migrations", "views", "data", "public", "tmp", "logs", "middleware"},
	}

	err := t.Init(pathConfig)
	if err != nil {
		return err
	}

	err = t.checkDotEnv(rootPath)
	if err != nil {
		return err
	}

	err = godotenv.Load(rootPath + "/.env")
	if err != nil {
		return err
	}

	infoLog, errorLog := t.startLoggers()
	t.InfoLog = infoLog
	t.ErrorLog = errorLog
	t.Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	t.Version = versioninfo.Short()
	t.RootPath = rootPath
	t.Routes = t.routes().(*chi.Mux)

	t.config = config{
		port:     os.Getenv("PORT"),
		renderer: os.Getenv("RENDERER"),
	}

	var views = jet.NewSet(
		jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
		jet.InDevelopmentMode(),
	)

	t.JetViews = views
	t.createRenderer()

	return nil
}

func (t *Toothless) Init(p initPaths) error {
	root := p.rootPath
	for _, path := range p.folderNames {
		err := t.CreateDirIfNotExist(root + "/" + path)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Toothless) ListenAndServe() {
	srv := http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("PORT")),
		ErrorLog:     t.ErrorLog,
		Handler:      t.Routes,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second,
	}

	t.InfoLog.Printf("Listening on http://127.0.0.1:%s", os.Getenv("PORT"))
	err := srv.ListenAndServe()
	t.ErrorLog.Fatal(err)
}
func (t *Toothless) checkDotEnv(path string) error {
	err := t.CreateFileIfNotExists(fmt.Sprintf("%s/.env", path))
	if err != nil {
		return err
	}
	return nil
}

func (t *Toothless) startLoggers() (*log.Logger, *log.Logger) {
	var infoLog *log.Logger
	var errorLog *log.Logger

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	return infoLog, errorLog
}

func (t *Toothless) createRenderer() {
	myRenderer := render.Render{
		Renderer: t.config.renderer,
		RootPath: t.RootPath,
		Port:     t.config.port,
		JetViews: t.JetViews,
	}

	t.Render = &myRenderer
}
