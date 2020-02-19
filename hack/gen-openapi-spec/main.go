package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/emicklei/go-restful"
	"github.com/go-openapi/spec"
	"github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"

	"k8s.io/kube-openapi/pkg/builder"
	"k8s.io/kube-openapi/pkg/common"
	"k8s.io/kube-openapi/pkg/util"
)

const defaultSwaggerFile = "swagger.json"

// Generate OpenAPI spec definitions for Network Attachment Definition Resource
func main() {
	if len(os.Args) <= 1 {
		log.Fatal("Excpected arguments: <version> [swagger-file-name]")
	}

	version := os.Args[1]

	var swaggerFilename string
	if len(os.Args) > 2 {
		swaggerFilename = os.Args[2]
	} else {
		swaggerFilename = defaultSwaggerFile
	}

	var defNames []string
	for name := range v1.GetOpenAPIDefinitions(func(name string) spec.Ref {
		return spec.Ref{}
	}) {
		defNames = append(defNames, name)
	}

	config := createOpenAPIBuilderConfig(version)
	config.GetDefinitions = v1.GetOpenAPIDefinitions
	swagger, err := builder.BuildOpenAPISpec(createWebServices(), config)
	if err != nil {
		log.Fatalf("Failed to create open-api spec: %s", err.Error())
	}

	specBytes, err := json.MarshalIndent(swagger, "", "  ")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = ioutil.WriteFile(swaggerFilename, specBytes, 0644)
	if err != nil {
		log.Fatalf("stdout write error: %s", err.Error())
	}
	log.Printf("spec file generated: %s", swaggerFilename)
}

func createOpenAPIBuilderConfig(version string) *common.Config {
	return &common.Config{
		ProtocolList:   []string{"https"},
		IgnorePrefixes: []string{"/swaggerapi"},
		Info: &spec.Info{
			InfoProps: spec.InfoProps{
				Title:   "k8snetworkplumbingwg",
				Version: version,
			},
		},
		ResponseDefinitions: map[string]spec.Response{
			"NotFound": spec.Response{
				ResponseProps: spec.ResponseProps{
					Description: "Entity not found.",
				},
			},
		},
		CommonResponses: map[int]spec.Response{
			404: *spec.ResponseRef("#/responses/NotFound"),
		},
	}
}

func createWebServices() []*restful.WebService {
	w := new(restful.WebService)
	w.Route(buildRouteForType(w, "v1", "NetworkAttachmentDefinition"))
	w.Route(buildRouteForType(w, "v1", "NetworkAttachmentDefinitionList"))
	return []*restful.WebService{w}
}

// Implements OpenAPICanonicalTypeNamer
var _ = util.OpenAPICanonicalTypeNamer(&typeNamer{})

type typeNamer struct {
	pkg  string
	name string
}

func (t *typeNamer) OpenAPICanonicalTypeName() string {
	return fmt.Sprintf("github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/%s.%s", t.pkg, t.name)
}

func buildRouteForType(ws *restful.WebService, pkg, name string) *restful.RouteBuilder {
	namer := typeNamer{
		pkg:  pkg,
		name: name,
	}
	return ws.GET(fmt.Sprintf("apis/k8s.cni.cncf.io/%s/%s", pkg, strings.ToLower(name))).
		To(func(*restful.Request, *restful.Response) {}).
		Writes(&namer)
}

