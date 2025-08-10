package swagger

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Config struct {
	Enabled bool   `mapstructure:"enabled"`
	Route   string `mapstructure:"route"`
}

type SwaggerModule struct {
	route string
}

func New(route string) *SwaggerModule {
	return &SwaggerModule{route: route}
}

func (s *SwaggerModule) Register(r *gin.Engine) {
	r.GET(s.route, ginSwagger.WrapHandler(swaggerFiles.Handler))
}
