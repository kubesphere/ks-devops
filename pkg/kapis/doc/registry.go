package doc

import (
	"github.com/emicklei/go-restful"
	swagger "github.com/emicklei/go-restful-swagger12"
)

// AddSwaggerService adds the Swagger service
func AddSwaggerService(wss []*restful.WebService, c *restful.Container) {
	config := swagger.Config{
		WebServices:     wss,
		ApiPath:         "/apidocs.json",
		SwaggerPath:     "/apidocs/",
		SwaggerFilePath: "bin/swagger-ui"}
	swagger.RegisterSwaggerService(config, c)
}
