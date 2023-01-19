package iris_extend_helper

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/mvc"
)

func RegisterMacros(app *iris.Application) {
	uuidRegex := regexp.MustCompile(`[0-9a-f\-]{32,36}`)
	app.Macros().Get("string").RegisterFunc("uuid", uuidRegex.MatchString)
}

func RegisterParties(app *iris.Application, parties iris.Map) {
	for path, controller := range parties {
		mvc.New(app.Party(path)).Handle(controller)
	}
}

func RegisterErrorHandlers(app *iris.Application, codes iris.Map) {
	for code, handler := range codes {
		code, err := strconv.Atoi(code)
		if err != nil {
			log.Fatalln(err)
		}

		handler := handler.(func(iris.Context))
		app.Any(fmt.Sprintf("/%v", code), handler)
		app.OnErrorCode(code, handler)
	}
}

func ServeStaticFiles(app *iris.Application, path string) {
	if strings.HasPrefix(path, "/") {
		app.HandleDir(path, "./app/public")
	}
}

func RouteResources(route context.RouteReadOnly, resource string) []string {
	resources := []string{"*", "/*", resource}
	staticPath := route.StaticPath()
	path := route.Path()
	if staticPath != resource {
		if strings.HasSuffix(staticPath, "/") {
			staticPath += "*"
		} else {
			staticPath += "/*"
		}
		resources = append(resources, staticPath)
	}
	if path != resource {
		pathRegex := regexp.MustCompile(`\{[^\}\/]+\}`)
		path = pathRegex.ReplaceAllString(resource, "*")
		if path != staticPath {
			resources = append(resources, path)
		}
	}
	return resources
}
